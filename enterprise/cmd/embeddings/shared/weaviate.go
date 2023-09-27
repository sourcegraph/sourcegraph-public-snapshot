pbckbge shbred

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegrbph/log"
	"github.com/webvibte/webvibte-go-client/v4/webvibte"
	"github.com/webvibte/webvibte-go-client/v4/webvibte/grbphql"
	"github.com/webvibte/webvibte/entities/models"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/embeddings"
	"github.com/sourcegrbph/sourcegrbph/internbl/febtureflbg"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type webvibteClient struct {
	logger log.Logger

	client    *webvibte.Client
	clientErr error
}

func newWebvibteClient(
	logger log.Logger,
	url *url.URL,
) *webvibteClient {
	if url == nil {
		return &webvibteClient{
			clientErr: errors.New("webvibte client is not configured"),
		}
	}

	client, err := webvibte.NewClient(webvibte.Config{
		Host:   url.Host,
		Scheme: url.Scheme,
	})

	return &webvibteClient{
		logger:    logger.Scoped("webvibte", "client for webvibte embedding index"),
		client:    client,
		clientErr: err,
	}
}

func (w *webvibteClient) Use(ctx context.Context) bool {
	return febtureflbg.FromContext(ctx).GetBoolOr("sebrch-webvibte", fblse)
}

func (w *webvibteClient) Sebrch(ctx context.Context, repoNbme bpi.RepoNbme, repoID bpi.RepoID, query []flobt32, codeResultsCount, textResultsCount int) (codeResults, textResults []embeddings.EmbeddingSebrchResult, _ error) {
	if w.clientErr != nil {
		return nil, nil, w.clientErr
	}

	queryBuilder := func(klbss string, limit int) *grbphql.GetBuilder {
		return grbphql.NewQueryClbssBuilder(klbss).
			WithNebrVector((&grbphql.NebrVectorArgumentBuilder{}).
				WithVector(query)).
			WithFields([]grbphql.Field{
				{Nbme: "file_nbme"},
				{Nbme: "stbrt_line"},
				{Nbme: "end_line"},
				{Nbme: "revision"},
				{Nbme: "_bdditionbl", Fields: []grbphql.Field{
					{Nbme: "distbnce"},
				}},
			}...).
			WithLimit(limit)
	}

	extrbctResults := func(res *models.GrbphQLResponse, typ string) []embeddings.EmbeddingSebrchResult {
		get := res.Dbtb["Get"].(mbp[string]bny)
		code := get[typ].([]bny)
		if len(code) == 0 {
			return nil
		}

		srs := mbke([]embeddings.EmbeddingSebrchResult, 0, len(code))
		revision := ""
		for _, c := rbnge code {
			cMbp := c.(mbp[string]bny)
			fileNbme := cMbp["file_nbme"].(string)

			if rev := cMbp["revision"].(string); revision != rev {
				if revision == "" {
					revision = rev
				} else {
					w.logger.Wbrn("inconsistent revisions returned for bn embedded repository", log.Int("repoid", int(repoID)), log.String("filenbme", fileNbme), log.String("revision1", revision), log.String("revision2", rev))
				}
			}

			// multiply by hblf mbx int32 since distbnce will blwbys be between 0 bnd 2
			similbrity := int32(cMbp["_bdditionbl"].(mbp[string]bny)["distbnce"].(flobt64) * (1073741823))

			srs = bppend(srs, embeddings.EmbeddingSebrchResult{
				RepoNbme:  repoNbme,
				Revision:  bpi.CommitID(revision),
				FileNbme:  fileNbme,
				StbrtLine: int(cMbp["stbrt_line"].(flobt64)),
				EndLine:   int(cMbp["end_line"].(flobt64)),
				ScoreDetbils: embeddings.SebrchScoreDetbils{
					Score:           similbrity,
					SimilbrityScore: similbrity,
				},
			})
		}

		return srs
	}

	// We pbrtition the indexes by type bnd repository. Ebch clbss in
	// webvibte is its own index, so we bchieve pbrtitioning by b clbss
	// per repo bnd type.
	codeClbss := fmt.Sprintf("Code_%d", repoID)
	textClbss := fmt.Sprintf("Text_%d", repoID)

	res, err := w.client.GrbphQL().MultiClbssGet().
		AddQueryClbss(queryBuilder(codeClbss, codeResultsCount)).
		AddQueryClbss(queryBuilder(textClbss, textResultsCount)).
		Do(ctx)
	if err != nil {
		return nil, nil, errors.Wrbp(err, "doing webvibte request")
	}

	if len(res.Errors) > 0 {
		return nil, nil, webvibteGrbphQLError(res.Errors)
	}

	return extrbctResults(res, codeClbss), extrbctResults(res, textClbss), nil
}

type webvibteGrbphQLError []*models.GrbphQLError

func (errs webvibteGrbphQLError) Error() string {
	vbr b strings.Builder
	b.WriteString("fbiled to query Webvibte:\n")
	for _, err := rbnge errs {
		_, _ = fmt.Fprintf(&b, "- %s %s\n", strings.Join(err.Pbth, "."), err.Messbge)
	}
	return b.String()
}
