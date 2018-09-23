package annotation2

import "fmt"

type UnitAPI interface{}

type lemmaDB struct{}

func newUnit(pkg *Package, lemmas *lemmaDB, input interface{}) UnitAPI {
	return nil
}

type Pipeline interface {
	// Add a step to the pipeline, which will run after the previous step.
	// Steps have a name and run a function of a Unit.
	// Steps may log non-fatal errors using UnitAPI.Errorf.
	// Steps return a result, and an optional error. If an error is returned by
	// any step, the pipeline aborts there and does not continue.
	AddStep(name string, run func(UnitAPI) (interface{}, error))
	// What you'd expect
	Run() error
}

func NewPipeline(loader Loader) Pipeline {
	return &pipeline{
		loader: loader,
		steps:  make([]step, 1),
		// TODO: always append our "Parse annotations" step?
	}
}

type step struct {
	name string
	run  func(UnitAPI) (interface{}, error)
}

type pipeline struct {
	steps  []step
	loader Loader
}

func (p *pipeline) AddStep(name string, run func(UnitAPI) (interface{}, error)) {
	p.steps = append(p.steps, step{name, run})
}

func (p *pipeline) Run() error {
	pkg, err := p.loader.Load()
	if pkg == nil {
		return err
	}
	// TODO: finish
	// need to figure out UnitAPI impl first.
	lemmas := &lemmaDB{}
	out := interface{}(nil)
	for i, s := range p.steps {
		unit := newUnit(pkg, lemmas, out)
		out, err = s.run(unit)
		if err != nil {
			return fmt.Errorf("step %d %q: %v", i+1, s.name, err)
		}
	}
	// TODO: UnitAPI has an error logging facility. What to do about that?
	return nil
}
