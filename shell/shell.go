package shell

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// DefaultShell is the default shell used for new Shell instances to run scripts.
var DefaultShell = []string{"bash", "-c"}

// Shell provides several ways to execute shell scripts. It allows configuring
// a default Stdout and Stderr for executed scripts. Most Shell methods will
// either ignore or return if the script results in an exec.ExitError, and will
// panic on other errors.
//
// A Shell should not be shared between goroutines - it's simple enough to use
// sh.Copy() to create a new one if needed.
type Shell struct {
	// Eg, []string{"bash", "-c"}
	DefaultArgs []string
	// If set, Stdout will be connected to any executed script's Stdout if possible.
	Stdout io.Writer
	// If set, Stderr will be connected to any executed script's Stderr if possible.
	Stderr io.Writer
	// If true, methods will log unexpected errors instead of panicing.
	IgnoreUnexpectedErrors bool
	// If true, methods will not chomp the last newline of a command's stdout or stderr.
	PreserveTrailingNewline bool
	// Will be added to any commands if not nil
	ctx context.Context
	// Sometimes useful to reference the status of Succeeds or Cmd invocations
	lastError *exec.ExitError
	// Sometimes useful to panic only for one call.
	panicOnNextError bool
}

// WithContext creates a new shell with the given context.
func WithContext(ctx context.Context) *Shell {
	return &Shell{ctx: ctx}
}

// Cmd returns an exec.Cmd that executes the given script. If one of the other
// methods does not suite your purpose, you can implement a different
// abstraction using this command.
//
// The returned command does not have Stdout or Stderr assigned, as some Cmd
// methods require nil Stdout or Stderr.
//
//   cmd := shell.Cmd(`echo 'hello world'`)
//   output, err := cmd.Output()
func (sh *Shell) Cmd(script string) *exec.Cmd {
	if len(sh.DefaultArgs) == 0 {
		sh.DefaultArgs = DefaultShell
	}
	rest := append(sh.DefaultArgs[1:], script)
	if sh.ctx != nil {
		return exec.CommandContext(sh.ctx, sh.DefaultArgs[0], rest...)
	}
	return exec.Command(sh.DefaultArgs[0], rest...)
}

// Ways to run a script:

// Out captures the Stdout of a script and returns it as a string, minus the
// last trailing newline. This is analagous to `$(...)` in Bash. If an error
// occurs, it will be printed to the default Stderr.
func (sh *Shell) Out(script string) string {
	cmd := sh.Cmd(script)
	cmd.Stderr = sh.Stderr
	out, err := cmd.Output()
	sh.onError(err)
	return sh.trim(out)
}

// OutStatus captures the Stdout of a script and returns it as a string, minus
// the last trailing newline. If an error occurs or the command exits non-zero,
// a non-nil error is returned.
func (sh *Shell) OutStatus(script string) (string, error) {
	cmd := sh.Cmd(script)
	cmd.Stderr = sh.Stderr
	out, err := cmd.Output()
	sh.onError(err)
	return sh.trim(out), err
}

// OutErrStatus captures the Stdout and Stderr of a script and returns each as
// a string, minus the last trailing newline. If an error occurs or the command
// exits non-zero, a non-nil error is returned.
func (sh *Shell) OutErrStatus(script string) (string, string, error) {
	var stderr bytes.Buffer
	cmd := sh.Cmd(script)
	cmd.Stderr = &stderr
	out, status := cmd.Output()
	sh.onError(status)
	return sh.trim(out), sh.trim(stderr.Bytes()), status
}

// Run runs the given script to completion.
func (sh *Shell) Run(script string) error {
	cmd := sh.Cmd(script)
	cmd.Stdout = sh.Stdout
	cmd.Stderr = sh.Stderr
	err := cmd.Run()
	sh.onError(err)
	return err
}

// Succeeds runs the script and returns true if the script exited 0, or false
// otherise.
func (sh *Shell) Succeeds(script string) bool {
	err := sh.Run(script)
	return err == nil
}

// LastError returns the last ExitError of script run. This can be useful for
// checking the exit code of Succeeds or Out calls. Note that if you share a
// Shell across several goroutines, the LastError may change unexpectedly.
func (sh *Shell) LastError() *exec.ExitError {
	return sh.lastError
}

// Must configures the shell to panic if the next script execution exits
// non-zero.
//
//   pid := sh.Must().Out(`cat /var/run/yolo.pid`)
//   sh.Must().Run(`kill -9 `+pid)
func (sh *Shell) Must() *Shell {
	sh.panicOnNextError = true
	return sh
}

func (sh *Shell) onError(err error) {
	defer func() { sh.panicOnNextError = false }()
	// Update last error
	if err == nil {
		sh.lastError = nil
		return
	}
	if status, ok := err.(*exec.ExitError); ok {
		sh.lastError = status
		if sh.panicOnNextError {
			panic(status)
		}
		return
	}

	if sh.IgnoreUnexpectedErrors {
		if sh.Stderr != nil {
			fmt.Fprintln(sh.Stderr, err)
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		return
	}
	panic(err)
}

func (sh *Shell) trim(data []byte) string {
	if sh.PreserveTrailingNewline {
		return string(data)
	}
	return string(bytes.TrimSuffix(data, []byte("\n")))
}
