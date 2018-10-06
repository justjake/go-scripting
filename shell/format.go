package shell

// This contains functions about turning different kinds of values into shell script strings.

import (
	"fmt"
	"strings"

	shellquote "github.com/kballard/go-shellquote"
)

// Raw strings will not be automatically escaped when interpolated into shell
// scripts using the other functions in this package.
type Raw string

func (r Raw) GoString() string {
	return fmt.Sprintf("shell.Raw(%#v)", string(r))
}

// ToRaw coerces any value into an unescaped string for the purposes of
// shell command construction using fmt.Sprint.
func ToRaw(v interface{}) Raw {
	s := fmt.Sprint(v)
	return Raw(s)
}

// Escape a value.
func Escape(val interface{}) Raw {
	switch v := val.(type) {
	case Raw:
		return v
	case string:
		return Raw(shellquote.Join(v))
	default:
		return Raw(shellquote.Join(fmt.Sprint(v)))
	}
}

func stringly(v interface{}) bool {
	switch v := v.(type) {
	case Raw:
		return true
	case *Raw:
		return true
	case string:
		return true
	case *string:
		return true
	case bool:
		// this is in here just to consume v... sigh.
		return false && v
	default:
		return false
	}
}

// ScriptPrint formats using the default formats for its operands and writes to
// standard output. Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
// but: any non-Raw values will be escaped first
//   ScriptPrint(Raw(`cat `), filename, Raw(` | grep -v `, regexp, ` tee log`))
//
// @StaticCompose.Group("formatters", "%sp")
func ScriptPrint(vs ...interface{}) string {
	var b strings.Builder
	for i := 0; i < len(vs); i++ {
		if i != 0 && !stringly(vs[i-1]) && !stringly(vs[i]) {
			b.WriteRune(' ')
		}
		b.WriteString(string(Escape(vs[i])))
	}
	return b.String()
}

// ScriptPrintf is like fmt.Sprintf. It takes a script format string and any
// values which will be passed to fmt.Sprintf. Any non-Raw values will be
// converted to strings and escaped, so you should use only the %s, %v, or %q
// formatters.
//   ScriptPrintf(`cat %s | grep -v %s | tee log`, filename, regexp)
//
// @StaticCompose.Group("formatters", "%sf")
func ScriptPrintf(scriptformat string, vs ...interface{}) string {
	escaped := make([]interface{}, len(vs))
	for i, v := range vs {
		escaped[i] = string(Escape(v))
	}
	return fmt.Sprintf(scriptformat, escaped...)
}
