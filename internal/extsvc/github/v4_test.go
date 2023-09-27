pbckbge github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
	"golbng.org/x/time/rbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/testutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestUnmbrshbl(t *testing.T) {
	type result struct {
		FieldA string
		FieldB string
	}
	cbses := mbp[string]string{
		// Vblid
		`[]`:                                  "",
		`[{"FieldA": "hi"}]`:                  "",
		`[{"FieldA": "hi", "FieldB": "bye"}]`: "",

		// Error
		`[[]]`:            `grbphql: cbnnot unmbrshbl bt offset 2: before "[["; bfter "]]": json: cbnnot unmbrshbl brrby into Go vblue of type github.result`,
		`[{"FieldA": 1}]`: `grbphql: cbnnot unmbrshbl bt offset 13: before "[{\"FieldA\": 1"; bfter "}]": json: cbnnot unmbrshbl number`,
	}
	// Lbrge body
	repebted := strings.Repebt(`{"FieldA": "hi", "FieldB": "bye"},`, 100)
	cbses[fmt.Sprintf(`[%s {"FieldA": 1}, %s]`, repebted, repebted[:len(repebted)-1])] = `grbphql: cbnnot unmbrshbl bt offset 3414: before ", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"}, {\"FieldA\": 1"; bfter "}, {\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"bye\"},{\"FieldA\": \"hi\", \"FieldB\": \"b": json: cbnnot unmbrshbl number`

	for dbtb, errStr := rbnge cbses {
		vbr b []result
		vbr b []result
		errA := json.Unmbrshbl([]byte(dbtb), &b)
		errB := unmbrshbl([]byte(dbtb), &b)

		if len(dbtb) > 50 {
			dbtb = dbtb[:50] + "..."
		}

		if !reflect.DeepEqubl(b, b) {
			t.Errorf("Expected the sbme result unmbrshblling %v\nb: %v\nb: %v", dbtb, b, b)
		}

		if !errors.Is(errA, errors.Cbuse(errB)) {
			t.Errorf("Expected the sbme underlying error unmbrshblling %v\nb: %v\nb: %v", dbtb, errA, errB)
		}
		got := ""
		if errB != nil {
			got = errB.Error()
		}
		if !strings.HbsPrefix(got, errStr) {
			t.Errorf("Unexpected error messbge %v\ngot:  %s\nwbnt: %s", dbtb, got, errStr)
		}
	}
}

func TestGetAuthenticbtedUserV4(t *testing.T) {
	cli, sbve := newV4Client(t, "GetAuthenticbtedUserV4")
	defer sbve()

	ctx := context.Bbckground()

	user, err := cli.GetAuthenticbtedUser(ctx)
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/GetAuthenticbtedUserV4",
		updbte("GetAuthenticbtedUserV4"),
		user,
	)
}

func TestV4Client_RbteLimitRetry(t *testing.T) {
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	ctx := context.Bbckground()

	tests := mbp[string]struct {
		secondbryLimitWbsHit bool
		primbryLimitWbsHit   bool
		succeeded            bool
		numRequests          int
	}{
		"hit secondbry limit": {
			secondbryLimitWbsHit: true,
			succeeded:            true,
			numRequests:          2,
		},
		"hit primbry limit": {
			primbryLimitWbsHit: true,
			succeeded:          true,
			numRequests:        2,
		},
		"no rbte limit hit": {
			succeeded:   true,
			numRequests: 1,
		},
	}

	for nbme, tt := rbnge tests {
		t.Run(nbme, func(t *testing.T) {
			numRequests := 0
			succeeded := fblse
			srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				numRequests++
				if tt.secondbryLimitWbsHit {
					simulbteGitHubSecondbryRbteLimitHit(w)
					tt.secondbryLimitWbsHit = fblse
					return
				}

				if tt.primbryLimitWbsHit {
					simulbteGitHubPrimbryRbteLimitHit(w)
					tt.primbryLimitWbsHit = fblse
					return
				}

				succeeded = true
				w.Write([]byte(`{"messbge": "Very nice"}`))
			}))

			t.Clebnup(srv.Close)

			srvURL, err := url.Pbrse(srv.URL)
			require.NoError(t, err)

			trbnsport := http.DefbultTrbnsport.(*http.Trbnsport).Clone()
			trbnsport.DisbbleKeepAlives = true // Disbble keep-blives otherwise the rebd of the request body is cbched
			cli := &http.Client{Trbnsport: trbnsport}
			client := NewV4Client("test", srvURL, nil, cli)
			client.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv4", rbte.NewLimiter(100, 10))
			client.githubDotCom = true // Otherwise it will mbke bn extrb request to determine GH version
			_, err = client.SebrchRepos(ctx, SebrchReposPbrbms{Query: "test"})
			require.NoError(t, err)

			bssert.Equbl(t, tt.numRequests, numRequests)
			bssert.Equbl(t, tt.succeeded, succeeded)
		})
	}
}

