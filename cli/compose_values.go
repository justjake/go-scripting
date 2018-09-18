package cli

import (
	"bytes"
	"flag"
	"fmt"
	"reflect"
)

type repeatedFlag struct {
	// Function that returns a new flag Getter who's Get() returns a T.
	maker func() flag.Getter
	// must be a []T, where T is retured by New().Get()
	out interface{}
	// The values created. All of these values have had Set() called on them
	// once, successfully. They exist for String().
	getters []flag.Getter
}

// Repeated is a type of flag that will record any number of instances of a
// command-line option.
//
// Example:
//   regexes := []*regexp.Regexp
//   flag.Var(RepeatedFlag{NewRegexpValue, &regexes}, "exclude", "Exclude any paths matching `regexp`. Pass more than once.")
//   flag.Parse()
func Repeated(maker func() flag.Getter, out interface{}) flag.Getter {
	return &repeatedFlag{maker, out, []flag.Getter{}}
}

// Get is part of flag.Getter interface
func (rf *repeatedFlag) Get() interface{} {
	return rf.out
}

// Set is part of flag.Getter interface
func (rf *repeatedFlag) Set(next string) error {
	newval := rf.maker()
	err := newval.Set(next)
	if err != nil {
		return err
	}

	outval := reflect.ValueOf(rf.out)
	toAppend := reflect.ValueOf(newval.Get())
	if outval.Type() != reflect.PtrTo(reflect.SliceOf(toAppend.Type())) {
		return fmt.Errorf("RepeatedFlag: type mismatch for T %v: out not *[]T, instead %v", toAppend.Type(), outval.Type())
	}
	// This will panic if the type of outval is not *[]T
	// This will panic if the type of toAppend is not T
	slice := reflect.Append(outval.Elem(), toAppend)
	outval.Set(slice)

	// save the getter, too, for String()
	rf.getters = append(rf.getters, newval)
	return nil
}

// String is part of flag.Getter interface
func (rf *repeatedFlag) String() string {
	if rf == nil || len(rf.getters) == 0 {
		return ""
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "(\n")
	for _, v := range rf.getters {
		fmt.Fprintf(&buf, "  %s\n", v.String())
	}
	fmt.Fprintf(&buf, ")")
	return buf.String()
}

type maxFlag struct {
	max int
	flag.Getter
	count int
}

// Max limits the calls to a flag.Value's Set() method. Calls after the limit
// is reached return an error. Use it to limit the count of a repeated flag
// with Repeated().
//
// Example:
//   teams := []string{}
//   val := Max(16, Repeated(NewStringValue, &teams))
//   flag.Var(val, "team", "Add team `name` to the tourney. Max 16 teams.)
func Max(n int, getter flag.Getter) flag.Getter {
	return &maxFlag{n, getter, 0}
}

func (v *maxFlag) Set(s string) error {
	if v.count == v.max {
		return fmt.Errorf("Already have max %d", v.max)
	}
	v.count++
	return v.Getter.Set(s)
}
