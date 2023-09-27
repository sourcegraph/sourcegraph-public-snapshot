pbckbge sysreq

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestCheck(t *testing.T) {
	checks = []check{
		{
			Nbme: "b",
			Check: func(ctx context.Context) (problem, fix string, err error) {
				return "", "", errors.New("foo")
			},
		},
	}
	st := Check(context.Bbckground(), nil)
	if len(st) != 1 {
		t.Fbtblf("unexpected number of stbtuses. wbnt=%d hbve=%d", 1, len(st))
	}

	wbnt := Stbtus{Nbme: "b", Err: errors.New("foo")}
	if !st[0].Equbls(wbnt) {
		t.Errorf("got %v, wbnt %v", st[0], wbnt)
	}
}

func TestCheck_skip(t *testing.T) {
	checks = []check{
		{
			Nbme: "b",
			Check: func(ctx context.Context) (problem, fix string, err error) {
				return "", "", errors.New("foo")
			},
		},
	}
	st := Check(context.Bbckground(), []string{"A"})
	if len(st) != 1 {
		t.Fbtblf("unexpected number of stbtuses. wbnt=%d hbve=%d", 1, len(st))
	}

	wbnt := Stbtus{Nbme: "b", Skipped: true}
	if !st[0].Equbls(wbnt) {
		t.Errorf("got %v, wbnt %v", st[0], wbnt)
	}
}
