package annotation

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadPackageString(importPath, text string) (*token.FileSet, *ast.Package) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "example.go", text, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	pkg := &ast.Package{
		Name: file.Name.Name,
		Files: map[string]*ast.File{
			"example.go": file,
		},
	}

	return fset, pkg
}

func TestParse(t *testing.T) {
	text := `
package main

// @OnImport()
import "fmt"

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

// @OnVar()
var SomeVar = 5

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
// @NotCorrectSyntax.Foo.Bar + 1
// @MistakeInCall(foo bar)
// @OkayCall("seems legit", 1 + 1, Foo.Bar())
type Foo int
	`
	expectedStrings := []string{
		`Hit{"OnImport"}`,
		`Hit{"OnType"}`,
		`Hit{"OnField"}`,
		`Hit{"OnFunc"}`,
		`Hit{"OnVar"}`,
		`Hit{"Literals" with "a string" 5 -0.125}`,
		`Hit{"LocalRefs" with Ref{"Thing"} Ref{"Thing.Greeting"} Ref{"Thing.Name"} Ref{"somePriv"}}`,
		`Hit{"RemoteRefs" with Ref{"fmt"} Ref{"fmt.Stringer"} Ref{"fmt.Stringer.String"} Ref{"fmt.Sprintf"}}`,
	}

	expectedErrs := []string{
		`bad annotation "@NotCorrectSyntax.Foo.Bar + 1": not a func call, instead *ast.BinaryExpr`,
		`bad annotation "@MistakeInCall(foo bar)": 1:19: missing ',' in argument list`,
		`bad annotation "@OkayCall(\"seems legit\", 1 + 1, Foo.Bar())": arg 1: unsupported syntax "1 + 1"`,
		`bad annotation "@OkayCall(\"seems legit\", 1 + 1, Foo.Bar())": arg 2: unsupported syntax "Foo.Bar()"`,
	}

	_, pkg := loadPackageString("github.com/justjake/foo/bar", text)
	p := NewProcessor()
	ast.Walk(p, pkg.Files["example.go"])

	assert.Len(t, p.Errors, len(expectedErrs), "has expected errors count")
	for i := range p.Errors {
		assert.Equal(t, expectedErrs[i], p.Errors[i].Error())
	}
	assert.Len(t, p.Hits, len(expectedStrings), "has expected hits count")
	for i := range p.Hits {
		assert.Equal(t, expectedStrings[i], p.Hits[i].String())
	}
}
