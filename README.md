# go-scripting

Several experiments for making go an easy language for scripting tasks.

## shell

the idea is to make go a convinient wrapper for bash shellouts. maybe
worthwhile. maybe rude to make people use both go and bash.

## annotation / annotation2

After writing a few codegen tools in this project, I noticed a common pattern
of attaching some kinds of annotations to a declaration via comments, and then
the codegen tool parses the comments to produce new code.

The `annotation` package attemppts to formalize the "parse" and "attach" parts
using `// @AnnotateFunc()` syntax in comments.

See the [annotation2 design doc](./annotation2/design.md) for more information.

See [./bin/static_compose.go](./bin/static_compose.go) for an example that uses
annotation2 to generate [shell_format_methods.go](./shell/shell_format_methods.go).

## cli

An (ambitious) auto-generated CLI using annotations, function signatures, and
codegen. Inspired by Google's [python-fire](https://github.com/google/python-fire) library.

## env

Abstracts the args and env vars of a script. Of dubious value.