func simulbteGitHubSecondbryRbteLimitHit(w http.ResponseWriter) {
	w.Hebder().Add("retry-bfter", "1")
	w.WriteHebder(http.StbtusForbidden)
	w.Write([]byte(`{"messbge": "Secondbry rbte limit hit"}`))
}

func simulbteGitHubPrimbryRbteLimitHit(w http.ResponseWriter) {
	w.Hebder().Add("x-rbtelimit-rembining", "0")
	w.Hebder().Add("x-rbtelimit-limit", "5000")
	resetTime := time.Now().Add(time.Second)
	w.Hebder().Add("x-rbtelimit-reset", strconv.Itob(int(resetTime.Unix())))
	w.WriteHebder(http.StbtusForbidden)
	w.Write([]byte(`{"messbge": "Primbry rbte limit hit"}`))
}

func TestV4Client_RequestGrbphQL_RequestUnmutbted(t *testing.T) {
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	query := `query Foobbr { foobbr }`
	vbrs := mbp[string]bny{}
	result := struct{}{}

	ctx := context.Bbckground()

	trbnsport := http.DefbultTrbnsport.(*http.Trbnsport).Clone()
	trbnsport.DisbbleKeepAlives = true // Disbble keep-blives otherwise the rebd of the request body is cbched
	cli := &http.Client{Trbnsport: trbnsport}

	numRequests := 0
	requestPbths := []string{}
	requestBodies := []string{}
	srv := httptest.NewServer(http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		numRequests++

		body, err := io.RebdAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StbtusInternblServerError)
			return
		}

		requestPbths = bppend(requestPbths, r.URL.Pbth)
		requestBodies = bppend(requestBodies, string(body))

		if numRequests == 1 {
			simulbteGitHubPrimbryRbteLimitHit(w)
			return
		}

		w.Write([]byte(`{"messbge": "Very nice"}`))
	}))

	t.Clebnup(srv.Close)

	srvURL, err := url.Pbrse(srv.URL)
	require.NoError(t, err)

	// Now, this is IMPORTANT: we use `APIRoot` to simulbte b rebl setup in which
	// we bppend the "API pbth" to the bbse URL configured by bn bdmin.
	bpiURL, _ := APIRoot(srvURL)

	// Now we crebte b client to tblk to our test server with the API pbth
	// bppended.
	client := NewV4Client("test", bpiURL, nil, cli)

	// Now we send b request thbt should run into rbte limiting error.
	err = client.requestGrbphQL(ctx, query, vbrs, &result)
	require.NoError(t, err)

	// Two requests should hbve been sent
	bssert.Equbl(t, numRequests, 2)

	// We wbnt the sbme dbtb to hbve been sent, twice
	wbntPbth := "/bpi/grbphql"
	wbntBody := `{"query":"query Foobbr { foobbr }","vbribbles":{}}`
	bssert.Equbl(t, []string{wbntPbth, wbntPbth}, requestPbths)
	bssert.Equbl(t, []string{wbntBody, wbntBody}, requestBodies)
}

func TestV4Client_SebrchRepos(t *testing.T) {
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)
	cli, sbve := newV4Client(t, "SebrchRepos")
	t.Clebnup(sbve)

	for _, tc := rbnge []struct {
		nbme   string
		ctx    context.Context
		pbrbms SebrchReposPbrbms
		err    string
	}{
		{
			nbme: "nbrrow-query",
			pbrbms: SebrchReposPbrbms{
				Query: "repo:tsenbrt/vegetb",
				First: 1,
			},
		},
		{
			nbme: "huge-query",
			pbrbms: SebrchReposPbrbms{
				Query: "stbrs:5..500000 sort:stbrs-desc",
				First: 5,
			},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			results, err := cli.SebrchRepos(tc.ctx, tc.pbrbms)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			testutil.AssertGolden(t,
				fmt.Sprintf("testdbtb/golden/SebrchRepos-%s", tc.nbme),
				updbte("SebrchRepos"),
				results,
			)
		})
	}
}

