/*
Package annotation implements a system for generating go code based on
annotations in comments.

Annotations are of the form "@SomeName(arg1, arg2, arg3)", where
SomeName(arg1, arg2, arg3) is valid go syntax for a function call with
literal arguments. As a special case, arguments can also be type names, or
fields of a type.

Currently, annotations support only basic literals as arguments: strings and
numbers, and negative numbers.
*/
package annotation

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"strconv"
	"strings"
)

// Func is the type of annotation functions. All annotation functions take as
// their first argument the ast.Node to which the annotation is attatched. The
// remaining interface arguments are the user-supplied literals in the
// annotation comment, or are ast.Nodes if the user supplied a type name.
//
// Any error returned will be wrapped in additional information about the
// source location and node name.
type Func func(ast.Node, ...interface{}) error

// Hit describes a successful application of an annotation
type Hit struct {
	// C
	*ast.CallExpr
	// Node the annotation is attatched to
	From ast.Node
	// Evaluated arguments
	Args []interface{}
}

// FuncName returns the name of the annotation function
func (hit *Hit) FuncName() string {
	return toStr(hit.CallExpr.Fun)
}

func (hit *Hit) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "Hit{")
	fmt.Fprintf(&buf, "%q", toStr(hit.CallExpr.Fun))
	if len(hit.Args) > 0 {
		fmt.Fprint(&buf, " with")
		for _, arg := range hit.Args {
			fmt.Fprintf(&buf, " %#v", arg)
		}
	}
	fmt.Fprint(&buf, "}")
	return buf.String()
}

// RefKind describes the kind of reference to a code entity.
type RefKind string

const (
	// InvalidKind is some unknown kind of reference
	InvalidKind = RefKind("Invalid")
	// LocalFunc references is a function
	LocalFunc = RefKind("Func")
	// LocalType is a type
	LocalType = RefKind("Type")
	// LocalTypeField is a field or func in a type
	LocalTypeField = RefKind("Type.Field")
	// Pkg is an imported package
	Pkg = RefKind("pkg")
	// PkgFunc is a function in an importated package
	PkgFunc = RefKind("pkg.Func")
	// PkgType is a type in an imported package
	PkgType = RefKind("pkg.Type")
	// PkgTypeField is a field of a type in an imported package
	PkgTypeField = RefKind("pkg.Type.Field")
)

// Ref represents a reference to a type, a method of a type, a variable, or a
// constant in an annotation call.
type Ref struct {
	// The reference node itself, parsed from an annotation comment. It's type is
	// either an *ast.Ident or an *ast.SelectorExpr.
	ast.Node
	// The node the annotation is attatched to.
	From ast.Node
}

// Selector returns the referenced path as a dot-seperated string
func (r *Ref) Selector() string {
	return toStr(r.Node)
}

func (r *Ref) String() string {
	return r.GoString()
}

func (r *Ref) GoString() string {
	return fmt.Sprintf("Ref{%q}", r.Selector())
}

// Processor parses the comments of a Go AST for annotation comments and calls configured
// annotation functions.
//
// TODO: split into Parse and Eval stages
type Processor struct {
	// Map of annotation names to Funcs that process those annotations.
	Funcs map[string]Func
	// Filled with successful annotation hits. For stats.
	Hits []*Hit
	// Filled with unsuccessful annotation hits. For stats.
	Errors []error
}

// NewProcessor returns a new Processor with initialized fields
func NewProcessor() *Processor {
	return &Processor{
		Funcs:  make(map[string]Func),
		Hits:   []*Hit{},
		Errors: []error{},
	}
}

// Visit implements ast.Visitor for Processor.
func (p *Processor) Visit(nodeIface ast.Node) ast.Visitor {
	switch node := nodeIface.(type) {
	case *ast.Field:
		// TODO: is this correct, or should this be handled within gendecl?
		p.onField(node)
	case *ast.GenDecl:
		p.onGenDecl(node)
	case *ast.FuncDecl:
		p.onFuncDecl(node)
	}
	return p
}

func (p *Processor) onField(decl *ast.Field) {
	text := decl.Doc.Text()
	hits, errs := p.ParseAnnotations(text, decl)
	p.Errors = append(p.Errors, errs...)
	p.Hits = append(p.Hits, hits...)
}

