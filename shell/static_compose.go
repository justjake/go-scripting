package main

//+build ignore

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
)

const (
	marker = `+StaticCompose `
)

var (
	in    = flag.String("in", ".", "Directory of go files to process")
	out   = flag.String("out", "static_compose_generated.go", "Output file")
	state = state{}
)

type visitor struct {
	groups     map[string]group
	directives []*directive
	fset       *token.FileSet
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	return v
}

// Scan the file for comments which may contain directives, and process those.
func (v *visitor) VisitFile(file *ast.File) {
	cmap := ast.NewCommentMap(v, fset, file, file.Comments)
	for _, node := range file.Decls {
		if fdecl, ok := node.(*ast.FuncDecl); ok {
			v.VisitFuncDecl(fdecl, cmap)
		}
	}
}

func (v *visitor) VisitFuncDecl(decl *ast.FuncDecl, cmap ast.CommentMap) {
	if decl.Doc == nil {
		return
	}

	directive := v.ParseCommentGroup(decl.Doc)
	if directive == nil {
		return
	}
	directive.FuncDecl = decl
	// TODO
}

// attempt to parse directives from the given comment.
func (v *visitor) ParseCommentGroup(cg *ast.CommentGroup) *directive {
	// skip comments unless they are associated with a function declaration
	for _, c := range cg.List {
		trimmed := strings.TrimSpace(c.Text)
		if i := strings.Index(trimmed, marker); i != 0 {
			// comment does not start with +StaticCompose
			continue
		}
		rest := strings.TrimPrefix(s, marker+` `)
		log.Println("Found directive: ", trimmed)
		tags := reflect.StructTag(rest)
		d := &directive{
			Inside: tags.Get("inside"),
			Group:  tags.Get("group"),
			Append: tags.Get("append"),
		}
		log.Printf("Parsed directive: %#v", d)
		return d
	}
	return nil
}

type group struct {
	Name string
	Fns  []string // TODO
}

type directive struct {
	Inside
	Group
	Append string
}

func main() {
	flag.Parse()
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, *in, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	v := &visitor{
		groups:     make(map[string]group),
		directives: make([]directive),
		fset:       fset,
	}
	for _, p := range packages {
		for _, f := range p.Files {
			v.VisitFile(f)
		}
	}
}
