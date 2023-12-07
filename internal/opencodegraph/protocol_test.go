package opencodegraph

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestDecodeRequestMessage(t *testing.T) {
	tests := []struct {
		input            string
		wantMethod       string
		wantCapabilities *schema.CapabilitiesParams
		wantAnnotations  *schema.AnnotationsParams
	}{
		{
			input:            `{"method":"capabilities","params":{}}`,
			wantMethod:       "capabilities",
			wantCapabilities: &schema.CapabilitiesParams{},
		},
		{
			input:      `{"method":"annotations","params":{"file":"file:///a","content":"c"}}`,
			wantMethod: "annotations",
			wantAnnotations: &schema.AnnotationsParams{
				File:    "file:///a",
				Content: "c",
			},
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			method, capabilities, annotations, err := DecodeRequestMessage(json.NewDecoder(strings.NewReader(test.input)))
			if err != nil {
				t.Fatal(err)
			}
			if method != test.wantMethod {
				t.Errorf("got method %q, want %q", method, test.wantMethod)
			}
			if !reflect.DeepEqual(capabilities, test.wantCapabilities) {
				t.Errorf("got capabilities %+v, want %+v", capabilities, test.wantCapabilities)
			}
			if !reflect.DeepEqual(annotations, test.wantAnnotations) {
				t.Errorf("got annotations %+v, want %+v", annotations, test.wantAnnotations)
			}
		})
	}
}
