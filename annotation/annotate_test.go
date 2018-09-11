package annotation

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
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

func parseTestFile(filename string) (*token.FileSet, *ast.Package) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	pkg := &ast.Package{
		Name: file.Name.Name,
		Files: map[string]*ast.File{
			filename: file,
		},
	}
	return fset, pkg
}

func TestParse(t *testing.T) {
	successToStr := `Hit{"OnImport"}
Hit{"OnType"}
Hit{"OnField"}
Hit{"OnFunc"}
Hit{"OnVar"}
Hit{"Literals" with "a string" 5 -0.125}
Hit{"LocalRefs" with Ref{"Thing"} Ref{"Thing.Greeting"} Ref{"Thing.Name"} Ref{"somePriv"}}
Hit{"RemoteRefs" with Ref{"fmt"} Ref{"fmt.Stringer"} Ref{"fmt.Stringer.String"} Ref{"fmt.Sprintf"}}`

	errToStr := `bad annotation "@NotCorrectSyntax.Foo.Bar + 1": not a func call, instead *ast.BinaryExpr
bad annotation "@MistakeInCall(foo bar)": 1:19: missing ',' in argument list
bad annotation "@OkayCall(\"seems legit\", 1 + 1, Foo.Bar())": arg 1: unsupported syntax "1 + 1"
bad annotation "@OkayCall(\"seems legit\", 1 + 1, Foo.Bar())": arg 2: unsupported syntax "Foo.Bar()"`

	fset, pkg := parseTestFile("testdata/annotation_types.go")
	p := NewParser(fset)
	p.Parse(pkg)

	allHits := join(len(p.Hits), func(i int) string { return p.Hits[i].String() })
	allErrs := join(len(p.Errors), func(i int) string { return p.Errors[i].Error() })

	assert.Equal(t, successToStr, allHits)
	assert.Equal(t, errToStr, allErrs)
}

func join(l int, f func(i int) string) string {
	res := make([]string, l)
	for i := range res {
		res[i] = f(i)
	}
	return strings.Join(res, "\n")
}
