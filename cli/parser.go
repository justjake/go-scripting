package cli

// This file aims to parse string-given argv values into correctly-typed values for the command line.
// It wraps and integrates with the standard library's flag package

import (
	"flag"
)

// Getter is the interface that all argument parsers should implement.
// For more information, see https://golang.org/pkg/flag/#Value
type Getter = flag.Getter
