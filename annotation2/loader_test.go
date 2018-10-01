package annotation2

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader(t *testing.T) {
	file := "testdata/annotation_types.go"
	dir := "testdata"

	var loader Loader

	table := []struct {
		name string
		run  func(t *testing.T)
	}{
		{"IncludeFileReader", func(t *testing.T) {
			loader = NewLoader()
			f, err := os.Open(file)
			require.NoError(t, err)
			loader.IncludeFileReader(file, f)
		}},
		{"IncludeFile", func(t *testing.T) {
			loader = NewLoader()
			loader.IncludeFile(file)
		}},
		{"IncludeDir no filter", func(t *testing.T) {
			loader = NewLoader()
			loader.IncludeDir(dir, nil)
		}},
	}

	for _, c := range table {
		t.Run(c.name, func(t *testing.T) {
			c.run(t)
			pkg, err := loader.Load()
			require.NoError(t, err)
			require.NotNil(t, pkg)
			assert.NotNil(t, pkg.Fset, "has FileSet")
			assert.Len(t, pkg.Syntax, 1, "has one file")
			assert.NotNil(t, pkg.Pkg, "has Pkg")
			assert.NotNil(t, pkg.Info, "has Info")
		})
	}
}
