package shell

import (
	"testing"
)

func TestEscape(t *testing.T) {
	cases := []struct {
		in  interface{}
		out string
	}{
		{"foo", "foo"},
		{"foo bar", "'foo bar'"},
		{Raw("foo bar"), "foo bar"},
		{Raw("foo"), "foo"},
		{Raw("foo ; bar"), "foo ; bar"},
	}

	for _, c := range cases {
		actual := Escape(c.in)
		if actual != Raw(c.out) {
			t.Errorf("Escape(%q) -> %q != %q", c.in, actual, c.out)
		}
	}
}

func TestScriptPrint(t *testing.T) {
	cases := []struct {
		in  []interface{}
		out string
	}{
		{[]interface{}{"foo", "bar"}, "foobar"},
		{[]interface{}{"foo", "bar baz"}, "foo'bar baz'"},
		{[]interface{}{"foo", Raw("bar baz"), "quux"}, "foobar bazquux"},
	}

	for _, c := range cases {
		actual := ScriptPrint(c.in...)
		if actual != c.out {
			t.Errorf("ScriptPrint(%#v...) -> %q != %q", c.in, actual, c.out)
		}
	}
}

func TestScriptPrintf(t *testing.T) {
	cases := []struct {
		format string
		vs     []interface{}
		out    string
	}{
		{"%s", []interface{}{"foo bar"}, "'foo bar'"},
		{"foo %s bar", []interface{}{"$first $last"}, "foo '$first $last' bar"},
		{"foo %s bar", []interface{}{Raw("$first $last")}, "foo $first $last bar"},
	}
	for _, c := range cases {
		actual := ScriptPrintf(c.format, c.vs...)
		if actual != c.out {
			t.Errorf("ScriptPrintf(%#v, %#v...) -> %#v != %#v", c.format, c.vs, actual, c.out)
		}
	}
}
