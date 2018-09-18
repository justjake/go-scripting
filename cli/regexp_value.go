package cli

import (
	"flag"
	"regexp"
)

type regexpValue regexp.Regexp

func newRegexpValue(val *regexp.Regexp) *regexpValue {
	return (*regexpValue)(val)
}

func (r *regexpValue) Set(s string) error {
	v, err := regexp.Compile(s)
	r = (*regexpValue)(v)
	return err
}

func (r *regexpValue) Get() interface{} { return (*regexp.Regexp)(r) }

func (r *regexpValue) String() string { return (*regexp.Regexp)(r).String() }

// @FlagValue(*regexp.Regexp)
func Regexp() flag.Getter {
	r := regexp.MustCompile("")
	return newRegexpValue(r)
}
