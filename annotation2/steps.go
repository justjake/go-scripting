package annotation2

import (
	"fmt"
	"go/types"
	"sort"
)

type AnnotationAPI interface {
	// Retrieve annotation
	All() []Annotation
	Named(name string) []Annotation
	ForObj(obj types.Object) []Annotation
	ForPkg(pkg *types.Package) []Annotation
	// Get all annotated objects and packages and annotation names
	Names() []string
	Objs() []types.Object
	Pkgs() []*types.Package
}

type adb struct {
	all  []Annotation
	name map[string][]Annotation
	pkg  map[*types.Package][]Annotation
	obj  map[types.Object][]Annotation
}

func newAdb() *adb {
	return &adb{
		name: make(map[string][]Annotation),
		pkg:  make(map[*types.Package][]Annotation),
		obj:  make(map[types.Object][]Annotation),
	}
}

func (db *adb) All() []Annotation {
	return append([]Annotation{}, db.all...)
}

func (db *adb) Named(name string) []Annotation {
	return append([]Annotation{}, db.name[name]...)
}

func (db *adb) ForObj(obj types.Object) []Annotation {
	return append([]Annotation{}, db.obj[obj]...)
}

func (db *adb) ForPkg(pkg *types.Package) []Annotation {
	return append([]Annotation{}, db.pkg[pkg]...)
}

func (db *adb) Names() []string {
	names := make([]string, 0, len(db.name))
	for k := range db.name {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

func (db *adb) Objs() []types.Object {
	objs := make([]types.Object, 0, len(db.obj))
	for k := range db.obj {
		objs = append(objs, k)
	}
	return objs
}

func (db *adb) Pkgs() []*types.Package {
	pkgs := make([]*types.Package, 0, len(db.pkg))
	for k := range db.pkg {
		pkgs = append(pkgs, k)
	}
	return pkgs
}

func (db *adb) addObj(obj types.Object, ann Annotation) {
	db.all = append(db.all, ann)
	db.obj[obj] = append(db.obj[obj], ann)
	db.name[ann.Name()] = append(db.name[ann.Name()], ann)
}

func (db *adb) addpkg(pkg *types.Package, ann Annotation) {
	db.all = append(db.all, ann)
	db.pkg[pkg] = append(db.pkg[pkg], ann)
	db.name[ann.Name()] = append(db.name[ann.Name()], ann)
}

// Parse parses all the annotations in the unit's AST.
func Parse(unit UnitAPI) (interface{}, error) {
	hits := []Annotation{}
	parser := Parser{unit.Errorf}
	files := unit.Package().Syntax
	for _, file := range files {
		hits = append(hits, parser.Parse(file)...)
	}
	if len(hits) < 0 {
		return nil, fmt.Errorf("parsed nothing in %d files", len(files))
	}
	return hits, nil
}

// Catalog recieves a []Annotation via Input, and builds a database
// of the types and nodes of each annotation and ref for querying and lookup.
func Catalog(unit UnitAPI) (interface{}, error) {
	info := unit.Package().Info
	pkg := unit.Package().Pkg
	hits := unit.Input().([]Annotation)
	db := newAdb()
	for _, hit := range hits {
		obj := hit.Object(unit.Package())
		if obj == nil {
			unit.Errorf(hit.Pos(), "%v: cannot find typed object", hit)
			continue
		}
		db.addObj(obj, hit)
		// also log errors about unresolvable refs, although we take no action.
		for _, arg := range hit.Args() {
			if ref, ok := arg.(Ref); ok {
				if refobj := ref.Object(unit.Package()); refobj == nil {
					unit.Errorf(ref.Pos(), "warning: %v: cannot find typed object", ref)
				}
			}
		}
	}
	// TODO
	return &AnnotationDB{}, nil
}
