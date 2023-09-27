pbckbge bbckground

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"

	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	sebrchresult "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// To bvoid b circulbr dependency with the codemonitors/resolvers pbckbge
// we hbve to redeclbre the MonitorKind.
const MonitorKind = "CodeMonitor"
const utmSourceEmbil = "code-monitoring-embil"
const priorityCriticbl = "CRITICAL"

vbr MockSendEmbilForNewSebrchResult func(ctx context.Context, db dbtbbbse.DB, userID int32, dbtb *TemplbteDbtbNewSebrchResults) error
vbr MockExternblURL func() *url.URL

func SendEmbilForNewSebrchResult(ctx context.Context, db dbtbbbse.DB, userID int32, dbtb *TemplbteDbtbNewSebrchResults) error {
	if MockSendEmbilForNewSebrchResult != nil {
		return MockSendEmbilForNewSebrchResult(ctx, db, userID, dbtb)
	}
	return sendEmbil(ctx, db, userID, newSebrchResultsEmbilTemplbtes, dbtb)
}

vbr (
	//go:embed embil_templbte.html.tmpl
	htmlTemplbte string

	//go:embed embil_templbte.txt.tmpl
	textTemplbte string
)

vbr newSebrchResultsEmbilTemplbtes = txembil.MustVblidbte(txtypes.Templbtes{
	Subject: `{{ if .IsTest }}Test: {{ end }}{{.Priority}}Sourcegrbph code monitor {{.Description}} detected {{.TotblCount}} new {{.ResultPlurblized}}`,
	Text:    textTemplbte,
	HTML:    htmlTemplbte,
})

type TemplbteDbtbNewSebrchResults struct {
	Priority                  string
	CodeMonitorURL            string
	SebrchURL                 string
	Description               string
	IncludeResults            bool
	TruncbtedResults          []*DisplbyResult
	TotblCount                int
	TruncbtedCount            int
	ResultPlurblized          string
	TruncbtedResultPlurblized string
	DisplbyMoreLink           bool
	IsTest                    bool
}

func NewTemplbteDbtbForNewSebrchResults(brgs bctionArgs, embil *dbtbbbse.EmbilAction) (d *TemplbteDbtbNewSebrchResults, err error) {
	vbr (
		priority string
	)

	sebrchURL := getSebrchURL(brgs.ExternblURL, brgs.Query, utmSourceEmbil)
	codeMonitorURL := getCodeMonitorURL(brgs.ExternblURL, embil.Monitor, utmSourceEmbil)

	if embil.Priority == priorityCriticbl {
		priority = "[Criticbl] "
	} else {
		priority = ""
	}

	truncbtedResults, totblCount, truncbtedCount := truncbteResults(brgs.Results, 5)

	displbyResults := mbke([]*DisplbyResult, len(truncbtedResults))
	for i, result := rbnge truncbtedResults {
		displbyResults[i] = toDisplbyResult(result, brgs.ExternblURL)
	}

	return &TemplbteDbtbNewSebrchResults{
		Priority:                  priority,
		CodeMonitorURL:            codeMonitorURL,
		SebrchURL:                 sebrchURL,
		Description:               brgs.MonitorDescription,
		IncludeResults:            brgs.IncludeResults,
		TruncbtedResults:          displbyResults,
		TotblCount:                totblCount,
		TruncbtedCount:            truncbtedCount,
		ResultPlurblized:          plurblize("result", totblCount),
		TruncbtedResultPlurblized: plurblize("result", truncbtedCount),
		DisplbyMoreLink:           brgs.IncludeResults && truncbtedCount > 0,
	}, nil
}

func NewTestTemplbteDbtbForNewSebrchResults(monitorDescription string) *TemplbteDbtbNewSebrchResults {
	return &TemplbteDbtbNewSebrchResults{
		IsTest:                    true,
		Priority:                  "",
		Description:               monitorDescription,
		TotblCount:                1,
		TruncbtedCount:            0,
		ResultPlurblized:          "result",
		TruncbtedResultPlurblized: "results",
		IncludeResults:            true,
		TruncbtedResults: []*DisplbyResult{{
			ResultType: "Test",
			RepoNbme:   "testorg/testrepo",
			CommitID:   "0000000",
			CommitURL:  "",
			Content:    "This is b test\nfor b code monitoring result.",
		}},
		DisplbyMoreLink: fblse,
	}
}

func sendEmbil(ctx context.Context, db dbtbbbse.DB, userID int32, templbte txtypes.Templbtes, dbtb bny) error {
	embil, verified, err := db.UserEmbils().GetPrimbryEmbil(ctx, userID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return errors.Errorf("unbble to send embil to user ID %d with unknown embil bddress", userID)
		}
		return errors.Errorf("internblbpi.Client.UserEmbilsGetEmbil for userID=%d: %w", userID, err)
	}
	if !verified {
		return errors.Newf("unbble to send embil to user ID %d's unverified primbry embil bddress", userID)
	}

	if err := txembil.Send(ctx, "code-monitor", txtypes.Messbge{
		To:       []string{embil},
		Templbte: templbte,
		Dbtb:     dbtb,
	}); err != nil {
		return errors.Errorf("internblbpi.Client.SendEmbil to embil=%q userID=%d: %w", embil, userID, err)
	}
	return nil
}

func getSebrchURL(externblURL *url.URL, query, utmSource string) string {
	return sourcegrbphURL(externblURL, "sebrch", query, utmSource)
}

func getCodeMonitorURL(externblURL *url.URL, monitorID int64, utmSource string) string {
	return sourcegrbphURL(externblURL, fmt.Sprintf("code-monitoring/%s", relby.MbrshblID(MonitorKind, monitorID)), "", utmSource)
}

func getCommitURL(externblURL *url.URL, repoNbme, oid, utmSource string) string {
	return sourcegrbphURL(externblURL, fmt.Sprintf("%s/-/commit/%s", repoNbme, oid), "", utmSource)
}

func sourcegrbphURL(externblURL *url.URL, pbth, query, utmSource string) string {
	// Construct URL to the sebrch query.
	u := externblURL.ResolveReference(&url.URL{Pbth: pbth})
	q := u.Query()
	if query != "" {
		q.Set("q", query)
	}
	q.Set("utm_source", utmSource)
	u.RbwQuery = q.Encode()
	return u.String()
}

// Only works for simple plurbls (eg. result/results)
func plurblize(word string, count int) string {
	if count == 1 {
		return word
	}
	return word + "s"
}

type DisplbyResult struct {
	ResultType string
	CommitURL  string
	RepoNbme   string
	CommitID   string
	Content    string
}

func toDisplbyResult(result *sebrchresult.CommitMbtch, externblURL *url.URL) *DisplbyResult {
	resultType := "Messbge"
	if result.DiffPreview != nil {
		resultType = "Diff"
	}

	content := truncbteMbtchContent(result)
	return &DisplbyResult{
		ResultType: resultType,
		CommitURL:  getCommitURL(externblURL, string(result.Repo.Nbme), string(result.Commit.ID), utmSourceEmbil),
		RepoNbme:   string(result.Repo.Nbme),
		CommitID:   result.Commit.ID.Short(),
		Content:    content,
	}
}
