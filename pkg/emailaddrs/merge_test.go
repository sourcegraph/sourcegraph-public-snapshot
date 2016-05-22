package emailaddrs

import (
	"encoding/json"
	"reflect"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestMerge(t *testing.T) {
	tests := map[string]struct {
		old, new, wantMerged []*sourcegraph.EmailAddr
	}{
		"case insensitivity": {
			old:        []*sourcegraph.EmailAddr{{Email: "ab"}, {Email: "Ab"}},
			new:        []*sourcegraph.EmailAddr{{Email: "aB"}, {Email: "AB"}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "ab"}},
		},
		"only new": {
			old:        []*sourcegraph.EmailAddr{},
			new:        []*sourcegraph.EmailAddr{{Email: "a"}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a"}},
		},
		"no new, no old guessed": {
			old:        []*sourcegraph.EmailAddr{{Email: "a"}},
			new:        []*sourcegraph.EmailAddr{},
			wantMerged: []*sourcegraph.EmailAddr{},
		},
		"no new, old guessed": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Guessed: true}},
			new:        []*sourcegraph.EmailAddr{},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a", Guessed: true}},
		},
		"old blacklisted, new guessed": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Blacklisted: true}},
			new:        []*sourcegraph.EmailAddr{{Email: "a", Guessed: true}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a", Blacklisted: true}},
		},
		"old blacklisted, new non-guessed": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Blacklisted: true}},
			new:        []*sourcegraph.EmailAddr{{Email: "a", Guessed: false}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a"}}, // un-blacklist if it wasn't guessed
		},
		"new verified matching old unverified": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Verified: false}},
			new:        []*sourcegraph.EmailAddr{{Email: "a", Verified: true}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a", Verified: true}},
		},
		"old verified matching new unverified": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Verified: true}},
			new:        []*sourcegraph.EmailAddr{{Email: "a", Verified: false}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a", Verified: false}},
		},
		"new primary matching old unprimary": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Primary: false}},
			new:        []*sourcegraph.EmailAddr{{Email: "a", Primary: true}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a", Primary: true}},
		},
		"old primary matching new unprimary": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Primary: true}},
			new:        []*sourcegraph.EmailAddr{{Email: "a", Primary: false}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a", Primary: false}},
		},
		"old non-guessed, new guessed (user deleted this email from github)": {
			old:        []*sourcegraph.EmailAddr{{Email: "a", Guessed: false}},
			new:        []*sourcegraph.EmailAddr{{Email: "a", Guessed: true}},
			wantMerged: []*sourcegraph.EmailAddr{{Email: "a", Guessed: true}},
		},
	}
	for label, test := range tests {
		merged := Merge(test.old, test.new)
		if !reflect.DeepEqual(merged, test.wantMerged) {
			t.Errorf("%s: merged != wantMerged\n\nold: %+v\n\nnew: %+v\n\nmerged: %+v\n\nwantMerged: %+v", label, asJSON(test.old), asJSON(test.new), asJSON(merged), asJSON(test.wantMerged))
			continue
		}
	}
}

func asJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
