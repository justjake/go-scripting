package shell

import (
	"fmt"
	"os"
)

// Env is an in-process store for string key-value pairs that fall back to the
// system environment variables. Env also provides memoization helpers, because
// a common task is to either use a user-supplied value from the environment,
// or compute it yourself.
type Env struct {
	// map var to result of calculating that var
	memo map[string]string
	// map var to non-exported, local varialbe for that
	vars map[string]string
}

// NewEnv constructs a new Env.
func NewEnv() *Env {
	return &Env{make(map[string]string), make(map[string]string)}
}

// Get a value from the environment, returning a string. If the result is "",
// panic unless a default value is given. You may pass "" as the default value
// to read values that can be empty strings.
func (env *Env) Get(name string, defaultValue ...string) string {
	var res string
	if val, found := env.vars[name]; found {
		res = val
	} else {
		res = os.Getenv(name)
	}

	if res != "" {
		return res
	}

	if len(defaultValue) != 1 {
		panic(fmt.Errorf("Variable undefined or empty: %s", name))
	}

	return defaultValue[0]
}

// Set a value into the local environment, but don't export it as a system environment
// variable. This supports eg command-line variables in the spirit of `make`.
func (env *Env) Set(name, value string) {
	env.vars[name] = value
}

// IsSet returns true if the given name is a defined, non-empty variable
func (env *Env) IsSet(name string) bool {
	return env.Get(name, "") != ""
}

// Memo calls the given function and remembers its return value, keyed by the
// given name.
func (env *Env) Memo(name string, fn func() string) string {
	if val, found := env.memo[name]; found {
		return val
	}

	res := fn()
	env.memo[name] = res
	return res
}

// GetMemo looks the given name up in the system's environment varialbes and
// returns it if it is defined and non-empty. If it is empty, instead we
// memoize a call to fn.
func (env *Env) GetMemo(name string, fn func() string) string {
	if val := env.Get(name, ""); val != "" {
		return val
	}

	return env.Memo(name, fn)
}
