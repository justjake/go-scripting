package cli

import (
	"fmt"
	"go/ast"

	"github.com/justjake/go-scripting/annotation"
)

var ui = &UI{
	Commands: []Command{},
	Args:     []Arg{},
}

// @AnnotationHandler()
func CLI(hit *annotation.Hit) error {
	// create a command for each of the hit type's public
	// methods
	for _, meth := range hit.Methods() {
		// After writing this and thinking about the APIs we need, it's clear that:
		// 1. annotation.Hit and annotation.Ref must both reference all the AST stuff,
		//    and all the type stuff, or, such stuff must be available in another object
		//    with query methods that take Ref or Hit
		// 2. After investigation, it appears that the go "tools" repo has a ton of utilities
		//    that will simplify our implementation, especially things that supercede custom
		//    implementations in annotation.
		//
		//    We could use this to structure our traversal of source:
		//    https://godoc.org/golang.org/x/tools/go/analysis
		//
		//    Coming soon is a new system for AST traversals:
		//    https://go-review.googlesource.com/c/tools/+/135655
		//
		//    In general golang.org/x/tools/go/types/typeutil looks
		//    great, especially https://godoc.org/golang.org/x/tools/go/types/typeutil#Map
		//    which can help store derived state for types for later
		//    program synthesis, although the Lemma system in anaylsis will
		//    be even more convinient.
		//    Also cool: https://godoc.org/golang.org/x/tools/go/types/typeutil#MethodSetCache
		//
		//    In the astutil package, there's this:
		//    https://godoc.org/golang.org/x/tools/go/ast/astutil#PathEnclosingInterval
		//    which maps from a Pos to a Node
		name := meth.Name()
		decl := meth.Node().(*ast.FuncDecl)
		desc, err := parseShortLong(name, decl.CommendGroup.Text())
		if err != nil {
			return err
		}
		if isCommandName(name) {
			ui.cmd(name, func(c *Command) {
				c.Name = name
				c.Original = name
			})
			continue
		}
		if isArgName(name) {
			ui.arg(name, func(a *Arg) {
				a.Name = name
				a.Original = name
			})
			continue
		}
		fmt.Println("ignored method %v", meth)
	}
}

// @AnnotationHandler()
func Optional(hit *annotation.Hit, args ...string) {
	ui.cmd(hit.Name(), func(c *Command) {
		c.Optional = append(c.Optional, args...)
	})
}

// @AnnotationHandler()
func Required(hit *annotation.Hit, args ...string) {
	ui.cmd(hit.Name(), func(c *Command) {
		c.Required = append(c.Required, args...)
	})
}
