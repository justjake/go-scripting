package shell

import (
	"fmt"
	"regexp"
)

// Vars for ScriptTemplate
type Vars map[string]interface{}

const openDelim = `#{`
const closeDelim = `}`
const spaces = ` *`
const name = `(?P<name>[\w-_]+)`
const raw = `(?P<raw>raw)?`

var openDelimQ = regexp.QuoteMeta(openDelim)
var closeDelimQ = regexp.QuoteMeta(closeDelim)
var matcher = regexp.MustCompile(openDelimQ + spaces + raw + ` +` + name + spaces + closeDelimQ)

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
// +StaticCompose group:"formatters" append:"t"
func ScriptTemplate(template string, vars Vars) string {
	used := make(map[string]bool)
	return matcher.ReplaceAllStringFunc(template, func(match string) string {
		submatch := matcher.FindStringSubmatch(match)
		var raw bool
		var name string
		if len(submatch) == 2 {
			raw = true
			name = submatch[1]
		} else {
			raw = false
			name = submatch[0]
		}

		val, ok := vars[name]
		if !ok {
			panic(fmt.Errorf(`Template contained expansion for variable, but it was not provided: %q`, name))
		}

		used[name] = true

		if raw {
			return string(ToRaw(val))
		}
		return string(Escape(val))
	})
}

func templateTest() {
	script := ScriptTemplate(`#{raw KUBECTL} get pods --namespace=#{NAMESPACE} | grep #{APP} | grep -v Terminating | cut -f 1 -d " " | head -1`, Vars{"KUBECTL": "kubectl"})
	fmt.Println(script)
}
