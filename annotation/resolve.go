package annotation

import (
	"fmt"
	"go/types"
	"strings"
)

// This file deals with type checking and resolution.

func ResolveTypes(hits []*Hit, pkg *types.Package) []error {
	errs := []error{}
	onErr := func(err error) {
		if err == nil {
			return
		}
		errs = append(errs, err)
	}
	for _, hit := range hits {
		// populate any refs in hit with type information. we could try to do this
		// earlier - like at ref creation, already have checked the types or
		// something. Seems bad, but also a lesser evil than mixing types into the
		// parse process.
		//
		// Maybe Parse() should secretly return a list of all refs? Ew.
		for _, arg := range hit.Args {
			if ref, ok := arg.(*Ref); ok {
				objs, err := resolveRef(ref, pkg)
				onErr(err)
				if err != nil {
					return err
				}
			}
		}
	}
	return errs
}

func resolveRef(r *Ref, pkg *types.Package) ([]types.Object, error) {
	scope := pkg.Scope().Innermost(r.From.Pos())
	path := strings.Split(r.Selector(), ".")
	var res interface{}
	var err error
	res = scope
	for _, name := range path {
		res, err = ResolveName(res, name)
		if err != nil {
			return nil, &RefError{r, err}
		}
	}
	// TODO: should we check this coercion?
	res2 := res.(types.Object)
	return res2, nil
}

// ResolveName finds the types.Object for the given name in a value from
// go/types. Support parent types are *types.Scope, *types.Package, and
// types.Object.
func ResolveName(parent interface{}, name string) (types.Object, error) {
	switch v := parent.(type) {
	case *types.Scope:
		_, obj := v.LookupParent(name, 0)
		if obj == nil {
			return nil, fmt.Errorf("%q not found in scope", name)
		}
		return obj, nil
	case *types.Package:
		obj, err := ResolveName(v.Scope(), name)
		if err != nil {
			return nil, fmt.Errorf("%q not found in %v", v)
		}
		return obj, nil
	case *types.PkgName:
		// a type of object describing a package name ref
		return ResolveName(v.Imported(), name)
	case types.Object:
		// all other objects
		t := v.Type()
		// TODO: is `true` the right choice here? Otherwise, we can't resolve
		// methods on pointer types...
		obj, _, _ := types.LookupFieldOrMethod(t, true, v.Pkg(), name)
		if obj == nil {
			return nil, fmt.Errorf("%q not found in %v", name, v)
		}
		return obj, nil
	default:
		return nil, fmt.Errorf("lookups in %v (%T) unsupported", v, v)
	}
}
