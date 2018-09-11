package annotation

import "go/ast"

// Func is the type of annotation functions. All annotation functions take as
// their first argument the ast.Node to which the annotation is attatched. The
// remaining interface arguments are the user-supplied literals in the
// annotation comment, or are ast.Nodes if the user supplied a type name.
//
// Any error returned will be wrapped in additional information about the
// source location and node name.
type Func func(ast.Node, ...interface{}) error