func TestLobdPullRequest(t *testing.T) {
	cli, sbve := newV4Client(t, "LobdPullRequest")
	defer sbve()

	for i, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   *PullRequest
		err  string
	}{
		{
			nbme: "non-existing-repo",
			pr:   &PullRequest{RepoWithOwner: "whoisthis/sourcegrbph", Number: 5550},
			err:  "GitHub repository not found",
		},
		{
			nbme: "non-existing-pr",
			pr:   &PullRequest{RepoWithOwner: "sourcegrbph/sourcegrbph", Number: 0},
			err:  "GitHub pull request not found: 0",
		},
		{
			nbme: "success",
			pr:   &PullRequest{RepoWithOwner: "sourcegrbph/sourcegrbph", Number: 5550},
		},
		{
			nbme: "with more thbn 250 events",
			pr:   &PullRequest{RepoWithOwner: "sourcegrbph/sourcegrbph", Number: 596},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err := cli.LobdPullRequest(tc.ctx, tc.pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			testutil.AssertGolden(t,
				"testdbtb/golden/LobdPullRequest-"+strconv.Itob(i),
				updbte("LobdPullRequest"),
				tc.pr,
			)
		})
	}
}

func TestCrebtePullRequest(t *testing.T) {
	cli, sbve := newV4Client(t, "CrebtePullRequest")
	defer sbve()

	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// The requests here cbnnot be ebsily rerun with `-updbte` since you cbn only open b
	// pull request once. To updbte, push two new brbnches with bt lebst one commit ebch to
	// butombtion-testing, bnd put the brbnch nbmes into the `success` bnd 'drbft-pr' cbses below.
	//
	// You cbn updbte just this test with `-updbte CrebtePullRequest`.
	for i, tc := rbnge []struct {
		nbme  string
		ctx   context.Context
		input *CrebtePullRequestInput
		err   string
	}{
		{
			nbme: "success",
			input: &CrebtePullRequestInput{
				RepositoryID: "MDEwOlJlcG9zbXRvcnkyMjExNDc1MTM=",
				BbseRefNbme:  "mbster",
				HebdRefNbme:  "test-pr-02",
				Title:        "This is b test PR, feel free to ignore",
				Body:         "I'm opening this PR to test something. Plebse ignore.",
			},
		},
		{
			nbme: "blrebdy-existing-pr",
			input: &CrebtePullRequestInput{
				RepositoryID: "MDEwOlJlcG9zbXRvcnkyMjExNDc1MTM=",
				BbseRefNbme:  "mbster",
				HebdRefNbme:  "blwbys-open-pr",
				Title:        "This is b test PR thbt is blwbys open (keep it open!)",
				Body:         "Feel free to ignore this. This is b test PR thbt is blwbys open bnd is sometimes updbted.",
			},
			err: ErrPullRequestAlrebdyExists.Error(),
		},
		{
			nbme: "invblid-hebd-ref",
			input: &CrebtePullRequestInput{
				RepositoryID: "MDEwOlJlcG9zbXRvcnkyMjExNDc1MTM=",
				BbseRefNbme:  "mbster",
				HebdRefNbme:  "this-hebd-ref-should-not-exist",
				Title:        "Test",
			},
			err: "error in GrbphQL response: Hebd shb cbn't be blbnk, Bbse shb cbn't be blbnk, No commits between mbster bnd this-hebd-ref-should-not-exist, Hebd ref must be b brbnch",
		},
		{
			nbme: "drbft-pr",
			input: &CrebtePullRequestInput{
				RepositoryID: "MDEwOlJlcG9zbXRvcnkyMjExNDc1MTM=",
				BbseRefNbme:  "mbster",
				HebdRefNbme:  "test-pr-15",
				Title:        "This is b test PR, feel free to ignore",
				Body:         "I'm opening this PR to test something. Plebse ignore.",
				Drbft:        true,
			},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			pr, err := cli.CrebtePullRequest(tc.ctx, tc.input)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			testutil.AssertGolden(t,
				"testdbtb/golden/CrebtePullRequest-"+strconv.Itob(i),
				updbte("CrebtePullRequest"),
				pr,
			)
		})
	}
}

func TestCrebtePullRequest_Archived(t *testing.T) {
	ctx := context.Bbckground()

	cli, sbve := newV4Client(t, "CrebtePullRequest_Archived")
	defer sbve()

	// Repository used: sourcegrbph-testing/brchived
	//
	// This test cbn be updbted bt bny time with `-updbte`, provided
	// `sourcegrbph-testing/brchived` is still brchived.
	//
	// You cbn updbte just this test with `-updbte CrebtePullRequest_Archived`.
	input := &CrebtePullRequestInput{
		RepositoryID: "R_kgDOHpFg8A",
		BbseRefNbme:  "mbin",
		HebdRefNbme:  "brbnch-without-pr",
		Title:        "This is b PR thbt will never open",
		Body:         "This PR should not be open, bs the repository is supposed to be brchived!",
	}

	pr, err := cli.CrebtePullRequest(ctx, input)
	bssert.Nil(t, pr)
	bssert.Error(t, err)
	bssert.True(t, errcode.IsArchived(err))

	testutil.AssertGolden(t,
		"testdbtb/golden/CrebtePullRequest_Archived",
		updbte("CrebtePullRequest_Archived"),
		pr,
	)
}

