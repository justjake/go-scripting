package script

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"go/doc"
	"go/parser"
	"go/token"
	"testing"
)

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
func LAST() {
	return os.Getenv("LAST")
}

// Show shows all the things. Use show when you need an extensive greeting.
func Show() {
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

	expectedOut := `wat`

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
	ui, err := Parse(fset, pkg)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	fmt.Println(Serialize(ui))
	assert.Equal(t, expected, ui)

	asFile := ToFileContents(ui)
	assert.Equal(t, expectedOut, asFile)
}
