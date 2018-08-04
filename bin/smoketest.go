package main

import (
	"fmt"
	"github.com/justjake/go-scripting/env"
	"github.com/justjake/go-scripting/shell"
	"os"
)

var Vars = env.NewVars()
var Args = env.SystemArgs()
var Sh = &shell.Shell{
	Stdout: os.Stdout,
	Stderr: os.Stderr,
}

var FIRST = Vars.DefaultGetter("FIRST", "Jake")
var LAST = Vars.DefaultGetter("LAST", "Teton-Landis")
var NAME = Vars.MemoGetter("NAME", func() string {
	return FIRST() + " " + LAST()
})

func main() {
	fmt.Println(shell.Escape("; exit 1"))
	fmt.Println(shell.ScriptPrintf("%s", "; exit 1"))
	Sh.Runf("echo first: %s, last: %s, name: %s", FIRST(), LAST(), NAME())
}
