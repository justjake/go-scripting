package main

// Write this file w/ :noautocmd w<Enter>
// which disables goimports mangling

/*
Foo
@OnLoneImport()
Bar
*/
import "fmt"

import (
	// @OnNamedImport()
	strlib "strings"

	// @OnDotImport()
	. "bytes"

	// @OnNormalImport()
	"io"
)

// @OnType()
type Thing struct {
	Name string
	// @OnField()
	Age int
}

// @OnFunc()
func (t *Thing) Greeting() string {
	return fmt.Sprintf("Hello, %s, you're %d", t.Name, t.Age)
}

// @OnLoneVar()
var SomeVar = 5

const (
	// @OnGroupedConst()
	none = iota
	// @OnGroupedConstNoValue()
	some
	quantum
)

// @OnDoubleVar()
var two, together = 3, 4

func somePriv() int {
	return 5
}

// string, int, float
// @Literals("a string", 5, -0.125)
//
// type, method of type, field of Type, func
// @LocalRefs(Thing, Thing.Greeting, Thing.Name, somePriv)
//
// package, type of package, method of type of package, func of package
// @RemoteRefs(fmt, fmt.Stringer, fmt.Stringer.String, fmt.Sprintf)
type Magnitude int

// Mistakes
// @NotACall.Foo.Bar + 1
// @BadCallSyntax(foo bar)
// @BadCallMath(1 + 1)
// @BadCallFn(-555, Foo.Bar())
type Foo int

func main() {
	fmt.Println((&Thing{"Bob", 99}).Greeting())
	fmt.Println(strlib.Join([]string{"foo", "bar"}, "\n"))
	var buf Buffer
	buf.WriteString("foo")
	// lol
	io.Copy(&buf, &buf)
}