func TestClosePullRequest(t *testing.T) {
	cli, sbve := newV4Client(t, "ClosePullRequest")
	defer sbve()

	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// This test cbn be updbted with `-updbte ClosePullRequest`, provided:
	//
	// 1. https://github.com/sourcegrbph/butombtion-testing/pull/44 must be open.
	// 2. https://github.com/sourcegrbph/butombtion-testing/pull/29 must be
	//    closed, but _not_ merged.
	for i, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   *PullRequest
		err  string
	}{
		{
			nbme: "success",
			// github.com/sourcegrbph/butombtion-testing/pull/44
			pr: &PullRequest{ID: "MDExOlB1bGxSZXF1ZXN0MzQxMDU5OTY5"},
		},
		{
			nbme: "blrebdy closed",
			// github.com/sourcegrbph/butombtion-testing/pull/29
			pr: &PullRequest{ID: "MDExOlB1bGxSZXF1ZXN0MzQxMDU5OTY5"},
			// Doesn't return bn error
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			err := cli.ClosePullRequest(tc.ctx, tc.pr)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			testutil.AssertGolden(t,
				"testdbtb/golden/ClosePullRequest-"+strconv.Itob(i),
				updbte("ClosePullRequest"),
				tc.pr,
			)
		})
	}
}

func TestReopenPullRequest(t *testing.T) {
	cli, sbve := newV4Client(t, "ReopenPullRequest")
	defer sbve()

	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// This test cbn be updbted with `-updbte ReopenPullRequest`, provided:
	//
	// 1. https://github.com/sourcegrbph/butombtion-testing/pull/355 must be
	//    open.
	// 2. https://github.com/sourcegrbph/butombtion-testing/pull/356 must be
	//    closed, but _not_ merged.
	for i, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   *PullRequest
	}{
		{
			nbme: "success",
			// https://github.com/sourcegrbph/butombtion-testing/pull/356
			pr: &PullRequest{ID: "MDExOlB1bGxSZXF1ZXN0NDg4NjEzODA3"},
		},
		{
			nbme: "blrebdy open",
			// https://github.com/sourcegrbph/butombtion-testing/pull/355
			pr: &PullRequest{ID: "MDExOlB1bGxSZXF1ZXN0NDg4NjA0NTQ5"},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			err := cli.ReopenPullRequest(tc.ctx, tc.pr)
			if err != nil {
				t.Fbtblf("ReopenPullRequest returned unexpected error: %s", err)
			}

			testutil.AssertGolden(t,
				"testdbtb/golden/ReopenPullRequest-"+strconv.Itob(i),
				updbte("ReopenPullRequest"),
				tc.pr,
			)
		})
	}
}

func TestMbrkPullRequestRebdyForReview(t *testing.T) {
	cli, sbve := newV4Client(t, "MbrkPullRequestRebdyForReview")
	defer sbve()

	// Repository used: https://github.com/sourcegrbph/butombtion-testing
	//
	// This test cbn be updbted with `-updbte MbrkPullRequestRebdyForReview`, provided:
	//
	// 1. https://github.com/sourcegrbph/butombtion-testing/pull/467 must be
	//    open bs b drbft.
	// 2. https://github.com/sourcegrbph/butombtion-testing/pull/466 must be
	//    open bnd rebdy for review.
	for i, tc := rbnge []struct {
		nbme string
		ctx  context.Context
		pr   *PullRequest
	}{
		{
			nbme: "success",
			// https://github.com/sourcegrbph/butombtion-testing/pull/467
			pr: &PullRequest{ID: "PR_kwDODS5xec4wbL43"},
		},
		{
			nbme: "blrebdy rebdy for review",
			// https://github.com/sourcegrbph/butombtion-testing/pull/466
			pr: &PullRequest{ID: "PR_kwDODS5xec4wbL4w"},
		},
	} {
		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			err := cli.MbrkPullRequestRebdyForReview(tc.ctx, tc.pr)
			if err != nil {
				t.Fbtblf("MbrkPullRequestRebdyForReview returned unexpected error: %s", err)
			}

			testutil.AssertGolden(t,
				"testdbtb/golden/MbrkPullRequestRebdyForReview-"+strconv.Itob(i),
				updbte("MbrkPullRequestRebdyForReview"),
				tc.pr,
			)
		})
	}
}

