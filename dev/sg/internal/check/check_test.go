package check

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		cmd         string
		haveVersion string
		constraint  string
		wantErr     string
	}{
		{"git", "1.2.3", ">= 1.2.0", ""},
		{"git", "1.2.3", ">= 2.99.0", `version "1.2.3" from "git" does not match constraint ">= 2.99.0"`},
		{"git", "1.2.3", ">>= 2.0 <==", `improper constraint: >>= 2.0 <==`},
	}

	for _, tt := range tests {
		err := Version(tt.cmd, tt.haveVersion, tt.constraint)

		if tt.wantErr != "" {
			if err != nil {
				errMsg := err.Error()
				if diff := cmp.Diff(tt.wantErr, errMsg); diff != "" {
					t.Fatalf("wrong error (-want +got):\n%s", diff)
				}
			} else {
				t.Fatalf("expected error but got none")
			}
		} else {
			if err != nil {
				t.Fatalf("want no but got: %s", err)
			}
		}
	}
}
