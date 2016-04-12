package sourcegraph

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestRepoResolution_JSON(t *testing.T) {
	tests := []*RepoResolution{
		{Result: &RepoResolution_Repo{
			Repo: &RepoSpec{URI: "r"},
		}},
		{Result: &RepoResolution_RemoteRepo{
			RemoteRepo: &RemoteRepo{GitHubID: 123},
		}},
	}
	for _, test := range tests {
		data, err := json.Marshal(test)
		if err != nil {
			t.Errorf("%v: Marshal: %s", test, err)
			continue
		}
		var got *RepoResolution
		if err := json.Unmarshal(data, &got); err != nil {
			t.Errorf("%s: Unmarshal: %s", data, err)
			continue
		}
		if !reflect.DeepEqual(test, got) {
			t.Errorf("%v != %v", test, got)
		}
	}
}

func TestReposCreateOp_JSON(t *testing.T) {
	tests := []*ReposCreateOp{
		{Op: &ReposCreateOp_New{
			New: &ReposCreateOp_NewRepo{
				URI: "r",
			},
		}},
		{Op: &ReposCreateOp_FromGitHubID{FromGitHubID: 123}},
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
