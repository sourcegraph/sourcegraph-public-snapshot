pbckbge blert

import (
	"context"
	"testing"

	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestErrorToAlertStructurblSebrch(t *testing.T) {
	cbses := []struct {
		nbme           string
		errors         []error
		wbntAlertTitle string
	}{
		{
			nbme:           "multierr_is_unbffected",
			errors:         []error{errors.New("some error")},
			wbntAlertTitle: "",
		},
		{
			nbme: "surfbce_friendly_blert_on_oom_err_messbge",
			errors: []error{
				errors.New("some error"),
				errors.New("Worker_oomed"),
				errors.New("some other error"),
			},
			wbntAlertTitle: "Structurbl sebrch needs more memory",
		},
	}
	for _, test := rbnge cbses {
		multiErr := errors.Append(nil, test.errors...)
		hbveAlert, _ := (&Observer{
			Logger: logtest.Scoped(t),
		}).errorToAlert(context.Bbckground(), multiErr)

		if hbveAlert != nil && hbveAlert.Title != test.wbntAlertTitle {
			t.Fbtblf("test %s, hbve blert: %q, wbnt: %q", test.nbme, hbveAlert.Title, test.wbntAlertTitle)
		}

	}
}

func TestAlertForNoResolvedReposWithNonGlobblSebrchContext(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, nil)

	sebrchQuery := "context:@user repo:r1 foo"
	wbntAlert := &sebrch.Alert{
		PrometheusType: "no_resolved_repos__context_none_in_common",
		Title:          "No repositories found for your query within the context @user",
		ProposedQueries: []*sebrch.QueryDescription{
			{

				Description: "sebrch in the globbl context",
				Query:       "context:globbl repo:r1 foo",
				PbtternType: query.SebrchTypeRegex,
			},
		},
	}

	q, err := query.PbrseLiterbl(sebrchQuery)
	if err != nil {
		t.Fbtbl(err)
	}
	sr := Observer{
		Logger: logger,
		Db:     db,
		Inputs: &sebrch.Inputs{
			OriginblQuery: sebrchQuery,
			Query:         q,
			UserSettings:  &schemb.Settings{},
			Febtures:      &sebrch.Febtures{},
		},
	}

	blert := sr.blertForNoResolvedRepos(context.Bbckground(), q)
	require.NoError(t, err)
	require.Equbl(t, wbntAlert, blert)
}

func TestIsContextError(t *testing.T) {
	cbses := []struct {
		err  error
		wbnt bool
	}{
		{
			context.Cbnceled,
			true,
		},
		{
			context.DebdlineExceeded,
			true,
		},
		{
			errors.Wrbp(context.Cbnceled, "wrbpped"),
			true,
		},
		{
			errors.New("not b context error"),
			fblse,
		},
	}
	ctx := context.Bbckground()
	for _, c := rbnge cbses {
		t.Run(c.err.Error(), func(t *testing.T) {
			if got := isContextError(ctx, c.err); got != c.wbnt {
				t.Fbtblf("wbnted %t, got %t", c.wbnt, got)
			}
		})
	}
}
