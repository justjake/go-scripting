// +build ignore

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/justjake/go-scripting/annotation2"
	"golang.org/x/tools/imports"
)

const (
	marker = `+StaticCompose`
)

var (
	in      = flag.String("in", ".", "Directory of go files to process")
	outPath = flag.String("out", "static_compose_generated.go", "Output file")
)

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

func (c *staticCompose) NewName() string {
	return fmt.Sprintf(c.OuterAppend(), c.InnerName())
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
// {{.NewName}} is equivalent to {{.InnerRecv}}{{.InnerName}}({{.OuterName}}({{.OuterArgsList}}))
func {{.InnerRecvDecl}} {{.NewName}}{{.OuterArgsDecl}} {{.InnerReturnDecl}} {
	return {{.InnerRecv}}{{.InnerName}}({{.OuterName}}({{.OuterArgsList}}))
}
`

var tmpl = template.Must(template.New("composed function").Parse(composed))

func main() {
	flag.Parse()
	loader := annotation2.NewLoader()
	loader.IncludeDir(*in, nil)
	pipeline := annotation2.DefaultPipeline(loader)
	pipeline.AddStep("StaticCompose: Find", find)
	pipeline.AddStep("StaticCompose: Generate", generateAndWrite)
	if err := pipeline.Run(); err != nil {
		panic(err)
	}
}

func find(unit annotation2.UnitAPI) (interface{}, error) {
	directives := []*directive{}
	db := unit.Input().(annotation2.AnnotationAPI)
	for _, hit := range db.All() {
		directive := new(directive)
		// add directive
		switch hit.Name() {
		case "StaticCompose.Inside":
			directive.Inside = hit.Args()[0].(string)
		case "StaticCompose.Group":
			directive.Group = hit.Args()[0].(string)
			directive.Append = hit.Args()[1].(string)
		default:
			unit.Errorf(hit.Pos(), "unknown annotation name: %#v", hit.Name())
			continue
		}
		directive.FuncDecl = hit.From().(*ast.FuncDecl)
		directives = append(directives, directive)
	}
	return directives, nil
}

func generateAndWrite(unit annotation2.UnitAPI) (interface{}, error) {
	directives := unit.Input().([]*directive)
	out, err := generate(unit.Package().Pkg.Name(), directives, unit.Package().Fset)
	if err != nil {
		return out, err
	}
	err = ioutil.WriteFile(*outPath, out, 0644)
	if err != nil {
		return out, err
	}
	return nil, nil
}

func generate(pkg string, directives []*directive, fset *token.FileSet) ([]byte, error) {
	var out bytes.Buffer
	fmt.Fprintf(&out, "package %s\n", pkg)
	fmt.Fprintln(&out, "")
	fmt.Fprintf(&out, "// AUTO-GENERATED WITH %s %v", filepath.Base(os.Args[0]), os.Args[1:])
	for _, d := range directives {
		if d.Inside == "" {
			continue
		}
		//log.Printf("Put inside %q: %s", d.Inside, funcString(fset, d.FuncDecl))
		// n^2 here we come.
		for _, g := range directives {
			if g.Group != d.Inside {
				continue
			}
			fmt.Printf("Put inside %q: outer: %s, inner: %s\n", d.Inside, funcString(fset, g.FuncDecl), funcString(fset, d.FuncDecl))
			c := &staticCompose{fset, d, g}
			//log.Println(c.Render())
			out.WriteRune('\n')
			out.WriteString(c.Render())
		}
	}
	outImports, err := imports.Process("generated_file.go", out.Bytes(), nil)
	if err != nil {
		return outImports, err
	}
	return format.Source(outImports)
}