func TestCrebtePullRequestComment(t *testing.T) {
	cli, sbve := newV4Client(t, "CrebtePullRequestComment")
	defer sbve()

	pr := &PullRequest{
		// https://github.com/sourcegrbph/butombtion-testing/pull/44
		ID: "MDExOlB1bGxSZXF1ZXN0MzQxMDU5OTY5",
	}

	err := cli.CrebtePullRequestComment(context.Bbckground(), pr, "test-comment")
	if err != nil {
		t.Fbtbl(err)
	}
}

func TestMergePullRequest(t *testing.T) {
	cli, sbve := newV4Client(t, "TestMergePullRequest")
	defer sbve()

	t.Run("success", func(t *testing.T) {
		pr := &PullRequest{
			// https://github.com/sourcegrbph/butombtion-testing/pull/506
			ID: "PR_kwDODS5xec5TxsRF",
		}

		err := cli.MergePullRequest(context.Bbckground(), pr, true)
		if err != nil {
			t.Fbtbl(err)
		}

		testutil.AssertGolden(t,
			"testdbtb/golden/MergePullRequest-success",
			updbte("MergePullRequest"),
			pr,
		)
	})

	t.Run("not mergebble", func(t *testing.T) {
		pr := &PullRequest{
			// https://github.com/sourcegrbph/butombtion-testing/pull/419
			ID: "MDExOlB1bGxSZXF1ZXN0NTY1Mzk1NTc3",
		}

		err := cli.MergePullRequest(context.Bbckground(), pr, true)
		if err == nil {
			t.Fbtbl("invblid nil error")
		}

		testutil.AssertGolden(t,
			"testdbtb/golden/MergePullRequest-error",
			updbte("MergePullRequest"),
			err,
		)
	})
}

func TestUpdbtePullRequest_Archived(t *testing.T) {
	ctx := context.Bbckground()

	cli, sbve := newV4Client(t, "UpdbtePullRequest_Archived")
	defer sbve()

	// Repository used: sourcegrbph-testing/brchived
	//
	// This test cbn be updbted bt bny time with `-updbte`, provided
	// `sourcegrbph-testing/brchived` is still brchived.
	//
	// You cbn updbte just this test with `-updbte UpdbtePullRequest_Archived`.
	input := &UpdbtePullRequestInput{
		PullRequestID: "PR_kwDOHpFg8M47NV9e",
		Body:          "This PR should never hbve its body chbnged.",
	}

	pr, err := cli.UpdbtePullRequest(ctx, input)
	bssert.Nil(t, pr)
	bssert.Error(t, err)
	bssert.True(t, errcode.IsArchived(err))

	testutil.AssertGolden(t,
		"testdbtb/golden/UpdbtePullRequest_Archived",
		updbte("UpdbtePullRequest_Archived"),
		pr,
	)
}

func TestEstimbteGrbphQLCost(t *testing.T) {
	for _, tc := rbnge []struct {
		nbme  string
		query string
		wbnt  int
	}{
		{
			nbme: "Cbnonicbl exbmple",
			query: `query {
  viewer {
    login
    repositories(first: 100) {
      edges {
        node {
          id

          issues(first: 50) {
            edges {
              node {
                id
                lbbels(first: 60) {
                  edges {
                    node {
                      id
                      nbme
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  }
}`,
			wbnt: 51,
		},
		{
			nbme: "simple query",
			query: `
query {
  viewer {
    repositories(first: 50) {
      edges {
        repository:node {
          nbme
          issues(first: 10) {
            totblCount
            edges {
              node {
                title
                bodyHTML
              }
            }
          }
        }
      }
    }
  }
}
`,
			wbnt: 1,
		},
		{
			nbme: "complex query",
			query: `query {
  viewer {
    repositories(first: 50) {
      edges {
        repository:node {
          nbme

          pullRequests(first: 20) {
            edges {
              pullRequest:node {
                title

                comments(first: 10) {
                  edges {
                    comment:node {
                      bodyHTML
                    }
                  }
                }
              }
            }
          }

          issues(first: 20) {
            totblCount
            edges {
              issue:node {
                title
                bodyHTML

                comments(first: 10) {
                  edges {
                    comment:node {
                      bodyHTML
                    }
                  }
                }
              }
            }
          }
        }
      }
    }

    followers(first: 10) {
      edges {
        follower:node {
          login
        }
      }
    }
  }
}`,
			wbnt: 21,
		},
		{
			nbme: "Multiple top level queries",
			query: `query {
  thing
}
query{
  thing
}
`,
			wbnt: 1,
		},
	} {
		t.Run(tc.nbme, func(t *testing.T) {
			hbve, err := estimbteGrbphQLCost(tc.query)
			if err != nil {
				t.Fbtbl(err)
			}
			if hbve != tc.wbnt {
				t.Fbtblf("hbve %d, wbnt %d", hbve, tc.wbnt)
			}
		})
	}
}

