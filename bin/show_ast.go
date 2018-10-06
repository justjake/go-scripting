package main

import (
	"go/ast"
	"go/parser"
	"go/token"
)

const src = `
package dumb

// comment on lone import
import "wat"

// comment on imports
import (
	// comment on foo
	foo "foo"
	// comment on bar
	. "bar"
	// comment on baz
	"baz" // comment after baz
)

const (
	// @OnFoo()
	foo = iota
	// @OnBar()
	bar
	// @OnBaz()
	baz
)
`

func main() {
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Print the AST.
	ast.Print(fset, f)
}
