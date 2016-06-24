package sourcegraph

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestReposCreateOp_JSON(t *testing.T) {
	tests := []*ReposCreateOp{
		{Op: &ReposCreateOp_New{
			New: &ReposCreateOp_NewRepo{
				URI: "r",
			},
		}},
		{Op: &ReposCreateOp_FromGitHubID{FromGitHubID: 123}},
		{Op: &ReposCreateOp_Origin{Origin: &Origin{ID: "123", Service: Origin_GitHub}}},
	}
	for _, test := range tests {
		data, err := json.Marshal(test)
		if err != nil {
			t.Errorf("%v: Marshal: %s", test, err)
			continue
		}
		var got *ReposCreateOp
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("%s: Unmarshal: %s", data, err)
			continue
		}
		if !reflect.DeepEqual(test, got) {
			t.Errorf("%v != %v", test, got)
		}
	}
}

func TestRepoRevSpec_IsAbs(t *testing.T) {
	tests := map[string]bool{
		"":  false,
		"a": false,
		strings.Repeat("x", 40):                    false,
		"0123456789abcdef0123456789abcdef01234567": true,
	}
	for id, want := range tests {
		if got := (RepoRevSpec{CommitID: id}).IsAbs(); got != want {
			t.Errorf("%s: got %v, want %v", id, got, want)
		}
	}
}
