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

func TestScriptTemplatePanics(t *testing.T) {
	defer func() {
		expectedError := "Template contained expansion for variable, but lookup failed: \"notokay\": \"notokay\" not in shell.Vars{\"ok\":1}"
		err := recover().(error)
		if err.Error() != expectedError {
			t.Errorf("%q != %q", err.Error(), expectedError)
		}
	}()

	ScriptTemplate("foo #{ok} bar #{notokay}", Vars{"ok": 1})
	t.Errorf("ScriptTemplate should panic")
}
