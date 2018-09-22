package cli

import "github.com/justjake/go-scripting/annotation"

var ui = &UI{
	Commands: []Command{},
	Args:     []Arg{},
}

func CLI(hit *annotation.Hit) {
	// create a command for each of the hit type's public
	// methods
}
