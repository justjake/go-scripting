# annotation design

## Goals

This package aims to make code generation tools easier to implement using a
common syntax for attatching non-syntactic information to Go declarations using
comments.

Tools using the annotation package should be able to easily perform the
following tasks:

1. read, parse, and validate go files or packages containing annotation comments,
1. query the annotations in a package,
1. query the declarations in a package, which can be referenced from an annotation,
   for both their AST structure and their types,
1. synthesize and emit new go files for a package,

## Approach

### Current State

After several iterations on the annotation package, it's clear that I need to sit back and write a design doc before getting mired in implementaiton details.

From the latest idealized user story and thinking about the APIs we need, it's
clear that the current approach in annotation.Hit and annotation.Ref are not
sufficient, on their own, to fufill the querying needs of consumers.

annotation.Hit and annotation.Ref must both reference all the AST stuff,
and all the type stuff, or, such stuff must be available in another object
with query methods that take Ref or Hit.

Also, after investigation, it appears that the go "tools" repo has a ton of
utilities that will simplify our implementation, especially things that
supercede custom implementations in annotation.

### Notes on golang.org/x/tools/go

We could use this to structure our traversal of source:
https://godoc.org/golang.org/x/tools/go/analysis

Coming soon is a new system for AST traversals:
https://go-review.googlesource.com/c/tools/+/135655

In general golang.org/x/tools/go/types/typeutil looks
great, especially https://godoc.org/golang.org/x/tools/go/types/typeutil#Map
which can help store derived state for types for later
program synthesis, although the Lemma system in anaylsis will
be even more convinient.
Also cool: https://godoc.org/golang.org/x/tools/go/types/typeutil#MethodSetCache

In the astutil package, there's this:
https://godoc.org/golang.org/x/tools/go/ast/astutil#PathEnclosingInterval
which maps from a Pos to a Node, which can be used to go from types.Object ->
token.Pos -> ast.Node.

### Loading

We need to load the files targeted by the user, parse them, and typecheck them,
so we can hand everything off to the analysis phases.

```go
type Loader interface {
  // Allows adding a file that doesn't exist on disk.
  IncludeFileContents(path string, contents []byte)
  // Include this file when loading the package
  IncludeFile(path string)
  // Include all go files in the given directory. If filter is non-nil,
  // include paths that return true from the filter func. Does not recurse.
  IncludeDir(path string, filter func(string) bool)
  // Parse the included files and return a new context. Note that a partial
  // context may be returned even if there is an error value.
  //
  // We might want to call Load multiple times if we're worried about analysis
  // consumers mutating the AST!
  Load() (*Package, error)
}

// The stuff loaded!
type Package struct {
    Fset   *token.FileSet // file position information
    Syntax []*ast.File    // the abstract syntax tree of each file
    Pkg    *types.Package // type information about the package
    Info   *types.Info    // type information about the syntax trees
}
```

### Analysis

We should follow the model of go/analysis, with several phases that operate with all
analysis unit data that flow into each other. Each unit operates on this struct:

```go
// Some fields and comments omitted.
type Unit struct {
    // syntax and type information
    Fset   *token.FileSet // file position information
    Syntax []*ast.File    // the abstract syntax tree of each file
    Pkg    *types.Package // type information about the package
    Info   *types.Info    // type information about the syntax trees

    // ObjectLemma retrieves a lemma associated with obj.
    // Given a value ptr of type *T, where *T satisfies Lemma,
    // ObjectLemma copies the value to *ptr.
    //
    // ObjectLemma may panic if applied to a lemma type that
    // the analysis did not declare among its LemmaTypes,
    // or if called after analysis of the unit is complete.
    //
    // ObjectLemma is not concurrency-safe.
    ObjectLemma func(obj types.Object, lemma Lemma) bool
    // Similar.
    PackageLemma func(pkg *types.Package, lemma Lemma) bool

    // SetObjectLemma associates a lemma of type *T with the obj,
    // replacing any previous lemma of that type.
    //
    // SetObjectLemma panics if the lemma's type is not among
    // Analysis.LemmaTypes, or if obj does not belong to the package
    // being analyzed, or if it is called after analysis of the unit
    // is complete.
    //
    // SetObjectLemma is not concurrency-safe.
    SetObjectLemma func(obj types.Object, lemma Lemma)
    // See comments for SetObjectLemma.
    SetPackageLemma func(lemma Lemma)

    // Output is an immutable result computed by this analysis unit
    // and set by the Run function.
    // It will be made available as an input to any analysis that
    // depends directly on this one; see Analysis.Requires.
    // Its type must match Analysis.OutputType.
    //
    // Outputs are available as Inputs to later analyses of the
    // same package. To pass analysis results between packages (and
    // thus potentially between address spaces), use Lemmas.
    Output interface{}
}
```