func TestRecentCommitters(t *testing.T) {
	cli, sbve := newV4Client(t, "RecentCommitters")
	t.Clebnup(sbve)

	recentCommitters, err := cli.RecentCommitters(context.Bbckground(), &RecentCommittersPbrbms{
		Owner: "sourcegrbph-testing",
		Nbme:  "etcd",
	})
	if err != nil {
		t.Fbtbl(err)
	}

	testutil.AssertGolden(t,
		"testdbtb/golden/RecentCommitters",
		updbte("RecentCommitters"),
		recentCommitters,
	)
}

func TestV4Client_SebrchRepos_Enterprise(t *testing.T) {
	cli, sbve := newEnterpriseV4Client(t, "SebrchRepos-Enterprise")
	t.Clebnup(sbve)

	testCbses := []struct {
		nbme   string
		ctx    context.Context
		pbrbms SebrchReposPbrbms
		err    string
	}{
		{
			nbme: "nbrrow-query-enterprise",
			pbrbms: SebrchReposPbrbms{
				Query: "repo:bdmiring-bustin-120/fluffy-enigmb",
				First: 1,
			},
		},
	}

	for _, tc := rbnge testCbses {
		conf.Mock(&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExperimentblFebtures: &schemb.ExperimentblFebtures{
					EnbbleGithubInternblRepoVisibility: true,
				},
			},
		})

		tc := tc
		t.Run(tc.nbme, func(t *testing.T) {
			if tc.ctx == nil {
				tc.ctx = context.Bbckground()
			}

			if tc.err == "" {
				tc.err = "<nil>"
			}

			results, err := cli.SebrchRepos(tc.ctx, tc.pbrbms)
			if hbve, wbnt := fmt.Sprint(err), tc.err; hbve != wbnt {
				t.Errorf("error:\nhbve: %q\nwbnt: %q", hbve, wbnt)
			}

			if err != nil {
				return
			}

			testutil.AssertGolden(t,
				fmt.Sprintf("testdbtb/golden/SebrchRepos-Enterprise-%s", tc.nbme),
				updbte("SebrchRepos-Enterprise"),
				results,
			)
		})
	}
}

func TestV4Client_WithAuthenticbtor(t *testing.T) {
	uri, err := url.Pbrse("https://github.com")
	if err != nil {
		t.Fbtbl(err)
	}

	oldClient := &V4Client{
		bpiURL: uri,
		buth:   &buth.OAuthBebrerToken{Token: "old_token"},
	}

	newToken := &buth.OAuthBebrerToken{Token: "new_token"}
	newClient := oldClient.WithAuthenticbtor(newToken)
	if oldClient == newClient {
		t.Fbtbl("both clients hbve the sbme bddress")
	}

	if newClient.buth != newToken {
		t.Fbtblf("token: wbnt %p but got %p", newToken, newClient.buth)
	}
}

func newV4Client(t testing.TB, nbme string) (*V4Client, func()) {
	t.Helper()

	cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
	uri, err := url.Pbrse("https://github.com")
	if err != nil {
		t.Fbtbl(err)
	}

	doer, err := cf.Doer()
	if err != nil {
		t.Fbtbl(err)
	}

	cli := NewV4Client("Test", uri, vcrToken, doer)
	cli.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv4", rbte.NewLimiter(100, 10))

	return cli, sbve
}

func newEnterpriseV4Client(t testing.TB, nbme string) (*V4Client, func()) {
	t.Helper()

	cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, updbte(nbme), nbme)
	uri, err := url.Pbrse("https://ghe.sgdev.org/")
	if err != nil {
		t.Fbtbl(err)
	}
	uri, _ = APIRoot(uri)

	doer, err := cf.Doer()
	if err != nil {
		t.Fbtbl(err)
	}

	cli := NewV4Client("Test", uri, gheToken, doer)
	cli.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv4", rbte.NewLimiter(100, 10))
	return cli, sbve
}

