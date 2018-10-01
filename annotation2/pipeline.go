package annotation2

import (
	"fmt"
	"go/token"
	"os"
)

type UnitAPI interface {
	// Data exposed to the unit
	Package() *Package
	Input() interface{}

	// Results of the unit.
	// Note an error at the given position. The error will not abort the pipeline,
	// but it will be reported to the user.
	Errorf(p token.Pos, t string, v ...interface{}) error
}

type unit struct {
	name   string
	pkg    *Package
	input  interface{}
	errors []error
}

func newUnit(name string, pkg *Package, input interface{}) *unit {
	return &unit{
		name:  name,
		pkg:   pkg,
		input: input,
	}
}

func (u *unit) Package() *Package {
	return u.pkg
}

func (u *unit) Input() interface{} {
	return u.input
}

// TODO: make easier constructors for this?
// TODO: should this hold a Spliced and a *FileSet?
type Error struct {
	token.Position
	error
	message string
}

func (e *Error) Error() string {
	if e.message != "" {
		return fmt.Sprintf("%v: %v", e.Position, e.message)
	}
	return fmt.Sprintf("%v: %v", e.Position, e.error)
}

func (u *unit) Errorf(p token.Pos, t string, v ...interface{}) error {
	position := u.Package().Fset.Position(p)
	msg := fmt.Sprintf(t, v...)
	err := &Error{position, nil, msg}
	if len(u.errors) == 0 {
		u.errors = []error{err}
	} else {
		u.errors = append(u.errors, err)
	}
	// XXX: remove fmt.Fprintln here?
	fmt.Fprintln(os.Stderr, fmt.Sprintf("(%s)", u.name), err)
	return err
}

type Runnable func(UnitAPI) (interface{}, error)

type Pipeline interface {
	// Add a step to the pipeline, which will run after the previous step.
	// Steps have a name and run a function of a Unit.
	// Steps may log non-fatal errors using UnitAPI.Errorf.
	// Steps return a result, and an optional error. If an error is returned by
	// any step, the pipeline aborts there and does not continue.
	AddStep(name string, run Runnable)
	// What you'd expect
	Run() error
}

func NewPipeline(loader Loader) Pipeline {
	return &pipeline{
		loader: loader,
		steps:  make([]step, 0, 1),
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

func (p *pipeline) AddStep(name string, run Runnable) {
	p.steps = append(p.steps, step{name, run})
}

func (p *pipeline) Run() error {
	pkg, err := p.loader.Load()
	if pkg == nil {
		return err
	}
	out := interface{}(nil)
	for i, s := range p.steps {
		unit := newUnit(s.name, pkg, out)
		out, err = s.run(unit)
		if err != nil {
			return fmt.Errorf("step %d %q: %v", i+1, s.name, err)
		}
	}
	// TODO: UnitAPI has an error logging facility. What to do about that?
	return nil
}
