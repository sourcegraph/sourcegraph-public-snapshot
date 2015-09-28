package makex

import (
	"strings"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		rules    []Rule
		makefile string
	}{
		{
			rules: []Rule{
				&BasicRule{
					"myTarget",
					[]string{"myPrereq0", "myPrereq1"},
					[]string{"foo bar"},
				},
			},
			makefile: `
myTarget: myPrereq0 myPrereq1
	foo bar
`,
		},
	}
	for _, test := range tests {
		makefile, err := Marshal(&Makefile{test.rules})
		if err != nil {
			t.Error(err)
			continue
		}
		if got, want := string(makefile), strings.TrimPrefix(test.makefile, "\n"); got != want {
			t.Errorf("bad Makefile\n=========== got Makefile\n%s\n\n=========== want Makefile\n%s", got, want)
		}
	}
}

func TestExpandAutoVars(t *testing.T) {
	tests := []struct {
		rule  Rule
		input string
		want  string
	}{
		{
			rule: &BasicRule{
				"myTarget",
				[]string{"myPrereq0", "myPrereq1"},
				[]string{"foo bar"},
			},
			input: "$@ : $^ : $<",
			want:  "myTarget : myPrereq0 myPrereq1 : myPrereq0",
		},
		{
			rule:  &BasicRule{PrereqFiles: []string{}},
			input: "$<",
			want:  "",
		},
	}
	for _, test := range tests {
		got := ExpandAutoVars(test.rule, test.input)
		if got != test.want {
			t.Errorf("got %q, want %q", got, test.want)
		}
	}
}
