package script

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/build"
	"go/doc"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"testing"
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

	//expectedOut := `wat`

	expected := &UI{
		Description: Description{},
		Commands: []Command{
			{
				Description: Description{
					Name:  "Show",
					Short: "Shows all the things.",
					Long:  "Use show when you need an extensive greeting.",
					Tags:  nil,
				},
				Optional: nil,
				Required: nil,
			},
			{
				Description: Description{
					Name:  "Greet",
					Short: "Shows a greeting to NAME.",
					Long:  "",
					Tags:  nil,
				},
				Optional: []string{"NAME"},
				Required: nil,
			},
		},
		Args: []Arg{
			{
				Description: Description{
					Name:  "LAST",
					Short: "Dog.",
					Long:  "Tags: foo, bar",
					Tags:  []string{"foo", "bar"},
				},
			},
			{
				Description: Description{
					Name:  "NAME",
					Short: "The user's full name.",
					Long:  "",
					Tags:  nil,
				},
			},
		},
	}

	fset, pkg := loadPackageString("github.com/justjake/examples", text)
	ui, err := Parse(fset, pkg, "*Fooer")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	fmt.Println(Serialize(ui))
	assert.Equal(t, expected, ui)

	asFile := ToFileContents(ui, "*Fooer")
	assert.Empty(t, os.Args[0])
	err = buildErrors([]file{
		{"main.go", text},
		{"generated.go", asFile},
	})
	assert.Empty(t, err)
	//assert.Equal(t, expectedOut, asFile)
}
