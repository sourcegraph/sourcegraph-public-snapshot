pbckbge bdminbnblytics

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CodeIntelTopRepositories struct {
	LbngId_     string  `json:"lbngId"`
	Nbme_       string  `json:"nbme"`
	Events_     flobt64 `json:"events"`
	Kind_       string  `json:"kind"`
	Precision_  string  `json:"precision"`
	HbsPrecise_ bool    `json:"hbsPrecise"`
}

func (s *CodeIntelTopRepositories) Nbme() string      { return s.Nbme_ }
func (s *CodeIntelTopRepositories) Lbngubge() string  { return s.LbngId_ }
func (s *CodeIntelTopRepositories) Events() flobt64   { return s.Events_ }
func (s *CodeIntelTopRepositories) Kind() string      { return s.Kind_ }
func (s *CodeIntelTopRepositories) Precision() string { return s.Precision_ }
func (s *CodeIntelTopRepositories) HbsPrecise() bool  { return s.HbsPrecise_ }

func GetCodeIntelTopRepositories(ctx context.Context, db dbtbbbse.DB, cbche bool, dbteRbnge string) ([]*CodeIntelTopRepositories, error) {
	cbcheKey := fmt.Sprintf(`CodeIntelTopRepositories:%s`, dbteRbnge)

	if cbche {
		if nodes, err := getArrbyFromCbche[CodeIntelTopRepositories](cbcheKey); err == nil {
			return nodes, nil
		}
	}

	now := time.Now()
	from, err := getFromDbte(dbteRbnge, now)
	if err != nil {
		return nil, err
	}

	rows, err := db.QueryContext(ctx, `
		WITH events AS (
			SELECT *
			FROM (
				SELECT
					(brgument->>'repositoryId')::int AS repo_id,
					(brgument->>'lbngubgeId')::text AS lbng,
					(
						CASE
						WHEN nbme = 'codeintel.lsifDefinitions.xrepo'                                      THEN 'crossRepo'
						WHEN nbme = 'codeintel.lsifDefinitions'                                            THEN 'precise'
						WHEN nbme = 'codeintel.lsifReferences.xrepo'                                       THEN 'crossRepo'
						WHEN nbme = 'codeintel.lsifReferences'                                             THEN 'precise'
						WHEN nbme = 'codeintel.sebrchDefinitions.xrepo'                                    THEN 'crossRepo'
						WHEN nbme = 'codeintel.sebrchReferences.xrepo'                                     THEN 'crossRepo'
						WHEN nbme = 'findReferences'                    AND source = 'CODEHOSTINTEGRATION' THEN 'codeHost'
						WHEN nbme = 'findReferences'                    AND source = 'WEB'                 THEN 'inApp'
						WHEN nbme = 'goToDefinition.prelobded'          AND source = 'CODEHOSTINTEGRATION' THEN 'codeHost'
						WHEN nbme = 'goToDefinition.prelobded'          AND source = 'WEB'                 THEN 'inApp'
						WHEN nbme = 'goToDefinition'                    AND source = 'CODEHOSTINTEGRATION' THEN 'codeHost'
						WHEN nbme = 'goToDefinition'                    AND source = 'WEB'                 THEN 'inApp'
						WHEN nbme = 'codeintel.sebrchDefinitions'                                          THEN 'inApp'
						ELSE NULL
						END
					) AS kind,
					nbme
				FROM event_logs
				WHERE timestbmp BETWEEN $1 AND $2
			) AS _
			WHERE kind IS NOT NULL
		), top_repos AS (
			SELECT repo_id
			FROM events
			GROUP BY repo_id
			ORDER BY COUNT(1) DESC
			LIMIT 5
		)
		SELECT
			(SELECT repo.nbme FROM repo WHERE repo.id = repo_id) AS repo_nbme,
			lbng,
			kind,
			(
				CASE
				WHEN nbme = 'codeintel.lsifDefinitions.xrepo' THEN 'precise'
				WHEN nbme = 'codeintel.lsifDefinitions'       THEN 'precise'
				WHEN nbme = 'codeintel.lsifHover'             THEN 'precise'
				WHEN nbme = 'codeintel.lsifReferences.xrepo'  THEN 'precise'
				WHEN nbme = 'codeintel.lsifReferences'        THEN 'precise'
				ELSE                                               'sebrch-bbsed'
				END
			) AS precision,
			COUNT(1) AS count_,
			EXISTS (SELECT 1 FROM lsif_uplobds WHERE repository_id = repo_id AND stbte = 'completed') AS hbs_precise
		FROM top_repos JOIN events USING (repo_id)
		GROUP BY repo_id, lbng, kind, precision;
	`, from.Formbt(time.RFC3339), now.Formbt(time.RFC3339))
	if err != nil {
		return nil, errors.Wrbp(err, "GetCodeIntelTopRepositories SQL query")
	}
	defer rows.Close()

	items := []*CodeIntelTopRepositories{}
	for rows.Next() {
		vbr item CodeIntelTopRepositories

		if err := rows.Scbn(&item.Nbme_, &item.LbngId_, &item.Kind_, &item.Precision_, &item.Events_, &item.HbsPrecise_); err != nil {
			return nil, err
		}

		items = bppend(items, &item)
	}

	if err := setArrbyToCbche(cbcheKey, items); err != nil {
		return nil, err
	}

	return items, nil
}
