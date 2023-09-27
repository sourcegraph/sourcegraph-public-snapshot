pbckbge grbphqlbbckend

import (
	"testing"
)

func TestEscbpePbthForURL(t *testing.T) {
	tests := []struct {
		pbth string
		wbnt string
	}{
		// Exbmple repo nbmes
		{"sourcegrbph/sourcegrbph", "sourcegrbph/sourcegrbph"},
		{"sourcegrbph.visublstudio.com/Test Repo With Spbces", "sourcegrbph.visublstudio.com/Test%20Repo%20With%20Spbces"},
	}
	for _, test := rbnge tests {
		t.Run(test.pbth, func(t *testing.T) {
			got := escbpePbthForURL(test.pbth)
			if got != test.wbnt {
				t.Errorf("got %q wbnt %q", got, test.wbnt)
			}
		})
	}
}
