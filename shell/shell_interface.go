// Created by interfacer; DO NOT EDIT

package shell

import (
	"os/exec"
)

// Interface is an interface generated for Shell.
type Interface interface {
	Cmd(string) *exec.Cmd
	Cmdf(string, ...interface{}) *exec.Cmd
	Cmdp(...interface{}) *exec.Cmd
	Cmdt(string, Lookuper) *exec.Cmd
	LastError() *exec.ExitError
	Must() *Shell
	Out(string) string
	OutErrStatus(string) (string, string, error)
	OutErrStatusf(string, ...interface{}) (string, string, error)
	OutErrStatusp(...interface{}) (string, string, error)
	OutErrStatust(string, Lookuper) (string, string, error)
	OutStatus(string) (string, error)
	OutStatusf(string, ...interface{}) (string, error)
	OutStatusp(...interface{}) (string, error)
	OutStatust(string, Lookuper) (string, error)
	Outf(string, ...interface{}) string
	Outp(...interface{}) string
	Outt(string, Lookuper) string
	Run(string) error
	Runf(string, ...interface{}) error
	Runp(...interface{}) error
	Runt(string, Lookuper) error
	Succeeds(string) bool
	Succeedsf(string, ...interface{}) bool
	Succeedsp(...interface{}) bool
	Succeedst(string, Lookuper) bool
}

// ensure compatible w/ Shell
var mockShellInterface Interface = &MockShell{}
var shellInterface Interface = &Shell{}
