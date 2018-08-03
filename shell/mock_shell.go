package shell

// An attempt to write a mocked implementation of Shell.
// This is probably the wrong way to go:
// Read https://npf.io/2015/06/testing-exec-command/

import (
	"fmt"
	"os/exec"
)

// MockShell can be substituted for a shell for testing purposes.
//
// Example:
//
//   functionUnderTest := func(sh Interface) error {
//     return sh.Runf("echo %s", "hello world")
//   }
//   sh := &MockShell{}
//   sh.AddMock(MockCall{"echo 'hello world'", 128, "", "bash: command not found: echo"})
//   res := functionUnderTest(sh)
//   if res == nil {
//     panic(fmt.Sprintf("Expected res to be an exit error w/ status 128"))
//   }
type MockShell struct {
	Shell
	Mocks         map[string][]MockCall
	mockProgress  map[string]int
	AllowUnmocked bool
	LoopMocks     bool
}

// AddMock adds a pushes a call to this mock shell for the script.
// Mocks are popped in the order in which the script is run.
func (sh *MockShell) AddMock(call MockCall) *MockShell {
	if sh.Mocks == nil {
		sh.Mocks = make(map[string][]MockCall)
	}

	calls := sh.Mocks[call.Script]
	if len(calls) == 0 {
		calls = []MockCall{call}
	} else {
		calls = append(calls, call)
	}

	sh.Mocks[call.Script] = calls
	return sh
}

// +StaticCompose inside:"formatters"
func (sh *MockShell) Out(script string) string {
	res := sh.popMock(script)
	if res == nil {
		return sh.Shell.Out(script)
	}
	return res.Stdout
}

// +StaticCompose inside:"formatters"
func (sh *MockShell) OutStatus(script string) (string, error) {
	res := sh.popMock(script)
	if res == nil {
		return sh.Shell.OutStatus(script)
	}
	return res.Stdout, res.ExitError()
}

// +StaticCompose inside:"formatters"
func (sh *MockShell) OutErrStatus(script string) (string, string, error) {
	res := sh.popMock(script)
	if res == nil {
		return sh.Shell.OutErrStatus(script)
	}
	return res.Stdout, res.Stderr, res.ExitError()
}

// +StaticCompose inside:"formatters"
func (sh *MockShell) Run(script string) error {
	res := sh.popMock(script)
	if res == nil {
		return sh.Shell.Run(script)
	}
	return res.ExitError()
}

// +StaticCompose inside:"formatters"
func (sh *MockShell) Succeeds(script string) bool {
	res := sh.popMock(script)
	if res == nil {
		return sh.Shell.Succeeds(script)
	}
	return res.ExitError() == nil
}

func (sh *MockShell) popMock(script string) *MockCall {
	mocks, found := sh.Mocks[script]
	if !found || len(mocks) == 0 {
		if sh.AllowUnmocked {
			return nil
		}

		panic(fmt.Errorf("No mocks configured for script: %s", script))
	}

	index := sh.mockProgress[script]
	if sh.LoopMocks {
		index = index % len(mocks)
	}

	mock := mocks[index]
	sh.mockProgress[script] = index + 1
	return &mock
}

// MockCall describes an expected script that will return the mocked version, instead.
type MockCall struct {
	Script     string
	ExitStatus int
	Stdout     string
	Stderr     string
}

// ExitError returns the *exec.ExitError for this mock call's ExitStatus, or
// nil if the ExitStatus is zero.
func (call MockCall) ExitError() *exec.ExitError {
	if call.ExitStatus == 0 {
		return nil
	}

	// uhh this is hard to mock because of os.ProcessState so just build one
	err := exec.Command("sh", "-c", fmt.Sprintf("exit %d", call.ExitStatus)).Run()
	return err.(*exec.ExitError)
}
