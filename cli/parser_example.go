package cli

// This file aims to parse string-given argv values into correctly-typed values for the command line.
// It wraps and integrates with the standard library's flag package

import (
	"flag"
	"fmt"
)

// Getter is the interface that all argument parsers should implement.
// For more information, see https://golang.org/pkg/flag/#Value
type Getter = flag.Getter

// Thing is an example type.
//
// CLI will make all public fields of this type available as options, with the
// following behavior:
//
//   - If a field x.Foo exists, but a method x.FOO() does not exist, CLI will generate
//     x.FOO().
//   - CLI will fail during generation if it notices reads from a field x.Foo
//     in a method containing lowercase letters, because it is better to prefer
//     the method x.FOO() to require that option inside a command body.
//
// @CLI()
type Thing struct {
	// First name
	First string
	// Last name
	Last string
	// Full name, including first and last
	Name string
}

// Add is an example of a pure function with required arguments.
// It returns a + b.
// @Tags("pure", "args")
func (ex *Thing) Add(a, b int) int {
	return a + b
}

// Send is an example of an errorful function with required arguments.
// @Tags("error", "args")
func (ex *Thing) Send(addr string) (string, error) {
	return "", fmt.Errorf("Send failed to address %q", addr)
}

// Greet is an example of a effectful function with optional arguments.
// @Tags("options")
// @Optional("FIRST", "LAST", "NAME")
func (ex *Thing) Greet() {
	fmt.Printf("Hello, %s! Great to see you.\n", ex.NAME())
}

// NAME is the first and last name
func (ex *Thing) NAME() string {
	if ex.Name != "" {
		return ex.Name
	}

	if ex.First != "" && ex.Last != "" {
		return ex.First + " " + ex.Last
	}

	panic("invalid name")
}
