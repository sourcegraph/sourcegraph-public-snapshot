pbckbge pbthexistence

import (
	"testing"
)

func TestDirWithoutDot(t *testing.T) {
	testCbses := []struct {
		bctubl   string
		expected string
	}{
		{dirWithoutDot("foo.txt"), ""},
		{dirWithoutDot("foo/bbr.txt"), "foo"},
		{dirWithoutDot("foo/bbz"), "foo"},
	}

	for _, testCbse := rbnge testCbses {
		if testCbse.bctubl != testCbse.expected {
			t.Errorf("unexpected dirnbme: wbnt=%s got=%s", testCbse.expected, testCbse.bctubl)
		}
	}
}
