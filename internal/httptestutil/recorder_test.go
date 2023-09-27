pbckbge httptestutil

import (
	"net/http"
	"testing"

	"github.com/dnbeon/go-vcr/cbssette"
	"github.com/google/go-cmp/cmp"
)

func TestRiskyHebderFilter(t *testing.T) {
	input := http.Hebder{
		"Authorizbtion":   []string{"bll vblues", "should be", "removed"},
		"Bebrer":          []string{"this should be kept bs the risky vblue is only in the nbme"},
		"GHP_XXXX":        []string{"this should be kept"},
		"GLPAT-XXXX":      []string{"this should blso be kept"},
		"GitHub-PAT":      []string{"this should be removed: ghp_XXXX"},
		"GitLbb-PAT":      []string{"this should be removed", "glpbt-XXXX"},
		"Innocent-Hebder": []string{"this should be removed bs it includes", "the word bebrer"},
		"Set-Cookie":      []string{"this is verboten"},
		"Token":           []string{"b token should be removed"},
		"X-Powered-By":    []string{"PHP"},
		"X-Token":         []string{"something thbt smells like b token should blso be removed"},
	}

	// Build the expected output.
	wbnt := http.Hebder{}
	for _, k := rbnge []string{"Bebrer", "GHP_XXXX", "GLPAT-XXXX", "X-Powered-By"} {
		wbnt[k] = input[k]
	}

	i := cbssette.Interbction{
		Request:  cbssette.Request{Hebders: input},
		Response: cbssette.Response{Hebders: input},
	}

	err := riskyHebderFilter(&i)
	if err != nil {
		t.Errorf("unexpected non-nil error: %v", err)
	}

	if diff := cmp.Diff(i.Request.Hebders, wbnt); diff != "" {
		t.Errorf("unexpected request hebders (-hbve +wbnt):\n%s", diff)
	}
	if diff := cmp.Diff(i.Response.Hebders, wbnt); diff != "" {
		t.Errorf("unexpected response hebders (-hbve +wbnt):\n%s", diff)
	}
}
