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
	"go/scanner"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

// Hit describes a successful application of an annotation
type Hit struct {
	// AST of the annotation. Location information here is garbage
	*ast.CallExpr
	// Node the annotation is attatched to
	From ast.Node
	// Evaluated arguments
	Args []interface{}
	// Location
	start token.Pos
	end   token.Pos
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
	// Location
	start token.Pos
	end   token.Pos
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

// Parser parses the comments of a Go AST for annotation comments and calls
// configured annotation functions.
type Parser struct {
	// Filled with successful annotation hits.
	Hits []*Hit
	// Filled with unsuccessful annotation hits.
	Errors []error
	fset   *token.FileSet
}

// NewParser returns a new Parser with initialized fields
func NewParser(fset *token.FileSet) *Parser {
	return &Parser{
		Hits:   []*Hit{},
		Errors: []error{},
		fset:   fset,
	}
}

func (p *Parser) Parse(node ast.Node) {
	ast.Walk(p, node)
}

// Visit implements ast.Visitor for Processor.
func (p *Parser) Visit(nodeIface ast.Node) ast.Visitor {
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

func (p *Parser) onField(decl *ast.Field) {
	hits, errs := p.ParseAnnotations(decl.Doc, decl)
	p.Errors = append(p.Errors, errs...)
	p.Hits = append(p.Hits, hits...)
}

func (p *Parser) onGenDecl(decl *ast.GenDecl) {
	// represents an import, constant, type or variable declaration
	// https://devdocs.io/go/go/ast/index#GenDecl
	hits, errs := p.ParseAnnotations(decl.Doc, decl)
	p.Errors = append(p.Errors, errs...)
	p.Hits = append(p.Hits, hits...)
}

func (p *Parser) onFuncDecl(decl *ast.FuncDecl) {
	hits, errs := p.ParseAnnotations(decl.Doc, decl)
	p.Errors = append(p.Errors, errs...)
	p.Hits = append(p.Hits, hits...)
}

// Most errors returned by ParseComment are ParseErrors.
type ParseError struct {
	// Position of error
	Pos token.Position
	// Syntax string error is associated with
	Context string
	// Inner error
	error
}

func (pe *ParseError) Error() string {
	return fmt.Sprintf("%v: %v in %q", pe.Pos, pe.error, pe.Context)
}

var annotationBeginSingle = regexp.MustCompile(`^// ?@`)
var annotationBeginMulti = regexp.MustCompile(`(?m)^@`)

// ParseComment parses the annotations in a single comment.
func ParseComment(fset *token.FileSet, comment *ast.Comment, from ast.Node) ([]*Hit, []error) {
	rg := annotationBeginSingle
	if strings.HasPrefix(comment.Text, "/*") {
		rg = annotationBeginMulti
	}

	ms := rg.FindAllStringIndex(comment.Text, -1)
	if ms == nil {
		return nil, nil
	}

	hits := []*Hit{}
	errs := []error{}

	for _, m := range ms {
		offset := m[1]
		atStart := comment.Text[offset:]
		end := strings.IndexRune(atStart, '\n')
		if end == -1 {
			end = len(atStart)
		}
		startPos := comment.Pos() + token.Pos(offset)
		chunk := atStart[:end]
		hit, err := parseAnnotationAt(fset, startPos, chunk, from)
		if err != nil {
			errs = append(errs, err)
		} else {
			hits = append(hits, hit)
		}
	}
	return hits, errs

}

func parseAnnotationAt(fset *token.FileSet, startPos token.Pos, chunk string, from ast.Node) (*Hit, error) {
	makeErr := func(pos token.Pos, msg interface{}) error {
		posi := fset.Position(startPos + pos)
		return &ParseError{posi, chunk, fmt.Errorf("%v", msg)}
	}

	// must be an expression
	expr, err := parser.ParseExpr(chunk)
	if err != nil {
		switch err2 := err.(type) {
		case *scanner.Error:
			// rewrite scanner errors to have the correct position.
			return nil, makeErr(token.Pos(err2.Pos.Column-1), fmt.Errorf(err2.Msg))
		case scanner.ErrorList:
			// Only return the first error, which is good enough.
			return nil, makeErr(token.Pos(err2[0].Pos.Column), fmt.Errorf(err2[0].Msg))
		default:
			return nil, makeErr(0, err2)
		}
	}

	// must be a function call expression
	call, ok := expr.(*ast.CallExpr)
	if !ok {
		return nil, makeErr(expr.Pos(), fmt.Errorf("not a func call, instead %T", expr))
	}

	// evaluate arguments. Literals to literals, refs to Ref
	args := make([]interface{}, len(call.Args))
	for j, unknownArg := range call.Args {
		switch arg := unknownArg.(type) {
		case *ast.Ident:
			if err := identOnlySelector(arg); err != nil {
				return nil, makeErr(arg.Pos(), err)
			}
			ref := &Ref{
				Node:  arg,
				From:  from,
				start: startPos + arg.Pos(),
				end:   startPos + arg.End(),
			}
			args[j] = ref
		case *ast.SelectorExpr:
			if err := identOnlySelector(arg); err != nil {
				return nil, makeErr(arg.Pos(), err)
			}
			ref := &Ref{
				Node:  arg,
				From:  from,
				start: startPos + arg.Pos(),
				end:   startPos + arg.End(),
			}
			args[j] = ref
		case *ast.BasicLit:
			val, err := evalLit(arg)
			if err != nil {
				return nil, makeErr(arg.Pos(), err)
			}
			args[j] = val
		case *ast.UnaryExpr:
			val, err := evalLit(arg)
			if err != nil {
				return nil, makeErr(arg.Pos(), err)
			}
			args[j] = val
		default:
			return nil, makeErr(unknownArg.Pos(), fmt.Errorf("unsupported syntax %q", toStr(unknownArg)))
		}
	}

	// tada!
	return &Hit{
		CallExpr: call,
		From:     from,
		Args:     args,
		start:    startPos + call.Pos(),
		end:      startPos + call.End(),
	}, nil
}

// ParseAnnotations parses the given text, returning applied annotation hits
// attatched to the given node. If errors are encountered, returns nil hits,
// and the errors.
//
// TODO: re-work to parse directly from Comment nodes so we can track position exactly
// for Hit, and also make Hit an ast.Node.
func (p *Parser) ParseAnnotations(cg *ast.CommentGroup, node ast.Node) ([]*Hit, []error) {
	if cg == nil || len(cg.List) == 0 {
		return nil, nil
	}
	errs := []error{}
	hits := []*Hit{}

	for _, comment := range cg.List {
		hit, err := ParseComment(p.fset, comment, node)
		hits = append(hits, hit...)
		errs = append(errs, err...)
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

// verify a selectorexpr contains only selectorexpr and ident nodes
func identOnlySelector(sel ast.Node) error {
	var err error
	ast.Inspect(sel, func(node ast.Node) bool {
		if node == nil {
			return false
		}
		switch v := node.(type) {
		case *ast.SelectorExpr:
			return true
		case *ast.Ident:
			return true
		default:
			err = fmt.Errorf("unsupported syntax %T in ref %q", toStr(sel), v)
			return false
		}
	})
	return err
}

// Evals the given node, returning the value that it declars. The node must be
// a BasicLit or a UnaryExpr of a BasicLit.
func evalLit(node ast.Node) (interface{}, error) {
	str := toStr(node)
	var lit *ast.BasicLit
	if unary, ok := node.(*ast.UnaryExpr); ok {
		lit, ok = unary.X.(*ast.BasicLit)
		if !ok {
			return nil, fmt.Errorf("not a basic literal: %v", unary.X)
		}
		if unary.Op != token.SUB {
			return nil, fmt.Errorf("unsupported unary operator %v in %q", unary.Op, str)
		}
	}
	if thelit, ok := node.(*ast.BasicLit); ok {
		lit = thelit
	}
	if lit == nil {
		return nil, fmt.Errorf("not a basic literal or unary expr: %v", node)
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