To prevent depending directly on the still-under-development
golang.org/x/tools/go/analysis, we should produce an interface that distills
the useful features for our purpose. This would be the first argument passed to
the `run()` function of any consumer. An idea:

```go
type UnitAPI interface {
    // syntax and type information
    Fset() *token.FileSet // file position information
    Syntax() []*ast.File    // the abstract syntax tree of each file
    Pkg() *types.Package // type information about the package
    Info() *types.Info    // type information about the syntax trees

    // We might omit these in the first iteration.
    ObjectLemma(obj types.Object, lemma Lemma) bool
    PackageLemma(pkg *types.Package, lemma Lemma) bool
    SetObjectLemma func(obj types.Object, lemma Lemma)
    SetPackageLemma func(pkg *types.Package, lemma Lemma)

    // Simplified for linear units:
    // Get previous analysis of this package
    // For user packages, this returns an AnnotationAPI.
    Input() interface{}
    // XXX: maybe just return (out interface{}, err error)?
    // Pass to the next analysis of this package
    SetOutput(interface{})

    // This might be helpful, too.
    // Add an error to this unit for the given node.
    Errorf(token.Pos, t string, v ...interface{}) error
}
```

In this system, we would add annotations as lemmas to objects. We probably also want
to record a database of all annotations, so that can be queried too.

```go
type AnnotationAPI interface {
  All() []*annotation.Annotation
  Named(name string) []*annotation.Annotation
  ForObj(obj types.Object) []*annotation.Annotation
  ForPkg(pkg *types.Package) []*annotation.Annotation
  // Get all annotated objects and packages
  Objs() []types.Object
  Pkgs() []*types.Package

  // implied private methods:
  addObj(types.Object, *annotation.Annotation)
  addPkg(*types.Package, *annotation.Annotation)
}
```

Our architecture is then a clear multi-phase system. The first unit we run is
the "annotation" unit, which inspects the package's AST, and constructs an
AnnotationAPI. Although it's unclear if necessary, annotation lemmas are added
to each object. An AnnotationAPI is the output of the first unit, and the input
to the second unit.

Successive user-defined units then run, each passing the AnnotationAPI or a
custom output on to the next unit as input.


```
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
  // ...
}
```

This is significantly simplified from the []


### Synthesis

The final unit should construct and emit a Synthesis value. I don't have a good
idea of what the API of Synthesis is, but it should at least provide a
declarative way to emit Go files. Perhaps:

```go
type Synthesis interface {
  // Create or retrieve a handle to the given file name, which will eventually be
  // into the package's directory
  File(basename string) GoSourceFile
}

type GoSourceFile interface {
  // can just read/write to it as normal
  io.Writer
  // maybe offer some other utility methods?
  // like those in https://godoc.org/golang.org/x/tools/go/ast/astutil
  // maybe a lot of shenanigans are needed wrapping a token.FileSet?
  // like we use AST files??
  //
  // If we mutate in-place our ASTs we will be sad, so maybe some util to
  // duplicate ASTs is needed?
}
```

Another system for golang generation is https://github.com/kubernetes/gengo,
which was split out of kubernetes. It unfortunatley uses its own type system
and such, but has some comprehensive examples of generators, and its own
kind-of contex stuff. Taking a deeper look at
https://godoc.org/k8s.io/gengo/generator may serve as a good inspiration.
Specifically, there's a whole system for naming variables and other things,
which is cool.

Gengo seems to be focused on generating one or more complete packages, where I
expect that the annotation system will mostly be used for generating
boilerplate files inside a mostly-hand-written package.
