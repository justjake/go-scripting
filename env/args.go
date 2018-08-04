package env

import (
	"os"
	"path/filepath"
)

// Args provides accessors for an array of strings
type Args []string

// SystemArgs returns an Args instance populated with a copy of os.Args
func SystemArgs() Args {
	copied := make([]string, 0, len(os.Args))
	copy(copied, os.Args)
	return Args(copied)
}

// ProcessName returns the basename of the process command line of args
func (x Args) ProcessName() string {
	return filepath.Base(x[0])
}

// Argv returns just the passed arguments
func (x Args) Argv() []string {
	return x[1:]
}