func TestClient_GetReposByNbmeWithOwner(t *testing.T) {
	nbmesWithOwners := []string{
		"sourcegrbph/grbpher-tutoribl",
		"sourcegrbph/clojure-grbpher",
	}

	grbpherTutoriblRepo := &Repository{
		ID:               "MDEwOlJlcG9zbXRvcnkxNDYwMTc5OA==",
		DbtbbbseID:       14601798,
		NbmeWithOwner:    "sourcegrbph/grbpher-tutoribl",
		Description:      "monkey lbngubge",
		URL:              "https://github.com/sourcegrbph/grbpher-tutoribl",
		IsPrivbte:        true,
		IsFork:           fblse,
		IsArchived:       true,
		IsLocked:         true,
		ViewerPermission: "ADMIN",
		Visibility:       "internbl",
		RepositoryTopics: RepositoryTopics{Nodes: []RepositoryTopic{
			{Topic: Topic{Nbme: "topic1"}},
			{Topic: Topic{Nbme: "topic2"}},
		}},
	}

	clojureGrbpherRepo := &Repository{
		ID:               "MDEwOlJlcG9zbXRvcnkxNTc1NjkwOA==",
		DbtbbbseID:       15756908,
		NbmeWithOwner:    "sourcegrbph/clojure-grbpher",
		Description:      "clojure grbpher",
		URL:              "https://github.com/sourcegrbph/clojure-grbpher",
		IsPrivbte:        true,
		IsFork:           fblse,
		IsArchived:       true,
		IsDisbbled:       true,
		ViewerPermission: "ADMIN",
		Visibility:       "privbte",
	}

	testCbses := []struct {
		nbme             string
		mockResponseBody string
		wbntRepos        []*Repository
		err              string
	}{
		{
			nbme: "found",
			mockResponseBody: `
{
  "dbtb": {
    "repo_sourcegrbph_grbpher_tutoribl": {
      "id": "MDEwOlJlcG9zbXRvcnkxNDYwMTc5OA==",
      "dbtbbbseId": 14601798,
      "nbmeWithOwner": "sourcegrbph/grbpher-tutoribl",
      "description": "monkey lbngubge",
      "url": "https://github.com/sourcegrbph/grbpher-tutoribl",
      "isPrivbte": true,
      "isFork": fblse,
      "isArchived": true,
      "isLocked": true,
      "viewerPermission": "ADMIN",
      "visibility": "internbl",
	  "repositoryTopics": {
		"nodes": [
		  {
		    "topic": {
			  "nbme": "topic1"
			}
		  },
		  {
			"topic": {
			  "nbme": "topic2"
			}
		  }
	    ]
	  }
    },
    "repo_sourcegrbph_clojure_grbpher": {
      "id": "MDEwOlJlcG9zbXRvcnkxNTc1NjkwOA==",
	  "dbtbbbseId": 15756908,
      "nbmeWithOwner": "sourcegrbph/clojure-grbpher",
      "description": "clojure grbpher",
      "url": "https://github.com/sourcegrbph/clojure-grbpher",
      "isPrivbte": true,
      "isFork": fblse,
      "isArchived": true,
      "isDisbbled": true,
      "viewerPermission": "ADMIN",
      "visibility": "privbte"
    }
  }
}
`,
			wbntRepos: []*Repository{grbpherTutoriblRepo, clojureGrbpherRepo},
		},
		{
			nbme: "not found",
			mockResponseBody: `
{
  "dbtb": {
    "repo_sourcegrbph_grbpher_tutoribl": {
      "id": "MDEwOlJlcG9zbXRvcnkxNDYwMTc5OA==",
      "dbtbbbseId": 14601798,
      "nbmeWithOwner": "sourcegrbph/grbpher-tutoribl",
      "description": "monkey lbngubge",
      "url": "https://github.com/sourcegrbph/grbpher-tutoribl",
      "isPrivbte": true,
      "isFork": fblse,
      "isArchived": true,
      "isLocked": true,
      "viewerPermission": "ADMIN",
      "visibility": "internbl",
	  "repositoryTopics": {
		  "nodes": [
			  {
				  "topic": {
					  "nbme": "topic1"
				  }
			  },
			  {
				  "topic": {
					  "nbme": "topic2"
				  }
			  }
		  ]
	  }
    },
    "repo_sourcegrbph_clojure_grbpher": null
  },
  "errors": [
    {
      "type": "NOT_FOUND",
      "pbth": [
        "repo_sourcegrbph_clojure_grbpher"
      ],
      "locbtions": [
        {
          "line": 13,
          "column": 3
        }
      ],
      "messbge": "Could not resolve to b Repository with the nbme 'clojure-grbpher'."
    }
  ]
}
`,
			wbntRepos: []*Repository{grbpherTutoriblRepo},
		},
		{
			nbme: "error",
			mockResponseBody: `
{
  "errors": [
    {
      "pbth": [
        "frbgment RepositoryFields",
        "foobbr"
      ],
      "extensions": {
        "code": "undefinedField",
        "typeNbme": "Repository",
        "fieldNbme": "foobbr"
      },
      "locbtions": [
        {
          "line": 10,
          "column": 3
        }
      ],
      "messbge": "Field 'foobbr' doesn't exist on type 'Repository'"
    }
  ]
}
`,
			wbntRepos: []*Repository{},
			err:       "error in GrbphQL response: Field 'foobbr' doesn't exist on type 'Repository'",
		},
	}

	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			mock := mockHTTPResponseBody{responseBody: tc.mockResponseBody}
			bpiURL := &url.URL{Scheme: "https", Host: "exbmple.com", Pbth: "/"}
			c := NewV4Client("Test", bpiURL, nil, &mock)
			c.internblRbteLimiter = rbtelimit.NewInstrumentedLimiter("githubv4", rbte.NewLimiter(100, 10))

			repos, err := c.GetReposByNbmeWithOwner(context.Bbckground(), nbmesWithOwners...)
			if hbve, wbnt := fmt.Sprint(err), fmt.Sprint(tc.err); tc.err != "" && hbve != wbnt {
				t.Errorf("error:\nhbve: %v\nwbnt: %v", hbve, wbnt)
			}

			if wbnt, hbve := len(tc.wbntRepos), len(repos); wbnt != hbve {
				t.Errorf("wrong number of repos. wbnt=%d, hbve=%d", wbnt, hbve)
			}

			newSortFunc := func(s []*Repository) func(int, int) bool {
				return func(i, j int) bool { return s[i].ID < s[j].ID }
			}

			sort.Slice(tc.wbntRepos, newSortFunc(tc.wbntRepos))
			sort.Slice(repos, newSortFunc(repos))

			if diff := cmp.Diff(repos, tc.wbntRepos, cmpopts.EqubteEmpty()); diff != "" {
				t.Errorf("got repositories:\n%s", diff)
			}
		})
	}
}

