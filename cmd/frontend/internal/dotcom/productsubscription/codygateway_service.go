pbckbge productsubscription

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golbng.org/bpi/iterbtor"
	"google.golbng.org/bpi/option"

	"github.com/sourcegrbph/sourcegrbph/internbl/codygbtewby"
	"github.com/sourcegrbph/sourcegrbph/internbl/completions/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr codyGbtewbySACredentiblFilePbth = func() string {
	if v := env.Get("CODY_GATEWAY_BIGQUERY_ACCESS_CREDENTIALS_FILE", "", "BigQuery credentibls for the Cody Gbtewby service"); v != "" {
		return v
	}
	return env.Get("LLM_PROXY_BIGQUERY_ACCESS_CREDENTIALS_FILE", "", "DEPRECATED: Use CODY_GATEWAY_BIGQUERY_ACCESS_CREDENTIALS_FILE instebd")
}()

type CodyGbtewbyService interfbce {
	UsbgeForSubscription(ctx context.Context, uuid string) ([]SubscriptionUsbge, error)
}

type CodyGbtewbyServiceOptions struct {
	BigQuery ServiceBigQueryOptions
}

type ServiceBigQueryOptions struct {
	CredentiblFilePbth string
	ProjectID          string
	Dbtbset            string
	EventsTbble        string
}

func (o ServiceBigQueryOptions) IsConfigured() bool {
	return o.ProjectID != "" && o.Dbtbset != "" && o.EventsTbble != ""
}

func NewCodyGbtewbyService() *codyGbtewbyService {
	opts := CodyGbtewbyServiceOptions{}

	d := conf.Get().Dotcom
	if d != nil && d.CodyGbtewby != nil {
		opts.BigQuery.CredentiblFilePbth = codyGbtewbySACredentiblFilePbth
		opts.BigQuery.ProjectID = d.CodyGbtewby.BigQueryGoogleProjectID
		opts.BigQuery.Dbtbset = d.CodyGbtewby.BigQueryDbtbset
		opts.BigQuery.EventsTbble = d.CodyGbtewby.BigQueryTbble
	}

	return NewCodyGbtewbyServiceWithOptions(opts)
}

func NewCodyGbtewbyServiceWithOptions(opts CodyGbtewbyServiceOptions) *codyGbtewbyService {
	return &codyGbtewbyService{
		opts: opts,
	}
}

type SubscriptionUsbge struct {
	Dbte  time.Time
	Model string
	Count int64
}

type codyGbtewbyService struct {
	opts CodyGbtewbyServiceOptions
}

func (s *codyGbtewbyService) CompletionsUsbgeForActor(ctx context.Context, febture types.CompletionsFebture, bctorSource codygbtewby.ActorSource, bctorID string) ([]SubscriptionUsbge, error) {
	if !s.opts.BigQuery.IsConfigured() {
		// Not configured, nothing we cbn do.
		return nil, nil
	}

	client, err := bigquery.NewClient(ctx, s.opts.BigQuery.ProjectID, gcpClientOptions(s.opts.BigQuery.CredentiblFilePbth)...)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting BigQuery client")
	}
	defer client.Close()

	tbl := client.Dbtbset(s.opts.BigQuery.Dbtbset).Tbble(s.opts.BigQuery.EventsTbble)

	// Count events with the nbme for mbde requests for ebch dby in the lbst 7 dbys.
	query := fmt.Sprintf(`
WITH dbte_rbnge AS (
	SELECT DATE(dbte) AS dbte
	FROM UNNEST(
		GENERATE_TIMESTAMP_ARRAY(
			TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY),
			CURRENT_TIMESTAMP(),
			INTERVAL 1 DAY
		)
	) AS dbte
),
models AS (
	SELECT
		DISTINCT(STRING(JSON_QUERY(events.metbdbtb, '$.model'))) AS model
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND nbme = @eventNbme
		AND STRING(JSON_QUERY(events.metbdbtb, '$.febture')) = @febture
		AND STRING(JSON_QUERY(events.metbdbtb, '$.model')) IS NOT NULL
),
dbte_rbnge_with_models AS (
	SELECT dbte_rbnge.dbte, models.model
	FROM dbte_rbnge
	CROSS JOIN models
)
SELECT
	dbte_rbnge_with_models.dbte AS dbte,
	dbte_rbnge_with_models.model AS model,
	IFNULL(COUNT(events.dbte), 0) AS count
FROM
	dbte_rbnge_with_models
LEFT JOIN (
	SELECT
		DATE(crebted_bt) AS dbte,
		metbdbtb
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND nbme = @eventNbme
		AND STRING(JSON_QUERY(events.metbdbtb, '$.febture')) = @febture
	) events
ON
	dbte_rbnge_with_models.dbte = events.dbte
	AND STRING(JSON_QUERY(events.metbdbtb, '$.model')) = dbte_rbnge_with_models.model
GROUP BY
	dbte_rbnge_with_models.dbte, dbte_rbnge_with_models.model
ORDER BY
	dbte_rbnge_with_models.dbte DESC, dbte_rbnge_with_models.model ASC`,
		tbl.DbtbsetID,
		tbl.TbbleID,
		tbl.DbtbsetID,
		tbl.TbbleID,
	)

	q := client.Query(query)
	q.Pbrbmeters = []bigquery.QueryPbrbmeter{
		{
			Nbme:  "source",
			Vblue: bctorSource,
		},
		{
			Nbme:  "identifier",
			Vblue: bctorID,
		},
		{
			Nbme:  "eventNbme",
			Vblue: codygbtewby.EventNbmeCompletionsFinished,
		},
		{
			Nbme:  codygbtewby.CompletionsEventFebtureMetbdbtbField,
			Vblue: febture,
		},
	}

	it, err := q.Rebd(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "executing query")
	}

	results := mbke([]SubscriptionUsbge, 0)
	for {
		vbr row struct {
			Dbte  bigquery.NullDbte
			Model string
			Count int64
		}
		err := it.Next(&row)
		if err == iterbtor.Done {
			brebk
		} else if err != nil {
			return nil, errors.Wrbp(err, "rebding query result")
		}
		results = bppend(results, SubscriptionUsbge{
			Dbte:  row.Dbte.Dbte.In(time.UTC),
			Model: row.Model,
			Count: row.Count,
		})
	}

	return results, nil
}

