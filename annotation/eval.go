package annotation

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/token"
	"go/types"
	"reflect"
)

// Func is the type of annotation functions, which apply a Hit. Any error
// returned will be wrapped in additional information about Hit's source
// location.
type Func func(*Hit) error

// WrapFunc wraps any function with a type-safe dispatch of a Hit's arguments.
// Use WrapFunc around your annotation implementations to avoid type-switching
// each argument.
//
// See CallFunc for the restrictions on fn's type.
func WrapFunc(fn interface{}) Func {
	return func(hit *Hit) error {
		// TODO: hit as first arg?
		return CallFunc(fn, hit.Args)
	}
}

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
	// The object referred to.
	types.Object
	// The node the annotation is attatched to.
	From ast.Node
	// The reference syntax, parsed from an annotation comment. It's type is
	// either an *ast.Ident or an *ast.SelectorExpr.
	ast.Node
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
func Eval(hits []*Hit, funcs map[string]Func) []error {
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

		// populate any refs in hit with type information. we could try to do this
		// earlier - like at ref creation, already have checked the types or
		// something. Seems bad, but also a lesser evil than mixing types into the
		// parse process.
		//
		// Maybe Parse() should secretly return a list of all refs? Ew.
		for _, arg := range hit.Args {
			if ref, ok := arg.(*Ref); ok {
				ref.pkg = pkg
			}
		}
		err := fn(hit.From, hit.Args...)
		if err != nil {
			onErr(&HitError{hit, err})
		}
	}
	return errs
}

// CallFunc performs a type-safe call of fn, which can be a void function, or a
// function that returns an error. If any of the argument types do not match
// the function's signature, an error is returned.
// Functions with a ... argument are not supported.
// CallFunc will never panic due to mistyped arguments.
func CallFunc(fn interface{}, args ...interface{}) error {
	fval := reflect.ValueOf(fn)
	ftype := fval.Type()
	if fval.Kind() != reflect.Func {
		return fmt.Errorf("fn not a func, instead %T", fn)
	}
	if ftype.IsVariadic() {
		return fmt.Errorf("variadic fn not supported")
	}
	if ftype.NumOut() > 1 {
		return fmt.Errorf("fn returns >1 value")
	}
	if ftype.NumOut() == 1 {
		out := ftype.Out(0)
		err := fmt.Errorf("fn returns non error %v", out)
		if out != reflect.TypeOf(err) {
			return err
		}
	}
	if ftype.NumIn() != len(args) {
		return fmt.Errorf("need %d args, have %d args", ftype.NumIn(), len(args))
	}
	vargs := make([]reflect.Value, len(args))
	for i, arg := range args {
		need := ftype.In(i)
		have := reflect.TypeOf(arg)
		if have != need {
			return fmt.Errorf("arg %d: need %v, have %v", need, have)
		}
		vargs[i] = reflect.ValueOf(arg)
	}
	// ok, should be good to go?
	out := fval.Call(vargs)
	if len(out) == 1 {
		if err, ok := out[0].Interface().(error); ok {
			return err
		}
		return fmt.Errorf("typed as error, but, not an error: %v", out[0])
	}
	if len(out) > 1 {
		// unreachable??
		return fmt.Errorf("expected one return value, got %d", len(out))
	}
	return nil
}
