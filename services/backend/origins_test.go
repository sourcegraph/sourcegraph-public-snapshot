package backend

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
)

func TestCheckValidOriginAndSetDefaultURL(t *testing.T) {
	tests := []struct {
		origin  *sourcegraph.Origin
		want    *sourcegraph.Origin
		wantErr error
	}{
		{nil, nil, nil},
		{origin: &sourcegraph.Origin{Service: 12345}, wantErr: errors.New("Origin.Service value is not recognized")},
		{
			origin:  &sourcegraph.Origin{Service: sourcegraph.Origin_GitHub, ID: ""},
			wantErr: errors.New("Origin.ID must be set"),
		},
		{
			origin: &sourcegraph.Origin{Service: sourcegraph.Origin_GitHub, ID: "1"},
			want:   &sourcegraph.Origin{Service: sourcegraph.Origin_GitHub, ID: "1", APIBaseURL: "https://api.github.com"},
		},
		{
			origin: &sourcegraph.Origin{Service: sourcegraph.Origin_GitHub, ID: "1", APIBaseURL: "https://api.github.com"},
			want:   &sourcegraph.Origin{Service: sourcegraph.Origin_GitHub, ID: "1", APIBaseURL: "https://api.github.com"},
		},
		{
			origin:  &sourcegraph.Origin{Service: sourcegraph.Origin_GitHub, ID: "1", APIBaseURL: "https://invalid.example.com"},
			wantErr: errors.New("Origin.APIBaseURL value of"),
		},
	}
	for _, test := range tests {
		err := checkValidOriginAndSetDefaultURL(test.origin)
		if (err == nil) != (test.wantErr == nil) {
			t.Errorf("got err %v, want %v", err, test.wantErr)
			continue
		}
		if err != nil {
			if !strings.Contains(err.Error(), test.wantErr.Error()) {
				t.Errorf("got err %q, want it to contain %q", err, test.wantErr)
			}
			continue
		}
		if !reflect.DeepEqual(test.origin, test.want) {
			t.Errorf("got origin %#v, want %#v", test.origin, test.want)
		}
	}
}
