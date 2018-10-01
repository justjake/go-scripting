package annotation2

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func parseTestFile(filename string) (*token.FileSet, *ast.Package) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}
	pkg := &ast.Package{
		Name: file.Name.Name,
		Files: map[string]*ast.File{
			filename: file,
		},
	}
	return fset, pkg
}

type parseErrorLog struct {
	*token.FileSet
	bytes.Buffer
}

func (l *parseErrorLog) Errorf(p token.Pos, f string, v ...interface{}) error {
	position := l.Position(p)
	err := fmt.Errorf(position.String()+": "+f, v...)
	fmt.Fprintln(&l.Buffer, err)
	return err
}

func TestParse(t *testing.T) {
	expectedHits := strings.TrimSpace(`
Annotation{OnLoneImport()}
Annotation{OnNamedImport()}
Annotation{OnDotImport()}
Annotation{OnNormalImport()}
Annotation{OnType()}
Annotation{OnField()}
Annotation{OnFunc()}
Annotation{OnGroupedConst()}
Annotation{OnGroupedConstNoValue()}
Annotation{OnLoneVar()}
Annotation{OnDoubleVar()}
Annotation{Literals("a string", 5, -0.125)}
Annotation{LocalRefs(Ref{Thing}, Ref{Thing.Greeting}, Ref{Thing.Name}, Ref{somePriv})}
Annotation{RemoteRefs(Ref{fmt}, Ref{fmt.Stringer}, Ref{fmt.Stringer.String}, Ref{fmt.Sprintf})}
	`)

	expectedErrs := strings.TrimSpace(`
testdata/annotation_types.go:65:5: NotACall.Foo.Bar + 1: not a func call, instead *ast.BinaryExpr
testdata/annotation_types.go:66:23: BadCallSyntax(foo bar): missing ',' in argument list
testdata/annotation_types.go:67:17: BadCallMath(1 + 1): unsupported syntax "1 + 1"
testdata/annotation_types.go:68:21: BadCallFn(-555, Foo.Bar()): unsupported syntax "Foo.Bar()"
	`)
	fset, pkg := parseTestFile("testdata/annotation_types.go")

	log := &parseErrorLog{FileSet: fset}
	parser := &Parser{log.Errorf}
	hits := parser.Parse(pkg)

	allHits := join(len(hits), func(i int) string { return hits[i].String() })
	assert.Equal(t, expectedHits, allHits)
	assert.Equal(t, expectedErrs, strings.TrimSpace(log.String()))
}

func join(l int, f func(i int) string) string {
	res := make([]string, l)
	for i := range res {
		res[i] = f(i)
	}
	return strings.Join(res, "\n")
}
