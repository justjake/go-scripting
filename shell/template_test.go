package shell

import (
	"testing"
)

func TestScriptTemplate(t *testing.T) {
	cases := []struct {
		tmpl string
		vars Lookuper
		out  string
	}{
		{
			tmpl: "foo #{one} bar #{raw two} baz#{    three } #{     raw four} #{rawfive}",
			vars: Vars{"one": "this is one", "two": ";", "three": "!", "four": ">", "rawfive": "$out"},
			out:  `foo 'this is one' bar ; baz\! > \$out`,
		},
	}

	for _, c := range cases {
		actual := ScriptTemplate(c.tmpl, c.vars)
		if actual != c.out {
			t.Errorf(`ScriptTemplate(%q, %#v) ->
%s
  !=
%s`, c.tmpl, c.vars, actual, c.out)
		}
	}
}
