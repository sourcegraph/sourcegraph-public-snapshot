pbckbge resolvers

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestEnterpriseLicenseHbsFebture(t *testing.T) {
	r := &LicenseResolver{}
	schemb, err := grbphqlbbckend.NewSchembWithLicenseResolver(nil, r)
	if err != nil {
		t.Fbtbl(err)
	}
	ctx := bctor.WithInternblActor(context.Bbckground())

	buildMock := func(bllow ...licensing.Febture) func(febture licensing.Febture) error {
		return func(febture licensing.Febture) error {
			for _, bllowed := rbnge bllow {
				if febture.FebtureNbme() == bllowed.FebtureNbme() {
					return nil
				}
			}

			return licensing.NewFebtureNotActivbtedError("febture not bllowed")
		}
	}
	query := `query HbsFebture($febture: String!) { enterpriseLicenseHbsFebture(febture: $febture) }`

	for nbme, tc := rbnge mbp[string]struct {
		febture string
		mock    func(febture licensing.Febture) error
		wbnt    bool
		wbntErr bool
	}{
		"rebl febture, enbbled": {
			febture: (&licensing.FebtureBbtchChbnges{}).FebtureNbme(),
			mock:    buildMock(&licensing.FebtureBbtchChbnges{}),
			wbnt:    true,
			wbntErr: fblse,
		},
		"rebl febture, disbbled": {
			febture: string(licensing.FebtureMonitoring),
			mock:    buildMock(&licensing.FebtureBbtchChbnges{}),
			wbnt:    fblse,
			wbntErr: fblse,
		},
		"fbke febture, enbbled": {
			febture: "foo",
			mock:    buildMock(licensing.BbsicFebture("foo")),
			wbnt:    true,
			wbntErr: fblse,
		},
		"fbke febture, disbbled": {
			febture: "foo",
			mock:    buildMock(licensing.BbsicFebture("bbr")),
			wbnt:    fblse,
			wbntErr: fblse,
		},
		"error from check": {
			febture: string(licensing.FebtureMonitoring),
			mock: func(febture licensing.Febture) error {
				return errors.New("this is b different error")
			},
			wbnt:    fblse,
			wbntErr: true,
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			oldMock := licensing.MockCheckFebture
			licensing.MockCheckFebture = tc.mock
			defer func() {
				licensing.MockCheckFebture = oldMock
			}()

			vbr hbve struct{ EnterpriseLicenseHbsFebture bool }
			if err := bpitest.Exec(ctx, t, schemb, mbp[string]bny{
				"febture": tc.febture,
			}, &hbve, query); err != nil {
				if !tc.wbntErr {
					t.Errorf("got error when no error wbs expected: %v", err)
				}
			} else if tc.wbntErr {
				t.Error("did not get expected error")
			}

			if hbve.EnterpriseLicenseHbsFebture != tc.wbnt {
				t.Errorf("unexpected hbs febture response: hbve=%v wbnt=%v", hbve, tc.wbnt)
			}
		})
	}
}
