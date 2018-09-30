package annotation2

import (
	"fmt"
	"go/ast"
	"go/types"
	"strings"
)

// This file holds functions for performing type lookups and queries.

// LookupName finds the types.Object for the given name in a value from
// go/types. Support parent types are *types.Scope, *types.Package, and
// types.Object.
func LookupName(parent interface{}, name string) (types.Object, error) {
	switch v := parent.(type) {
	case *types.Scope:
		_, obj := v.LookupParent(name, 0)
		if obj == nil {
			return nil, fmt.Errorf("%q not found in scope", name)
		}
		return obj, nil
	case *types.Package:
		obj, err := LookupName(v.Scope(), name)
		if err != nil {
			return nil, fmt.Errorf("%q not found in %v", name, v)
		}
		return obj, nil
	case *types.PkgName:
		// a type of object describing a package name ref
		return LookupName(v.Imported(), name)
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

// XXX NEEDS WORK
func LookupObject(info *types.Info, unknown ast.Node) (types.Object, error) {
	switch node := unknown.(type) {
	case *ast.Field:
		if len(node.Names) == 0 {
			// anonymous field.
			// TODO: figure this one out, it should be possible to get the types.Var.
			return nil, fmt.Errorf("%T is anonymous field: %v", node, node)
		}
		// all names point to the same field, right??
		return info.ObjectOf(node.Names[0]), nil
	case *ast.FuncDecl:
		return info.ObjectOf(node.Name), nil
	case *ast.GenDecl:
		return nil, fmt.Errorf("%T contains []Spec, try one of those: %v", node, node)
	case *ast.ImportSpec:
		// TODO: construct or otherwise divine a *types.PkgName!
		return nil, fmt.Errorf("%T unimplemented (should return *types.PkgName): %v", node, node)
	case *ast.TypeSpec:
		return info.ObjectOf(node.Name), nil
	case *ast.ValueSpec:
		if len(node.Names) == 1 && len(node.Values) == 1 {
			return info.ObjectOf(node.Names[0]), nil
		}
		// ambiguous reference
		return nil, fmt.Errorf("%T is ambigous because names %d !== values %d !== 1: %v", node, len(node.Names), len(node.Values), node)
	default:
		return nil, fmt.Errorf("unsupported node type %T: %v", node, node)
	}
}

func LookupRef(pkg *types.Package, r Ref) ([]types.Object, error) {
	scope := pkg.Scope().Innermost(r.From().Pos())
	path := strings.Split(r.Selector(), ".")
	objs := []types.Object{}
	var obj types.Object
	var err error
	for i, name := range path {
		if i == 0 {
			obj, err = LookupName(scope, name)
		} else {
			obj, err = LookupName(obj, name)
		}
		if err != nil {
			// TODO: show spliced position in error?
			return objs, err
		}
		objs = append(objs, obj)
	}
	return objs, nil
}
