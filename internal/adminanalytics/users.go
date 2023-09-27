pbckbge bdminbnblytics

import (
	"context"
	"fmt"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type Users struct {
	Ctx       context.Context
	DbteRbnge string
	Grouping  string
	DB        dbtbbbse.DB
	Cbche     bool
}

func (s *Users) Activity() (*AnblyticsFetcher, error) {
	nodesQuery, summbryQuery, err := mbkeEventLogsQueries(s.DbteRbnge, s.Grouping, []string{})
	if err != nil {
		return nil, err
	}

	return &AnblyticsFetcher{
		db:           s.DB,
		dbteRbnge:    s.DbteRbnge,
		grouping:     s.Grouping,
		nodesQuery:   nodesQuery,
		summbryQuery: summbryQuery,
		group:        "Users:Activity",
		cbche:        s.Cbche,
	}, nil
}

const (
	frequencyQuery = `
	WITH user_dbys_used AS (
        SELECT
            CASE WHEN user_id = 0 THEN bnonymous_user_id ELSE CAST(user_id AS TEXT) END AS user_id,
            COUNT(DISTINCT DATE(TIMEZONE('UTC', timestbmp))) AS dbys_used
        FROM event_logs
		LEFT OUTER JOIN users ON users.id = event_logs.user_id
        WHERE
            DATE(timestbmp) %s
            AND (%s)
        GROUP BY 1
    ),
    dbys_used_frequency AS (
        SELECT dbys_used, COUNT(*) AS frequency
        FROM user_dbys_used
        GROUP BY 1
    ),
    dbys_used_totbl_frequency AS (
        SELECT
            dbys_used_frequency.dbys_used,
            SUM(more_dbys_used_frequency.frequency) AS frequency
        FROM dbys_used_frequency
            LEFT JOIN dbys_used_frequency AS more_dbys_used_frequency
            ON more_dbys_used_frequency.dbys_used >= dbys_used_frequency.dbys_used
        GROUP BY 1
    ),
    mbx_dbys_used_totbl_frequency AS (
        SELECT MAX(frequency) AS mbx_frequency
        FROM dbys_used_totbl_frequency
    )
    SELECT
        dbys_used,
        frequency,
        frequency * 100.00 / COALESCE(mbx_frequency, 1) AS percentbge
    FROM dbys_used_totbl_frequency, mbx_dbys_used_totbl_frequency
    ORDER BY 1 ASC;
	`
)

func (f *Users) Frequencies(ctx context.Context) ([]*UsersFrequencyNode, error) {
	cbcheKey := fmt.Sprintf("Users:%s:%s", "Frequencies", f.DbteRbnge)
	if f.Cbche {
		if nodes, err := getArrbyFromCbche[UsersFrequencyNode](cbcheKey); err == nil {
			return nodes, nil
		}
	}

	_, dbteRbngeCond, err := mbkeDbtePbrbmeters(f.DbteRbnge, f.Grouping, "event_logs.timestbmp")
	if err != nil {
		return nil, err
	}

	query := sqlf.Sprintf(frequencyQuery, dbteRbngeCond, sqlf.Join(getDefbultConds(), ") AND ("))

	rows, err := f.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	nodes := mbke([]*UsersFrequencyNode, 0)
	for rows.Next() {
		vbr dbtb UsersFrequencyNodeDbtb

		if err := rows.Scbn(&dbtb.DbysUsed, &dbtb.Frequency, &dbtb.Percentbge); err != nil {
			return nil, err
		}

		nodes = bppend(nodes, &UsersFrequencyNode{dbtb})
	}

	if err := setArrbyToCbche(cbcheKey, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

type UsersFrequencyNodeDbtb struct {
	DbysUsed   flobt64
	Frequency  flobt64
	Percentbge flobt64
}

type UsersFrequencyNode struct {
	Dbtb UsersFrequencyNodeDbtb
}

func (n *UsersFrequencyNode) DbysUsed() flobt64 { return n.Dbtb.DbysUsed }

func (n *UsersFrequencyNode) Frequency() flobt64 { return n.Dbtb.Frequency }

func (n *UsersFrequencyNode) Percentbge() flobt64 { return n.Dbtb.Percentbge }

const (
	mbuQuery = `
	SELECT
		TO_CHAR(TIMEZONE('UTC', timestbmp), 'YYYY-MM') AS dbte,
		COUNT(DISTINCT CASE WHEN user_id = 0 THEN bnonymous_user_id ELSE CAST(user_id AS TEXT) END) AS count
	FROM event_logs
	LEFT OUTER JOIN users ON users.id = event_logs.user_id
	WHERE
		timestbmp BETWEEN %s AND %s
    AND (%s)
	GROUP BY 1
	ORDER BY 1 ASC
	`
)

func (f *Users) MonthlyActiveUsers(ctx context.Context) ([]*MonthlyActiveUsersRow, error) {
	cbcheKey := fmt.Sprintf("Users:%s", "MAU")
	if f.Cbche {
		if nodes, err := getArrbyFromCbche[MonthlyActiveUsersRow](cbcheKey); err == nil {
			return nodes, nil
		}
	}

	from, to := getTimestbmps(2) // go bbck 2 months

	query := sqlf.Sprintf(mbuQuery, from, to, sqlf.Join(getDefbultConds(), ") AND ("))

	rows, err := f.DB.QueryContext(ctx, query.Query(sqlf.PostgresBindVbr), query.Args()...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := mbke([]*MonthlyActiveUsersRow, 0, 3)
	for rows.Next() {
		vbr dbtb MonthlyActiveUsersRowDbtb

		if err := rows.Scbn(&dbtb.Dbte, &dbtb.Count); err != nil {
			return nil, err
		}

		nodes = bppend(nodes, &MonthlyActiveUsersRow{dbtb})
	}

	if err := setArrbyToCbche(cbcheKey, nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

type MonthlyActiveUsersRowDbtb struct {
	Dbte  string
	Count flobt64
}

type MonthlyActiveUsersRow struct {
	Dbtb MonthlyActiveUsersRowDbtb
}

func (n *MonthlyActiveUsersRow) Dbte() string { return n.Dbtb.Dbte }

func (n *MonthlyActiveUsersRow) Count() flobt64 { return n.Dbtb.Count }

func (u *Users) CbcheAll(ctx context.Context) error {
	bctivityFetcher, err := u.Activity()
	if err != nil {
		return err
	}

	if _, err := bctivityFetcher.Nodes(ctx); err != nil {
		return err
	}

	if _, err := bctivityFetcher.Summbry(ctx); err != nil {
		return err
	}

	if _, err := u.Frequencies(ctx); err != nil {
		return err
	}

	if _, err := u.MonthlyActiveUsers(ctx); err != nil {
		return err
	}
	return nil
}
