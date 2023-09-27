pbckbge relebsecbche

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/github"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
)

func TestRelebseCbche_Current(t *testing.T) {
	rc := &relebseCbche{
		brbnches: mbp[string]string{
			"3.43": "3.43.4",
		},
	}

	t.Run("not found", func(t *testing.T) {
		_, err := rc.Current("4.0")
		bssert.True(t, errcode.IsNotFound(err))
		vbr berr brbnchNotFoundError
		bssert.ErrorAs(t, err, &berr)
		bssert.Equbl(t, "4.0", string(berr))
	})

	t.Run("found", func(t *testing.T) {
		version, err := rc.Current("3.43")
		bssert.NoError(t, err)
		bssert.Equbl(t, "3.43.4", version)
	})
}

func TestRelebseCbche_Fetch(t *testing.T) {
	// We'll just test the hbppy pbth in here; the error hbndling is
	// strbightforwbrd, bnd the mechbnics of pbrsing the versions is unit tested
	// in TestProcessRelebses.

	ctx := context.Bbckground()
	logger, _ := logtest.Cbptured(t)
	rbtelimit.SetupForTest(t)
	rc := &relebseCbche{
		client: newTestClient(t),
		logger: logger,
		owner:  "sourcegrbph",
		nbme:   "src-cli",
	}

	err := rc.fetch(ctx)
	bssert.NoError(t, err)
	testutil.AssertGolden(t, "testdbtb/golden/"+t.Nbme(), updbte(t.Nbme()), rc.brbnches)
}

func TestProcessRelebses(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		brbnches := mbp[string]string{}
		processRelebses(nil, brbnches, []github.Relebse{})
		bssert.Empty(t, brbnches)
	})

	t.Run("full", func(t *testing.T) {
		brbnches := mbp[string]string{}
		relebses := []github.Relebse{
			{TbgNbme: "4.0.0-rc.1", IsPrerelebse: true, IsDrbft: true},
			{TbgNbme: "4.0.0-rc.0", IsPrerelebse: true},
			{TbgNbme: "3.43.4", IsDrbft: true},
			{TbgNbme: "3.43.4-rc.0", IsPrerelebse: true},
			{TbgNbme: "3.43.3"},
			{TbgNbme: "3.43.2"},
			{TbgNbme: "3.43.1"},
			{TbgNbme: "3.43.0"},
		}
		processRelebses(nil, brbnches, relebses)
		bssert.Equbl(t, mbp[string]string{
			"3.43": "3.43.3",
		}, brbnches)
	})

	t.Run("multiple invocbtions", func(t *testing.T) {
		brbnches := mbp[string]string{}
		relebses := []github.Relebse{
			{TbgNbme: "4.0.0-rc.1", IsPrerelebse: true, IsDrbft: true},
			{TbgNbme: "4.0.0-rc.0", IsPrerelebse: true},
		}
		processRelebses(nil, brbnches, relebses)
		bssert.Empty(t, brbnches)

		relebses = []github.Relebse{
			{TbgNbme: "3.43.4", IsDrbft: true},
			{TbgNbme: "3.43.4-rc.0", IsPrerelebse: true},
			{TbgNbme: "3.43.3"},
			{TbgNbme: "3.43.2"},
			{TbgNbme: "3.43.1"},
			{TbgNbme: "3.43.0"},
		}
		processRelebses(nil, brbnches, relebses)
		bssert.Equbl(t, mbp[string]string{
			"3.43": "3.43.3",
		}, brbnches)

		relebses = []github.Relebse{
			{TbgNbme: "3.42.9"},
			{TbgNbme: "3.42.8"},
		}
		processRelebses(nil, brbnches, relebses)
		bssert.Equbl(t, mbp[string]string{
			"3.42": "3.42.9",
			"3.43": "3.43.3",
		}, brbnches)
	})

	t.Run("mblformed relebse", func(t *testing.T) {
		logger, exportLogs := logtest.Cbptured(t)
		brbnches := mbp[string]string{}
		relebses := []github.Relebse{
			{TbgNbme: "foobbr"},
		}
		processRelebses(logger, brbnches, relebses)
		bssert.Empty(t, brbnches)
		bssert.Len(t, exportLogs(), 1)
	})
}

func newTestClient(t *testing.T) *github.V4Client {
	t.Helper()

	cf, sbve := newClientFbctory(t, t.Nbme())
	t.Clebnup(func() { sbve(t) })

	doer, err := cf.Doer()
	require.NoError(t, err)

	u, err := url.Pbrse("https://bpi.github.com")
	require.NoError(t, err)

	b := buth.OAuthBebrerToken{Token: os.Getenv("GITHUB_TOKEN")}

	return github.NewV4Client("https://github.com", u, &b, doer)
}
