package annotation

import (
	"go/ast"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEval(t *testing.T) {
	fset, pkg := parseTestFile("testdata/annotation_types.go")
	hits, _ := Parse(fset, pkg)

	typed, err := typecheck(
		"main",
		fset,
		[]*ast.File{pkg.Files["testdata/annotation_types.go"]},
	)

	assert.Empty(t, err, "no type errors")

	makefunc := func(name string) Func {
		return func(n ast.Node, args ...interface{}) error {
			t.Logf("Called %q w/ node %v, args: %v", name, n, args)
			for _, arg := range args {
				if ref, ok := arg.(*Ref); ok {
					obj, err := ref.Find(typed, fset)
					t.Logf("Obj: %v, err: %v", obj, err)
				}
			}
			return nil
		}
	}

	funcs := map[string]Func{
		"OnImport":   makefunc("OnImport"),
		"OnType":     makefunc("OnType"),
		"OnField":    makefunc("OnField"),
		"OnVar":      makefunc("OnVar"),
		"Literals":   makefunc("Literals"),
		"RemoteRefs": makefunc("RemoteRefs"),
		"LocalRefs":  makefunc("LocalRefs"),
	}

	errs := Eval(hits, funcs, fset, typed)

	// TODO: more robust output testing
	assert.Empty(t, errs, "no eval errs")
}
