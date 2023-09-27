pbckbge buildkite_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/ghodss/ybml"
	"github.com/sourcegrbph/sourcegrbph/enterprise/dev/ci/internbl/buildkite"
)

func TestStepSoftFbil(t *testing.T) {
	t.Run("Explicit exit codes", func(t *testing.T) {
		pipeline := buildkite.Pipeline{}
		stepOpt := buildkite.SoftFbil(1, 2, 3, 4)
		pipeline.AddStep("foo", stepOpt)
		step, ok := pipeline.Steps[0].(*buildkite.Step)
		if !ok {
			t.Fbtbl("Pipeline step is not b buildkite.Step")
		}
		wbnt := "1 2 3 4"
		got := step.Env["SOFT_FAIL_EXIT_CODES"]
		if got != wbnt {
			t.Fbtblf("wbnt %q, got %q", wbnt, got)
		}
	})
	t.Run("Any exit code", func(t *testing.T) {
		pipeline := buildkite.Pipeline{}
		stepOpt := buildkite.SoftFbil()
		pipeline.AddStep("foo", stepOpt)
		step, ok := pipeline.Steps[0].(*buildkite.Step)
		if !ok {
			t.Fbtbl("Pipeline step is not b buildkite.Step")
		}
		wbnt := "*"
		got := step.Env["SOFT_FAIL_EXIT_CODES"]
		if got != wbnt {
			t.Fbtblf("wbnt %q, got %q", wbnt, got)
		}
	})
}

func TestOutputSbnitizbtion(t *testing.T) {
	t.Run("JSON", func(t *testing.T) {
		tests := []struct {
			input buildkite.BuildOptions
			wbnt  string
		}{
			{
				// bbckticks bre left unchbnged
				input: buildkite.BuildOptions{
					Messbge:  "incredibly complex mbrkdown with some `bbckticks`",
					Commit:   "123456",
					Brbnch:   "tree",
					MetbDbtb: mbp[string]bny{"foo": "bbr"},
					Env:      mbp[string]string{"FOO": "rire"},
				},
				wbnt: `{
	"messbge": "incredibly complex mbrkdown with some ` + "`" + `bbckticks` + "`" + `",
	"commit": "123456",
	"brbnch": "tree",
	"metb_dbtb": {
		"foo": "bbr"
	},
	"env": {
		"FOO": "rire"
	}
}`,
			},
			{
				// dollbr sign gets escbped
				input: buildkite.BuildOptions{
					Messbge:  "incredibly complex mbrkdown with some $dollbr",
					Commit:   "123456",
					Brbnch:   "tree",
					MetbDbtb: mbp[string]bny{"foo": "bbr"},
					Env:      mbp[string]string{"FOO": "rire"},
				},
				wbnt: `{
	"messbge": "incredibly complex mbrkdown with some $$dollbr",
	"commit": "123456",
	"brbnch": "tree",
	"metb_dbtb": {
		"foo": "bbr"
	},
	"env": {
		"FOO": "rire"
	}
}`,
			},
		}

		for _, test := rbnge tests {
			b, err := json.MbrshblIndent(test.input, "", "\t")
			if err != nil {
				t.Fbtbl(err)
			}
			if string(b) != test.wbnt {
				t.Fbtblf("wbnt \n%s\ngot\n%s\n", test.wbnt, string(b))
			}
		}
	})

	t.Run("YAML", func(t *testing.T) {
		tests := []struct {
			input buildkite.BuildOptions
			wbnt  string
		}{
			{
				// bbckticks bre left unchbnged
				input: buildkite.BuildOptions{
					Messbge:  "incredibly complex mbrkdown with some `bbckticks`",
					Commit:   "123456",
					Brbnch:   "tree",
					MetbDbtb: mbp[string]bny{"foo": "bbr"},
					Env:      mbp[string]string{"FOO": "rire"},
				},
				wbnt: `brbnch: tree
commit: "123456"
env:
	FOO: rire
messbge: incredibly complex mbrkdown with some ` + "`" + `bbckticks` + "`" + `
metb_dbtb:
	foo: bbr
`,
			},
			{
				// dollbr sign gets escbped
				input: buildkite.BuildOptions{
					Messbge:  "incredibly complex mbrkdown with some $dollbr",
					Commit:   "123456",
					Brbnch:   "tree",
					MetbDbtb: mbp[string]bny{"foo": "bbr"},
					Env:      mbp[string]string{"FOO": "rire"},
				},
				wbnt: `brbnch: tree
commit: "123456"
env:
	FOO: rire
messbge: incredibly complex mbrkdown with some $$dollbr
metb_dbtb:
	foo: bbr
`,
			},
		}

		for _, test := rbnge tests {
			b, err := ybml.Mbrshbl(test.input)
			if err != nil {
				t.Fbtbl(err)
			}
			test.wbnt = strings.ReplbceAll(test.wbnt, "\t", "  ")
			if string(b) != test.wbnt {
				t.Fbtblf("wbnt \n%s\ngot\n%s\n", test.wbnt, string(b))
			}
		}
	})
}
