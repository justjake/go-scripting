package env

import (
	"fmt"
	"os"
)

// Vars is an in-process store for string key-value pairs that fall back to the
// system vars.ronment variables - or some other way of looking up names.
//
// Vars also provides memoization helpers, because a common task is to either
// use a user-supplied value from the vars.ronment, or compute it yourself.
type Vars struct {
	// In-process variables, similar to unexported variables in a bash or make
	// script.
	Locals map[string]string
	// If defined, this function is used to look up variables not found in
	// Locals. Otherwise, an Vars will look at the vars.ronment variables.
	LookupParent func(name string) (val string, found bool)
	// Stores memoized values
	memo map[string]string
}

// NewVars constructs a new Vars.
func NewVars() *Vars {
	return &Vars{make(map[string]string), nil, make(map[string]string)}
}

// LookupEnv returns the value for the given variable name and true if the
// variable is defined, or an empty string and false if the variable is not
// defined.
func (vars *Vars) LookupEnv(name string) (val string, found bool) {
	if val, found = vars.Locals[name]; found {
		return
	}

	if vars.LookupParent != nil {
		return vars.LookupParent(name)
	}

	return os.LookupEnv(name)
}

// Get a variable in this vars.ronment, if it is defined and is not an empty
// string. If the variable is empty, panic unless a default value is given - in
// which case, return the default value.
func (vars *Vars) Get(name string, defaultValue ...string) string {
	var res string
	if val, found := vars.LookupEnv(name); found {
		res = val
	}

	if res != "" {
		return res
	}

	if len(defaultValue) != 1 {
		panic(fmt.Errorf("Variable undefined or empty: %s", name))
	}

	return defaultValue[0]
}

// Set a value into the local vars.ronment, but don't export it as a system
// vars.ronment variable. This supports eg command-line variables in the spirit
// of `make`.
func (vars *Vars) Set(name, value string) {
	vars.Locals[name] = value
}

// IsSet returns true if the given name is a defined, non-empty variable
func (vars *Vars) IsSet(name string) bool {
	_, set := vars.LookupEnv(name)
	return set
}

// GetMemo returns the named variable if it is defined, or it calls the compute
// function at once and caches its return value.
func (vars *Vars) GetMemo(name string, compute func() string) string {
	if val, found := vars.LookupEnv(name); found {
		return val
	}
	if val, found := vars.memo[name]; found {
		return val
	}

	val := compute()
	vars.memo[name] = val
	return val
}

// Lookup implements shell.Lookuper
func (vars *Vars) Lookup(name string) (val interface{}, err error) {
	stringVal, found := vars.LookupEnv(name)
	if !found {
		return "", fmt.Errorf("Not defined: %q", name)
	}
	return stringVal, nil
}