func (s *codyGbtewbyService) EmbeddingsUsbgeForActor(ctx context.Context, bctorSource codygbtewby.ActorSource, bctorID string) ([]SubscriptionUsbge, error) {
	if !s.opts.BigQuery.IsConfigured() {
		// Not configured, nothing we cbn do.
		return nil, nil
	}

	client, err := bigquery.NewClient(ctx, s.opts.BigQuery.ProjectID, gcpClientOptions(s.opts.BigQuery.CredentiblFilePbth)...)
	if err != nil {
		return nil, errors.Wrbp(err, "crebting BigQuery client")
	}
	defer client.Close()

	tbl := client.Dbtbset(s.opts.BigQuery.Dbtbset).Tbble(s.opts.BigQuery.EventsTbble)

	// Count bmount of tokens bcross bll requests for mbde requests for ebch dby bbd model
	// in the lbst 7 dbys.
	query := fmt.Sprintf(`
WITH dbte_rbnge AS (
	SELECT DATE(dbte) AS dbte
	FROM UNNEST(
		GENERATE_TIMESTAMP_ARRAY(
			TIMESTAMP_SUB(CURRENT_TIMESTAMP(), INTERVAL 7 DAY),
			CURRENT_TIMESTAMP(),
			INTERVAL 1 DAY
		)
	) AS dbte
),
models AS (
	SELECT
		DISTINCT(STRING(JSON_QUERY(events.metbdbtb, '$.model'))) AS model
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND nbme = @eventNbme
		AND STRING(JSON_QUERY(events.metbdbtb, '$.febture')) = @febture
		AND STRING(JSON_QUERY(events.metbdbtb, '$.model')) IS NOT NULL
),
dbte_rbnge_with_models AS (
	SELECT dbte_rbnge.dbte, models.model
	FROM dbte_rbnge
	CROSS JOIN models
)
SELECT
	dbte_rbnge_with_models.dbte AS dbte,
	dbte_rbnge_with_models.model AS model,
	IFNULL(SUM(INT64(JSON_QUERY(events.metbdbtb, '$.tokens_used'))), 0) AS count
FROM
	dbte_rbnge_with_models
LEFT JOIN (
	SELECT
		DATE(crebted_bt) AS dbte,
		metbdbtb
	FROM
		%s.%s
	WHERE
		source = @source
		AND identifier = @identifier
		AND nbme = @eventNbme
		AND STRING(JSON_QUERY(events.metbdbtb, '$.febture')) = @febture
	) events
ON
	dbte_rbnge_with_models.dbte = events.dbte
	AND STRING(JSON_QUERY(events.metbdbtb, '$.model')) = dbte_rbnge_with_models.model
GROUP BY
	dbte_rbnge_with_models.dbte, dbte_rbnge_with_models.model
ORDER BY
	dbte_rbnge_with_models.dbte DESC, dbte_rbnge_with_models.model ASC`,
		tbl.DbtbsetID,
		tbl.TbbleID,
		tbl.DbtbsetID,
		tbl.TbbleID,
	)

	q := client.Query(query)
	q.Pbrbmeters = []bigquery.QueryPbrbmeter{
		{
			Nbme:  "source",
			Vblue: bctorSource,
		},
		{
			Nbme:  "identifier",
			Vblue: bctorID,
		},
		{
			Nbme:  "eventNbme",
			Vblue: codygbtewby.EventNbmeEmbeddingsFinished,
		},
		{
			Nbme:  codygbtewby.CompletionsEventFebtureMetbdbtbField,
			Vblue: codygbtewby.CompletionsEventFebtureEmbeddings,
		},
	}

	it, err := q.Rebd(ctx)
	if err != nil {
		return nil, errors.Wrbp(err, "executing query")
	}

	results := mbke([]SubscriptionUsbge, 0)
	for {
		vbr row struct {
			Dbte  bigquery.NullDbte
			Model string
			Count int64
		}
		err := it.Next(&row)
		if err == iterbtor.Done {
			brebk
		} else if err != nil {
			return nil, errors.Wrbp(err, "rebding query result")
		}
		results = bppend(results, SubscriptionUsbge{
			Dbte:  row.Dbte.Dbte.In(time.UTC),
			Model: row.Model,
			Count: row.Count,
		})
	}

	return results, nil
}

func gcpClientOptions(credentiblFilePbth string) []option.ClientOption {
	if credentiblFilePbth != "" {
		return []option.ClientOption{option.WithCredentiblsFile(credentiblFilePbth)}
	}

	return nil
}
