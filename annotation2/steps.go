package annotation2

import (
	"fmt"
	"go/ast"
	"go/types"
	"sort"
)

type AnnotationAPI interface {
	// Retrieve annotation
	All() []Annotation
	Named(name string) []Annotation
	ForNode(node ast.Node) []Annotation
	ForObj(obj types.Object) []Annotation
	ForPkg(pkg *types.Package) []Annotation
	// Get all annotated objects and packages and annotation names
	Names() []string
	Objs() []types.Object
	Pkgs() []*types.Package
	Nodes() []ast.Node
}

type adb struct {
	all  []Annotation
	name map[string][]Annotation
	pkg  map[*types.Package][]Annotation
	obj  map[types.Object][]Annotation
	node map[ast.Node][]Annotation
}

func newAdb() *adb {
	return &adb{
		name: make(map[string][]Annotation),
		pkg:  make(map[*types.Package][]Annotation),
		obj:  make(map[types.Object][]Annotation),
		node: make(map[ast.Node][]Annotation),
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

func (db *adb) ForNode(node ast.Node) []Annotation {
	return append([]Annotation{}, db.node[node]...)
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

func (db *adb) Nodes() []ast.Node {
	nodes := make([]ast.Node, 0, len(db.node))
	for k := range db.node {
		nodes = append(nodes, k)
	}
	return nodes
}

func (db *adb) addObj(obj types.Object, ann Annotation) {
	db.obj[obj] = append(db.obj[obj], ann)
	db.add(ann)
}

func (db *adb) addpkg(pkg *types.Package, ann Annotation) {
	db.pkg[pkg] = append(db.pkg[pkg], ann)
	db.add(ann)
}

func (db *adb) add(ann Annotation) {
	db.all = append(db.all, ann)
	db.name[ann.Name()] = append(db.name[ann.Name()], ann)
	db.node[ann.From()] = append(db.node[ann.From()], ann)
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
	return []Annotation(hits), nil
}

// Catalog recieves a []Annotation via Input, and builds a database
// of the types and nodes of each annotation and ref for querying and lookup.
func Catalog(unit UnitAPI) (interface{}, error) {
	info := unit.Package().Info
	pkg := unit.Package().Pkg
	hits := unit.Input().([]Annotation)
	db := newAdb()
	for _, hit := range hits {
		obj, err := LookupObject(info, hit.From())
		if obj == nil {
			unit.Errorf(hit.From().Pos(), "%v: cannot find anchor object: %v", hit, err)
			db.add(hit)
			continue
		}
		db.addObj(obj, hit)
		// also log errors about unresolvable refs, although we take no action.
		// todo: with a Lemma DB, we could store back-references to the hit.
		for _, arg := range hit.Args() {
			if ref, ok := arg.(Ref); ok {
				_, err := LookupRef(pkg, ref)
				if err != nil {
					unit.Errorf(ref.Pos(), "warning: %v: cannot find referenced object: %v", ref, err)
				}
			}
		}
	}
	// TODO
	return AnnotationAPI(db), nil
}
