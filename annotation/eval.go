package annotation

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"strings"
)

// Func is the type of annotation functions. All annotation functions take as
// their first argument the ast.Node to which the annotation is attatched. The
// remaining interface arguments are the user-supplied literals in the
// annotation comment, or are ast.Nodes if the user supplied a type name.
//
// Any error returned will be wrapped in additional information about the
// source location and node name.
type Func func(ast.Node, ...interface{}) error

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
	start token.Position
	end   token.Position
}

// Selector returns the referenced path as a dot-seperated string
func (r *Ref) Selector() string {
	return toStr(r.Node)
}

func (r *Ref) String() string {
	return r.GoString()
}

func (r *Ref) Find(pkg *types.Package, fset *token.FileSet) (types.Object, error) {
	scope := pkg.Scope().Innermost(r.From.Pos())
	path := strings.Split(r.Selector(), ".")
	resolver := &resolver{pkg}
	var res interface{}
	var err error
	res = scope
	for _, name := range path {
		res, err = resolver.resolve(res, name)
		if err != nil {
			return nil, &RefError{r, err}
		}
	}
	// TODO: should we check this coercion?
	res2 := res.(types.Object)
	return res2, nil
}

type resolver struct {
	pkg *types.Package
}

func (r *resolver) resolve(parent interface{}, name string) (types.Object, error) {
	switch v := parent.(type) {
	case *types.Scope:
		_, obj := v.LookupParent(name, 0)
		if obj == nil {
			return nil, fmt.Errorf("not found in scope")
		}
		return obj, nil
	case *types.Package:
		// this is dead code rn, dunno why
		r.pkg = v
		obj, err := r.resolve(v.Scope(), name)
		if err != nil {
			return nil, fmt.Errorf("%q not found in pkg %v complete %v", name, v, v.Complete())
		}
		return obj, nil
	case types.Object:
		t := v.Type()
		// TODO: is `true` the right choice here? Otherwise, we can't resolve
		// methods on pointer types...
		obj, _, _ := types.LookupFieldOrMethod(t, true, v.Pkg(), name)
		if obj == nil {
			// what if obj represents a package????? wat
			return nil, fmt.Errorf("%q not found in obj %v", name, v)
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("lookups in %v (%T) unsupported", v, v)
	}
}

// GoString implements fmt.GoStringer for Ref.
// We need this for some fairly kludgy output formatting reasons.
func (r *Ref) GoString() string {
	return fmt.Sprintf("Ref{%q}", r.Selector())
}

type RefError struct {
	*Ref
	error
}

func (err *RefError) Error() string {
	return fmt.Sprintf("%v: %v: %v", err.Ref.start, err.Ref, err.error)
}

type HitError struct {
	*Hit
	error
}

func (err *HitError) Error() string {
	return fmt.Sprintf("%v: %v", err.Hit.start, err.error)
}

func typecheck(path string, fset *token.FileSet, files []*ast.File) (*types.Package, error) {
	config := &types.Config{
		Importer: importer.Default(),
	}
	return config.Check(path, fset, files, nil)
}

// Eval evaluates the annoations in hits with the given funcs.
func Eval(hits []*Hit, funcs map[string]Func, fset *token.FileSet, pkg *types.Package) []error {
	errs := []error{}
	onErr := func(err error) {
		if err == nil {
			return
		}
		errs = append(errs, err)
	}
	for _, hit := range hits {
		fn, ok := funcs[hit.FuncName()]
		if !ok {
			onErr(&HitError{hit, fmt.Errorf("undefined annotation function %q", hit.FuncName())})
			continue
		}
		err := fn(hit.From, hit.Args...)
		if err != nil {
			onErr(&HitError{hit, err})
		}
	}
	return errs
}
