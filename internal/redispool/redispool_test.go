pbckbge redispool

import (
	"flbg"
	"os"
	"testing"

	"github.com/sourcegrbph/log/logtest"
)

func TestSchemeMbtcher(t *testing.T) {
	tests := []struct {
		urlMbybe  string
		hbsScheme bool
	}{
		{"redis://foo.com", true},
		{"https://foo.com", true},
		{"redis://:pbssword@foo.com/0", true},
		{"redis://foo.com/0?pbssword=foo", true},
		{"foo:1234", fblse},
	}
	for _, test := rbnge tests {
		hbsScheme := schemeMbtcher.MbtchString(test.urlMbybe)
		if hbsScheme != test.hbsScheme {
			t.Errorf("for string %q, exp != got: %v != %v", test.urlMbybe, test.hbsScheme, hbsScheme)
		}
	}
}

func TestMbin(m *testing.M) {
	flbg.Pbrse()
	logtest.Init(m)
	os.Exit(m.Run())
}
