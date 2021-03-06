package shell

import (
	"fmt"
	"regexp"
)

// Lookuper is an environment that can be used for simple templating with StringTemplate.
// Use the Vars type as a shortcut where you would use a map[string]interface{}
type Lookuper interface {
	// Lookup a variable name. Should return the variable value if found without
	// error, or nil and an error.
	//
	// See Vars.Lookup for an example.
	Lookup(name string) (value interface{}, err error)
}

// Vars for ScriptTemplate
type Vars map[string]interface{}

// Lookup implements Lookuper for Vars
func (vars Vars) Lookup(name string) (interface{}, error) {
	if val, found := vars[name]; found {
		return val, nil
	}

	return nil, fmt.Errorf("%q not in %#v", name, vars)
}

const openDelim = `#{`
const closeDelim = `}`
const spaces = ` *`
const name = `(?P<name>[\w-_]+)`
const raw = `(?P<raw>raw +)?`

var openDelimQ = regexp.QuoteMeta(openDelim)
var closeDelimQ = regexp.QuoteMeta(closeDelim)
var matcher = regexp.MustCompile(openDelimQ + spaces + raw + name + spaces + closeDelimQ)

// ScriptTemplate renders a template of a shell script using the provided
// variables.
//
// Each occurrence of `#{varName}` is replaced with occurrence of with the
// corresponding value from the Vars map, converted to strings with Escape().
// This implies that Raw values will not be escaped.
//
// Occurences of `#{raw varName}` will be converted to strings with ToRaw if
// necessary, but not escaped.
//
// ScriptTemplate panics if a varName is not found in vars.
//
// @StaticCompose.Group("formatters", "%st")
func ScriptTemplate(template string, vars Lookuper) string {
	used := make(map[string]bool)
	return matcher.ReplaceAllStringFunc(template, func(match string) string {
		submatch := matcher.FindStringSubmatch(match)
		raw := submatch[1] != ""
		name := submatch[2]

		val, err := vars.Lookup(name)
		if err != nil {
			panic(fmt.Errorf(`Template contained expansion for variable, but lookup failed: %q: %v`, name, err))
		}

		used[name] = true

		var res string
		if raw {
			res = string(ToRaw(val))
		} else {
			res = string(Escape(val))
		}

		return res
	})
}

func templateTest() {
	script := ScriptTemplate(`#{raw KUBECTL} get pods --namespace=#{NAMESPACE} | grep #{APP} | grep -v Terminating | cut -f 1 -d " " | head -1`, Vars{"KUBECTL": "kubectl"})
	fmt.Println(script)
}
