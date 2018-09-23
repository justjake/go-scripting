package annotation2

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"
	"sort"
)

const parseMode = parser.ParseComments

type Loader interface {
	// Allows adding a file that doesn't exist on disk.
	IncludeFileReader(path string, contents io.Reader)
	// Include this file when loading the package
	IncludeFile(path string)
	// Include all go files in the given directory. If filter is non-nil,
	// include paths that return true from the filter func. Does not recurse.
	IncludeDir(path string, filter func(os.FileInfo) bool)
	// Parse the included files and return a new context. Note that a partial
	// context may be returned even if there is an error value.
	//
	// We might want to call Load multiple times if we're worried about analysis
	// consumers mutating the AST!
	Load(pkgPath string) (*Package, error)
}

// The stuff loaded!
type Package struct {
	Fset   *token.FileSet // file position information
	Syntax []*ast.File    // the abstract syntax tree of each file
	Pkg    *types.Package // type information about the package
	Info   *types.Info    // type information about the syntax trees
}

func NewLoader() Loader {
	return &loader{
		filedata: make(map[string]io.Reader),
		paths:    make([]string, 0),
		dirs:     make(map[string]func(os.FileInfo) bool),
	}
}

type loader struct {
	filedata map[string]io.Reader
	paths    []string
	dirs     map[string]func(os.FileInfo) bool
}

func (l *loader) IncludeFileReader(path string, contents io.Reader) {
	l.filedata[path] = contents
}

func (l *loader) IncludeFile(path string) {
	l.paths = append(l.paths, path)
}

func (l *loader) IncludeDir(path string, filter func(os.FileInfo) bool) {
	l.dirs[path] = filter
}

func (l *loader) Load(pkgPath string) (*Package, error) {
	out := &Package{
		Fset:   token.NewFileSet(),
		Syntax: []*ast.File{},
	}

	errs := []error{}
	addErr := func(err error) bool {
		if err != nil {
			errs = append(errs, err)
			return true
		}
		return false
	}

	for dir, filter := range l.dirs {
		// If the directory couldn't be read, a nil map and the respective error
		// are returned. If a parse error occurred, a non-nil but incomplete map
		// and the first error encountered are returned
		pkgs, err := parser.ParseDir(out.Fset, dir, filter, parseMode)
		addErr(err)
		if pkgs != nil {
			for _, pkg := range pkgs {
				for _, file := range pkg.Files {
					out.Syntax = append(out.Syntax, file)
				}
			}
		}
	}

	for _, path := range l.paths {
		file, err := parser.ParseFile(out.Fset, path, nil, parseMode)
		addErr(err)
		if file != nil {
			out.Syntax = append(out.Syntax, file)
		}
	}

	for path, reader := range l.filedata {
		file, err := parser.ParseFile(out.Fset, path, reader, parseMode)
		addErr(err)
		if file != nil {
			out.Syntax = append(out.Syntax, file)
		}
	}

	// check for unrecoverable errors from which we cannot return a partial
	// package.

	// no files at all - so can't analyze anything.
	if len(out.Syntax) == 0 {
		if len(errs) > 0 {
			return nil, errs[0]
		}
		return nil, fmt.Errorf("no files parsed")
	}

	// need a deterministic ordering, so sort files by name.
	sort.Slice(out.Syntax, func(i, j int) bool {
		left := out.Syntax[i]
		right := out.Syntax[j]
		return out.FileName(left) < out.FileName(right)
	})

	// TODO: should we manually check that our files don't have multiple
	// packages?

	config := &types.Config{
		Importer:                 importer.Default(),
		DisableUnusedImportCheck: true,
	}
	// get ALL the info.
	out.Info = &types.Info{
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Implicits:  make(map[ast.Node]types.Object),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
		Scopes:     make(map[ast.Node]*types.Scope),
		InitOrder:  make([]*types.Initializer, 0),
	}
	pkg, err := config.Check(pkgPath, out.Fset, out.Syntax, out.Info)
	addErr(err)

	out.Pkg = pkg
	return out, joinErrors(errs)
}

// FileName returns the file name that contains the given node.
func (p *Package) FileName(node ast.Node) string {
	position := p.Fset.Position(node.Pos())
	return position.Filename
}

func joinErrors(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	if len(errs) == 1 {
		return errs[0]
	}
	return multierror(errs)
}

type multierror []error

func (me multierror) Error() string {
	var out bytes.Buffer
	fmt.Fprintf(&out, "Multiple errors (%d):\n", len(me))
	for _, e := range me {
		fmt.Fprintf(&out, "%v\n", e)
	}
	return out.String()
}