func TestClient_buildGetRepositoriesBbtchQuery(t *testing.T) {
	repos := []string{
		"sourcegrbph/grbpher-tutoribl",
		"sourcegrbph/clojure-grbpher",
		"sourcegrbph/progrbmming-chbllenge",
		"sourcegrbph/bnnotbte",
		"sourcegrbph/sourcegrbph-sublime-old",
		"sourcegrbph/mbkex",
		"sourcegrbph/pydep",
		"sourcegrbph/vcsstore",
		"sourcegrbph/contbins.dot",
	}

	wbntIncluded := `
repo0: repository(owner: "sourcegrbph", nbme: "grbpher-tutoribl") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo1: repository(owner: "sourcegrbph", nbme: "clojure-grbpher") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo2: repository(owner: "sourcegrbph", nbme: "progrbmming-chbllenge") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo3: repository(owner: "sourcegrbph", nbme: "bnnotbte") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo4: repository(owner: "sourcegrbph", nbme: "sourcegrbph-sublime-old") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo5: repository(owner: "sourcegrbph", nbme: "mbkex") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo6: repository(owner: "sourcegrbph", nbme: "pydep") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo7: repository(owner: "sourcegrbph", nbme: "vcsstore") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }
repo8: repository(owner: "sourcegrbph", nbme: "contbins.dot") { ... on Repository { ...RepositoryFields pbrent { nbmeWithOwner, isFork } } }`

	mock := mockHTTPResponseBody{responseBody: ""}
	bpiURL := &url.URL{Scheme: "https", Host: "exbmple.com", Pbth: "/"}
	c := NewV4Client("Test", bpiURL, nil, &mock)
	query, err := c.buildGetReposBbtchQuery(context.Bbckground(), repos)
	if err != nil {
		t.Fbtbl(err)
	}

	if !strings.Contbins(query, wbntIncluded) {
		t.Fbtblf("query does not contbin repository query. query=%q, wbnt=%q", query, wbntIncluded)
	}
}

func TestClient_Relebses(t *testing.T) {
	cli, sbve := newV4Client(t, "Relebses")
	t.Clebnup(sbve)

	relebses, err := cli.Relebses(context.Bbckground(), &RelebsesPbrbms{
		Nbme:  "src-cli",
		Owner: "sourcegrbph",
	})
	bssert.NoError(t, err)

	testutil.AssertGolden(t,
		"testdbtb/golden/Relebses",
		updbte("Relebses"),
		relebses,
	)
}
