package cli

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

type file struct {
	filename string
	src      interface{}
}

type multierror []error

func (e multierror) Error() string {
	var out bytes.Buffer
	out.WriteRune('\n')
	for _, err := range e {
		fmt.Fprintln(&out, err)
	}
	return out.String()
}

// mostly cribbed from https://github.com/golang/tools/blob/master/cmd/gotype/gotype.go
func buildErrors(files []file) error {
	fset := token.NewFileSet()
	parsed := make([]*ast.File, len(files))
	parserMode := parser.AllErrors
	// parse files
	for i, file := range files {
		ast, err := parser.ParseFile(fset, file.filename, file.src, parserMode)
		if err != nil {
			return err
		}
		parsed[i] = ast
	}

	errors := []error{}

	// check types
	conf := types.Config{
		// disable C go checking - we don't use it
		FakeImportC: true,
		Error: func(err error) {
			errors = append(errors, err)
		},
		Importer: importer.Default(),
		Sizes:    types.SizesFor(build.Default.Compiler, build.Default.GOARCH),
	}
	conf.Check("pkg", fset, parsed, nil)

	if len(errors) > 0 {
		return multierror(errors)
	}
	return nil
}

func loadPackageString(importPath, text string) (*token.FileSet, *doc.Package) {
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

	return fset, doc.New(pkg, importPath, 0)
}

func TestParse(t *testing.T) {
	text := `
package main

import (
	"os"
	"fmt"
)

type Fooer struct {
	Name string
}

// NAME is the user's full name.
func (f *Fooer) NAME() {
	return os.Getenv("NAME")
}

// Greet shows a greeting to NAME.
// Optional: NAME
func (f *Fooer) Greet() {
	fmt.Println(NAME())
}

// Dog.
// Tags: foo, bar
func (f *Fooer) LAST() {
	return os.Getenv("LAST")
}

// Show shows all the things. Use show when you need an extensive greeting.
func (f *Fooer) Show() {
	fmt.Println(NAME())
}

// Yolo
//func Dog() {
//	fmt.Println("hi")
//}

func main() {
	Greet()
}
	`

	expected := &UI{
		Description: Description{},
		Commands: []Command{
			{
				Description: Description{
					Name:     "greet",
					Short:    "Shows a greeting to NAME.",
					Long:     "",
					Tags:     nil,
					Original: "Greet",
				},
				Optional: []string{"NAME"},
				Required: nil,
			},
			{
				Description: Description{
					Name:     "show",
					Short:    "Shows all the things.",
					Long:     "Use show when you need an extensive greeting.",
					Tags:     nil,
					Original: "Show",
				},
				Optional: nil,
				Required: nil,
			},
		},
		Args: []Arg{
			{
				Description: Description{
					Name:     "LAST",
					Short:    "Dog.",
					Long:     "",
					Tags:     []string{"foo", "bar"},
					Original: "LAST",
				},
			},
			{
				Description: Description{
					Name:     "NAME",
					Short:    "The user's full name.",
					Long:     "",
					Tags:     nil,
					Original: "NAME",
				},
			},
		},
	}

	fset, pkg := loadPackageString("github.com/justjake/examples", text)
	ui, err := Parse(fset, pkg, "*Fooer")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	assert.Equal(t, expected, ui)

	// TODO: figure out how to test that generated code compiles
	asFile := ToFileContents(ui, "*Fooer")
	assert.Equal(t, "", asFile)
	//err = buildErrors([]file{
	//{"main.go", text},
	//{"generated.go", asFile},
	//})
	//assert.Empty(t, err)
	//assert.Equal(t, expectedOut, asFile)
}
