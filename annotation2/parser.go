package annotation2

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

// Important public interfaces:

// Ref represents a reference to some type from an AST node in a
// comment or annotation.
type Ref interface {
	// Position information
	ast.Node
	// Syntax of the reference. Do not trust the position
	// information of this node or its children.
	Syntax() ast.Node
	// The AST node in the package that this ref or annotaiton is
	// attatched to.
	From() ast.Node
}

// Annotation in a comment, attatched to a Go syntax element.
type Annotation interface {
	Ref
	// Annotations are always function call expressions.
	CallExpr() *ast.CallExpr
	// The evaluated arguments of the annotation. Basic literals are
	// evaluated to their go types, and type references are returned
	// as Refs.
	Args() []interface{}
}

// A node that was moved from its initial parse location
// information to somewhere else. Children of this node
// contain bogus position information
type moved struct {
	ast.Node
	start token.Pos
}

func (n *moved) Pos() token.Pos {
	return n.start
}

func (n *moved) End() token.Pos {
	delta := n.Node.End() - n.Node.Pos()
	return n.start + delta
}

func (n *moved) Syntax() ast.Node {
	return n.Node
}

type ref struct {
	moved
	from ast.Node
}

func (r *ref) From() ast.Node {
	return r.from
}

type annotation struct {
	ref
	// gotta mess with this
	args []interface{}
}

func (an *annotation) CallExpr() *ast.CallExpr {
	return an.moved.Node.(*ast.CallExpr)
}

func (an *annotation) Args() []interface{} {
	return an.args
}

// Parser parses annotations in a package.
type Parser struct {
	// Parser will call Errorf once for every error encountered.
	Errorf func(token.Pos, string, ...interface{}) error
}

func (p *Parser) Parse(root ast.Node) []Annotation {
	hits := []Annotation{}
	ast.Inspect(root, func(nodeIface ast.Node) bool {
		if nodeIface == nil {
			return false
		}

		switch node := nodeIface.(type) {
		case *ast.Field:
			// TODO: is this correct, or should this be handled within gendecl?
			hits = append(hits, p.ParseCommentGroup(node.Doc, node)...)
		case *ast.GenDecl:
			hits = append(hits, p.ParseCommentGroup(node.Doc, node)...)
		case *ast.FuncDecl:
			hits = append(hits, p.ParseCommentGroup(node.Doc, node)...)
		}
		return true
	})
	return hits
}

func (p *Parser) onField(decl *ast.Field, out *[]Annotation) {
	hits := p.ParseCommentGroup(decl.Doc, decl)
	*out = append(*out, hits...)
}

func (p *Parser) onGenDecl(decl *ast.GenDecl, out *[]Annotation) {
	// represents an import, constant, type or variable declaration
	// https://devdocs.io/go/go/ast/index#GenDecl
	hits := p.ParseCommentGroup(decl.Doc, decl)
	*out = append(*out, hits...)
}

func (p *Parser) onFuncDecl(decl *ast.FuncDecl, out *[]Annotation) {
	hits := p.ParseCommentGroup(decl.Doc, decl)
	*out = append(*out, hits...)
}

var annotationBeginSingle = regexp.MustCompile(`^// ?@`)
var annotationBeginMulti = regexp.MustCompile(`(?m)^@`)

// ParseComment parses the annotations in a single comment.
func (p *Parser) ParseComment(comment *ast.Comment, from ast.Node) []Annotation {
	rg := annotationBeginSingle
	if strings.HasPrefix(comment.Text, "/*") {
		rg = annotationBeginMulti
	}

	ms := rg.FindAllStringIndex(comment.Text, -1)
	if ms == nil {
		return nil
	}

	hits := []Annotation{}

	for _, m := range ms {
		offset := m[1]
		atStart := comment.Text[offset:]
		end := strings.IndexRune(atStart, '\n')
		if end == -1 {
			end = len(atStart)
		}
		startPos := comment.Pos() + token.Pos(offset)
		chunk := atStart[:end]
		// ignore error since it's bubbled up as part of the whole UnitAPI shtick.
		hit, _ := p.parseAnnotationAt(startPos, chunk, from)
		if hit != nil {
			hits = append(hits, hit)
		}
	}
	return hits
}

func (p *Parser) parseAnnotationAt(startPos token.Pos, chunk string, from ast.Node) (*annotation, error) {
	makeErr := func(pos token.Pos, msg interface{}) error {
		return p.Errorf(pos, "%s: %v", chunk, msg)
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
			ref := &ref{
				moved{arg, startPos + arg.Pos()},
				from,
			}
			args[j] = ref
		case *ast.SelectorExpr:
			if err := identOnlySelector(arg); err != nil {
				return nil, makeErr(arg.Pos(), err)
			}
			ref := &ref{
				moved{arg, startPos + arg.Pos()},
				from,
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
	return &annotation{
		ref{
			moved{call, startPos + call.Pos()},
			from,
		},
		args,
	}, nil
}

// ParseAnnotations parses the given text, returning applied annotation hits
// attatched to the given node. If errors are encountered, returns nil hits,
// and the errors.
func (p *Parser) ParseCommentGroup(cg *ast.CommentGroup, from ast.Node) []Annotation {
	if cg == nil || len(cg.List) == 0 {
		return nil
	}
	hits := []Annotation{}

	for _, comment := range cg.List {
		hit := p.ParseComment(comment, from)
		hits = append(hits, hit...)
	}

	return hits
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
