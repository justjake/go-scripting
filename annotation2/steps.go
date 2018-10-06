package annotation2

import (
	"fmt"
	"go/ast"
	"go/types"
	"reflect"
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

type DispatchStep struct {
	Funcs map[string]interface{}
	Out   interface{}
	Unit  UnitAPI
}

func (ds *DispatchStep) On(name string, cb interface{}) *DispatchStep {
	if ds.Funcs == nil {
		ds.Funcs = make(map[string]interface{})
	}
	ds.Funcs[name] = cb
	return ds
}

func (ds *DispatchStep) Run(unit UnitAPI) (interface{}, error) {
	ds.Unit = unit
	db, ok := unit.Input().(AnnotationAPI)
	if !ok {
		return nil, fmt.Errorf("input %T is not an AnnotationAPI: %v", unit.Input(), unit.Input())
	}
	for _, hit := range db.All() {
		if fn, ok := ds.Funcs[hit.Name()]; ok {
			args := append([]interface{}{hit}, hit.Args()...)
			err := CallFunc(fn, args...)
			if err != nil {
				return ds.Out, unit.Errorf(hit.Pos(), "%v: %v", hit, err)
			}
		} else {
			// not found
			unit.Errorf(hit.Pos(), "handler undefined for %q in %v", hit.Name(), hit)
		}
	}
	return ds.Out, nil
}

// DefaultPipeline builds a pipeline that runs Parse and Catalog steps, handing
// a completed AnnotationAPI as the input to the next step added.
func DefaultPipeline(loader Loader) Pipeline {
	pipeline := NewPipeline(loader)
	pipeline.AddStep("annotation2.Parse", Parse)
	pipeline.AddStep("annotation2.Catalog", Catalog)
	return pipeline
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
		var canonicalError *error
		errtype := reflect.TypeOf(canonicalError).Elem()
		out := ftype.Out(0)
		if out != errtype {
			return fmt.Errorf("fn should return %v, but instead returns %v", errtype, out)
		}
	}
	if ftype.NumIn() != len(args) {
		return fmt.Errorf("need %d args, have %d args", ftype.NumIn(), len(args))
	}
	vargs := make([]reflect.Value, len(args))
	for i, arg := range args {
		need := ftype.In(i)
		have := reflect.TypeOf(arg)
		if have != need && !have.Implements(need) {
			return fmt.Errorf("arg %d: need %v, have %v", i, need, have)
		}
		vargs[i] = reflect.ValueOf(arg)
	}
	// ok, should be good to go?
	out := fval.Call(vargs)
	if len(out) == 1 {
		if err, ok := out[0].Interface().(error); ok {
			return err
		}
		if out[0].Interface() == nil {
			return nil
		}
		return fmt.Errorf("typed as error, but, not an error: %v", out[0])
	}
	if len(out) > 1 {
		// unreachable??
		return fmt.Errorf("expected one return value, got %d", len(out))
	}
	return nil
}