func (p *Processor) onGenDecl(decl *ast.GenDecl) {
	// represents an import, constant, type or variable declaration
	// https://devdocs.io/go/go/ast/index#GenDecl
	text := decl.Doc.Text()
	hits, errs := p.ParseAnnotations(text, decl)
	p.Errors = append(p.Errors, errs...)
	p.Hits = append(p.Hits, hits...)
}

func (p *Processor) onFuncDecl(decl *ast.FuncDecl) {
	text := decl.Doc.Text()
	hits, errs := p.ParseAnnotations(text, decl)
	p.Errors = append(p.Errors, errs...)
	p.Hits = append(p.Hits, hits...)
}

type parseError struct {
	Line string
	Num  int
	error
}

func (pe *parseError) Error() string {
	return fmt.Sprintf("bad annotation %q: %v", pe.Line, pe.error)
}

// ParseAnnotations parses the given text, returning applied annotation hits
// attatched to the given node. If errors are encountered, returns nil hits,
// and the errors.
func (p *Processor) ParseAnnotations(text string, node ast.Node) ([]*Hit, []error) {
	// base case: no doc
	if text == "" {
		return nil, nil
	}

	errs := []error{}
	hits := []*Hit{}

	// line-by-line
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if len(l) > 0 && l[0] == '@' {
			// must be an expression
			expr, err := parser.ParseExpr(l[1:])
			if err != nil {
				errs = append(errs, &parseError{l, i, err})
				continue
			}
			// must be a function call expression
			call, ok := expr.(*ast.CallExpr)
			if !ok {
				errs = append(errs, &parseError{l, i, fmt.Errorf("not a func call, instead %T", expr)})
				continue
			}

			// TODO: move this to evaluate stage
			// evaluate arguments. Literals to literals, refs to Ref
			args := make([]interface{}, len(call.Args))
			argsHasErrors := false
			for j, unknownArg := range call.Args {
				switch arg := unknownArg.(type) {
				case *ast.Ident:
					ref := &Ref{
						Node: arg,
						From: node,
					}
					// TODO: check ref now?
					args[j] = ref
				case *ast.SelectorExpr:
					ref := &Ref{
						Node: arg,
						From: node,
					}
					// TODO: check ref now?
					args[j] = ref
				case *ast.BasicLit:
					val, err := evalLit(arg)
					if err != nil {
						errs = append(errs, &parseError{l, i, fmt.Errorf("arg %d: %v", j, err)})
						argsHasErrors = true
						continue
					}
					args[j] = val
				case *ast.UnaryExpr:
					val, err := evalLit(arg)
					if err != nil {
						errs = append(errs, &parseError{l, i, fmt.Errorf("arg %d: %v", j, err)})
						argsHasErrors = true
						continue
					}
					args[j] = val
				default:
					errs = append(errs, &parseError{l, i, fmt.Errorf("arg %d: unsupported syntax %q", j, toStr(unknownArg))})
					argsHasErrors = true
				}
			}
			if argsHasErrors {
				continue
			}
			hits = append(hits, &Hit{call, node, args})
		}
	}

	return hits, errs
}

var emptyFset = token.NewFileSet()

func toStr(node ast.Node) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, emptyFset, node)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

// Evals the given node, which must be a BasicLit or a UnaryExpr of a BasicLit
func evalLit(node ast.Node) (interface{}, error) {
	str := toStr(node)
	var lit *ast.BasicLit
	if unary, ok := node.(*ast.UnaryExpr); ok {
		lit, ok = unary.X.(*ast.BasicLit)
		if !ok {
			return nil, fmt.Errorf("Not a basic literal: %v", unary.X)
		}
	}
	if thelit, ok := node.(*ast.BasicLit); ok {
		lit = thelit
	}
	if lit == nil {
		return nil, fmt.Errorf("Not a basic literal or unary expr: %v", node)
	}
	switch lit.Kind {
	case token.STRING:
		return strconv.Unquote(str)
	case token.INT:
		return strconv.Atoi(str)
	case token.FLOAT:
		return strconv.ParseFloat(str, 64)
	default:
		return nil, fmt.Errorf("Literal type %v not handled: %v", lit.Kind, str)
	}
}
