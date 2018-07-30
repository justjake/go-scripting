package main

// +build ignore

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"golang.org/x/tools/imports"
	"io/ioutil"
	"log"
	"reflect"
	"strings"
	"text/template"
)

const (
	marker = `+StaticCompose`
)

var (
	in      = flag.String("in", ".", "Directory of go files to process")
	outPath = flag.String("out", "static_compose_generated.go", "Output file")
)

type visitor struct {
	directives []*directive
	fset       *token.FileSet
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	return v
}

// Scan the file for comments which may contain directives, and process those.
func (v *visitor) VisitFile(file *ast.File) {
	cmap := ast.NewCommentMap(v.fset, file, file.Comments)
	for _, node := range file.Decls {
		if fdecl, ok := node.(*ast.FuncDecl); ok {
			v.VisitFuncDecl(fdecl, cmap, file)
		}
	}
}

func (v *visitor) VisitFuncDecl(decl *ast.FuncDecl, cmap ast.CommentMap, file *ast.File) {
	if decl.Doc == nil {
		return
	}

	directive := v.ParseCommentGroup(decl.Doc)
	if directive == nil {
		return
	}
	directive.FuncDecl = decl
	directive.File = file
	v.directives = append(v.directives, directive)
}

// attempt to parse directives from the given comment.
func (v *visitor) ParseCommentGroup(cg *ast.CommentGroup) *directive {
	lines := strings.Split(cg.Text(), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i := strings.Index(trimmed, marker); i != 0 {
			// comment does not start with +StaticCompose
			continue
		}
		rest := strings.TrimPrefix(trimmed, marker+` `)
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
	FuncDecl *ast.FuncDecl
	File     *ast.File
	Inside   string
	Group    string
	Append   string
}

func nodeString(fset *token.FileSet, node ast.Node) string {
	var out bytes.Buffer
	//if flist, ok := node.(*ast.FieldList); ok {
	//out.WriteString(flist.)
	//}
	printer.Fprint(&out, fset, node)
	return out.String()
}

func funcString(fset *token.FileSet, decl *ast.FuncDecl) string {
	proto := *decl
	proto.Body = nil
	proto.Doc = nil
	return nodeString(fset, &proto)
}

func funcFieldListString(fset *token.FileSet, fields *ast.FieldList) string {
	out := new(bytes.Buffer)
	out.WriteRune('(')
	for i, field := range fields.List {
		if i > 0 {
			out.WriteString(", ")
		}
		if len(field.Names) > 0 {
			out.WriteString(field.Names[0].String())
			out.WriteRune(' ')
		}
		out.WriteString(nodeString(fset, field.Type))
	}
	out.WriteRune(')')
	return out.String()
}

/* something like this:

func (sh *Shell) Outf(pattern string, vs ...interface{}) string {
	return sh.Out(ScriptPrintf(pattern, vs...))
}
*/

type staticCompose struct {
	fset  *token.FileSet
	inner *directive
	outer *directive
}

// first line

func (c *staticCompose) InnerRecvDecl() string {
	if c.inner.FuncDecl.Recv == nil {
		return ""
	}
	return funcFieldListString(c.fset, c.inner.FuncDecl.Recv)
}

func (c *staticCompose) InnerName() string {
	return c.inner.FuncDecl.Name.String()
}

func (c *staticCompose) OuterAppend() string {
	return c.outer.Append
}

func (c *staticCompose) OuterArgsDecl() string {
	return funcFieldListString(c.fset, c.outer.FuncDecl.Type.Params)
}

func (c *staticCompose) InnerReturnDecl() string {
	return funcFieldListString(c.fset, c.inner.FuncDecl.Type.Results)
}

func (c *staticCompose) InnerRecv() string {
	if c.inner.FuncDecl.Recv == nil {
		return ""
	}
	return nodeString(c.fset, c.inner.FuncDecl.Recv.List[0].Names[0]) + "."
}

func (c *staticCompose) OuterName() string {
	return c.outer.FuncDecl.Name.String()
}

func (c *staticCompose) OuterArgsList() string {
	out := new(bytes.Buffer)
	fields := c.outer.FuncDecl.Type.Params.List
	for i, f := range fields {
		if i > 0 {
			out.WriteString(", ")
		}
		out.WriteString(f.Names[0].String())
		if _, ok := f.Type.(*ast.Ellipsis); ok {
			out.WriteString("...")
		}
	}
	return out.String()
}

func (c *staticCompose) Render() string {
	out := new(bytes.Buffer)
	err := tmpl.Execute(out, c)
	if err != nil {
		panic(err)
	}
	return out.String()
}

const composed = `
// {{.InnerName}}{{.OuterAppend}} is equivalent to {{.InnerRecv}}{{.InnerName}}({{.OuterName}}({{.OuterArgsList}}))
func {{.InnerRecvDecl}} {{.InnerName}}{{.OuterAppend}}{{.OuterArgsDecl}} {{.InnerReturnDecl}} {
	return {{.InnerRecv}}{{.InnerName}}({{.OuterName}}({{.OuterArgsList}}))
}
`

var tmpl = template.Must(template.New("composed function").Parse(composed))

func main() {
	flag.Parse()
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, *in, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	v := &visitor{
		directives: []*directive{},
		fset:       fset,
	}
	var firstPackage string
	for name, p := range packages {
		if firstPackage == "" || firstPackage == "main" {
			firstPackage = name
		}
		for _, f := range p.Files {
			v.VisitFile(f)
		}
	}
	var out bytes.Buffer
	// this is a hack.
	fmt.Fprintf(&out, "package %s\n", firstPackage)
	for _, d := range v.directives {
		if d.Inside == "" {
			continue
		}
		//proto := *d.FuncDecl
		//proto.Body = nil
		//proto.Doc = nil
		log.Printf("Put inside %q: %s", d.Inside, funcString(fset, d.FuncDecl))
		// n^2 here we come
		for _, g := range v.directives {
			if g.Group != d.Inside {
				continue
			}
			log.Printf("Put inside %q: outer: %s, inner: %s", d.Inside, funcString(fset, g.FuncDecl), funcString(fset, d.FuncDecl))
			c := &staticCompose{fset, d, g}
			//log.Println(c.Render())
			out.WriteRune('\n')
			out.WriteString(c.Render())
		}
	}
	outImports, err := imports.Process("wat", out.Bytes(), nil)
	if err != nil {
		panic(err)
	}
	outFormat, err := format.Source(outImports)
	if err != nil {
		panic(err)
	}
	ioutil.WriteFile(*outPath, outFormat, 0755)
}
