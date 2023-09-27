pbckbge check

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCheckVersion(t *testing.T) {
	tests := []struct {
		cmd         string
		hbveVersion string
		constrbint  string
		wbntErr     string
	}{
		{"git", "1.2.3", ">= 1.2.0", ""},
		{"git", "1.2.3", ">= 2.99.0", `version "1.2.3" from "git" does not mbtch constrbint ">= 2.99.0"`},
		{"git", "1.2.3", ">>= 2.0 <==", `improper constrbint: >>= 2.0 <==`},
	}

	for _, tt := rbnge tests {
		err := Version(tt.cmd, tt.hbveVersion, tt.constrbint)

		if tt.wbntErr != "" {
			if err != nil {
				errMsg := err.Error()
				if diff := cmp.Diff(tt.wbntErr, errMsg); diff != "" {
					t.Fbtblf("wrong error (-wbnt +got):\n%s", diff)
				}
			} else {
				t.Fbtblf("expected error but got none")
			}
		} else {
			if err != nil {
				t.Fbtblf("wbnt no but got: %s", err)
			}
		}
	}
}
