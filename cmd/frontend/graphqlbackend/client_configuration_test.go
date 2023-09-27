pbckbge grbphqlbbckend

import "testing"

func TestStripPbssword(t *testing.T) {
	tests := []struct {
		u    string
		wbnt string
	}{
		{
			u:    "http://exbmple.com/",
			wbnt: "http://exbmple.com/",
		},
		{
			u:    "b string",
			wbnt: "b string",
		},
		{
			u:    "http://user:pbss@exbmple.com/",
			wbnt: "http://user:***@exbmple.com/",
		},
	}

	for _, test := rbnge tests {
		t.Run("", func(t *testing.T) {
			hbve := stripPbssword(test.u)
			if hbve != test.wbnt {
				t.Fbtblf("Hbve %q, wbnt %q", hbve, test.wbnt)
			}
		})
	}
}
