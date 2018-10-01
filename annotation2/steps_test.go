package annotation2

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// this is more of an integration test...

func newTestStep(t *testing.T, name string, body func(*testing.T, UnitAPI)) (string, func(UnitAPI) (interface{}, error)) {
	return name, func(u UnitAPI) (interface{}, error) {
		t.Run(name, func(inner *testing.T) {
			body(inner, u)
		})
		return u.Input(), nil
	}
}

func TestAllSteps(t *testing.T) {
	expectedHits := 14

	loader := NewLoader()
	loader.IncludeDir("testdata", nil)
	pipeline := NewPipeline(loader)
	pipeline.AddStep("Parse", Parse)
	pipeline.AddStep(newTestStep(t, "TestParseOutput", func(t *testing.T, unit UnitAPI) {
		require.NotNil(t, unit.Input(), "emits a result")
		require.IsType(t, unit.Input(), make([]Annotation, 0), "is []Annotation")
		require.Len(t, unit.Input(), expectedHits, "valid annotations")
	}))
	pipeline.AddStep("Catalog", Catalog)
	pipeline.AddStep(newTestStep(t, "TestCatalogOutput", func(t *testing.T, unit UnitAPI) {
		require.NotNil(t, unit.Input(), "emits a result")
		require.Implements(t, (*AnnotationAPI)(nil), unit.Input(), "correct output type")
		adb := unit.Input().(AnnotationAPI)
		assert.Len(t, adb.All(), expectedHits, "db has all valid annotations")
		assert.Len(t, adb.Names(), expectedHits, "db has all names")
	}))
	err := pipeline.Run()
	assert.NoError(t, err, "pipeline successful")
}
