package main

// go:generate go run ./generate_script_ui.go -out ui.go

import (
	"fmt"
	"github.com/justjake/go-scripting/env"
	"github.com/justjake/go-scripting/shell"
	"os"
)

// UI provides our main and such
var UI = &script.UI{
	Name:  "smoketest",
	Short: "validates ideas in go scripting",
	Long: `Smoketest uses different featuers from the go-scripting packages. It
serves both as an example and as a unit test.`,
}

type script struct {
	env.Args
	env.Vars
	shell.Interface
}

func newScript() *script {
	return &script{
		Args: env.SystemArgs(),
		Vars: *env.NewVars(),
		Interface: &shell.Shell{
			Stdout: os.Stdout,
			Stderr: os.Stderr,
		},
	}
}

// FIRST is the user's first name
func (sh *script) FIRST() string {
	return sh.Get("FIRST", "Jake")
}

// LAST is the user's last name
func (sh *script) LAST() string {
	return sh.Get("LAST", "Teton-Landis")
}

// NAME is the user's full name, computed from FIRST and LAST.
func (sh *script) NAME() string {
	return sh.GetMemo("NAME", func() string { return sh.FIRST() + " " + sh.LAST() })
}

func (sh *script) somePrivateAccessor() string {
	return "blarg"
}

// Greet shows a greeting to the user.
// Optional: NAME
func (sh *script) Greet() string {
	sh.Runf("echo Hello dearest %s: %s", sh.NAME(), sh.somePrivateAccessor())
}

func main() {
	sh := newScript()
	fmt.Println(shell.Escape("; exit 1"))
	fmt.Println(shell.ScriptPrintf("%s", "; exit 1"))
	sh.Runf("echo first: %s, last: %s, name: %s", sh.FIRST(), sh.LAST(), sh.NAME())
}
