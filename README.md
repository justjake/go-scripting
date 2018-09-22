# go-scripting

Several experiments for making go an easy language for scripting tasks.

## shell

the idea is to make go a convinient wrapper for bash shellouts. maybe
worthwhile. maybe rude to make people use both go and bash.

## annotation

After writing a few codegen tools in this project, I noticed a common pattern
of attaching some kinds of annotations to a declaration via comments, and then
the codegen tool parses the comments to produce new code.

The `annotation` package attemppts to formalize the "parse" and "attach" parts
using `// @AnnotateFunc()` syntax in comments.

## cli

An (ambitious) auto-generated CLI using annotations, function signatures, and
codegen. Inspired by Google's [python-fire](https://github.com/google/python-fire) library.

## env

Abstracts the args and env vars of a script. Of dubious value.
