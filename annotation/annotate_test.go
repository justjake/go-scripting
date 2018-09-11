package annotation

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
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
	`
	_, pkg := loadPackageString("github.com/justjake/foo/bar", text)
	p := NewProcessor()
	ast.Walk(p, pkg.Files["example.go"])
	for _, h := range p.Hits {
		fmt.Println(h)
	}
	t.FailNow() // just to get log lines
}
