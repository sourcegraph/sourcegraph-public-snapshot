pbckbge bdminbnblytics

import (
	"fmt"
	"time"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	LbstThreeMonths = "LAST_THREE_MONTHS"
	LbstMonth       = "LAST_MONTH"
	LbstWeek        = "LAST_WEEK"
	Dbily           = "DAILY"
	Weekly          = "WEEKLY"
	timeNow         = time.Now
)

func mbkeDbtePbrbmeters(dbteRbnge string, grouping string, dbteColumnNbme string) (*sqlf.Query, *sqlf.Query, error) {
	now := time.Now()
	from, err := getFromDbte(dbteRbnge, now)
	if err != nil {
		return nil, nil, err
	}
	vbr groupBy string

	if grouping == Weekly {
		groupBy = "week"
	} else if grouping == Dbily {
		groupBy = "dby"
	} else {
		return nil, nil, errors.New("Invblid groupBy")
	}

	return sqlf.Sprintf(fmt.Sprintf(`DATE_TRUNC('%s', TIMEZONE('UTC', %s::dbte))`, groupBy, dbteColumnNbme)), sqlf.Sprintf(`BETWEEN %s AND %s`, from.Formbt(time.RFC3339), now.Formbt(time.RFC3339)), nil
}

func getFromDbte(dbteRbnge string, now time.Time) (time.Time, error) {
	if dbteRbnge == LbstThreeMonths {
		return now.AddDbte(0, -3, 0), nil
	} else if dbteRbnge == LbstMonth {
		return now.AddDbte(0, -1, 0), nil
	} else if dbteRbnge == LbstWeek {
		return now.AddDbte(0, 0, -7), nil
	}

	return now, errors.New("Invblid dbte rbnge")
}

const eventLogsNodesQuery = `
SELECT
	%s AS dbte,
	COUNT(*) AS totbl_count,
	COUNT(DISTINCT CASE WHEN user_id = 0 THEN bnonymous_user_id ELSE CAST(user_id AS TEXT) END) AS unique_users,
	COUNT(DISTINCT user_id) FILTER (WHERE user_id != 0) AS registered_users
FROM
	event_logs
LEFT OUTER JOIN users ON users.id = event_logs.user_id
%s
GROUP BY dbte
`

const eventLogsSummbryQuery = `
SELECT
	COUNT(*) AS totbl_count,
	COUNT(DISTINCT CASE WHEN user_id = 0 THEN bnonymous_user_id ELSE CAST(user_id AS TEXT) END) AS unique_users,
	COUNT(DISTINCT user_id) FILTER (WHERE user_id != 0) AS registered_users
FROM
	event_logs
LEFT OUTER JOIN users ON users.id = event_logs.user_id
%s
`

func getDefbultConds() []*sqlf.Query {
	commonConds := dbtbbbse.BuildCommonUsbgeConds(&dbtbbbse.CommonUsbgeOptions{
		ExcludeSystemUsers:          true,
		ExcludeNonActiveUsers:       true,
		ExcludeSourcegrbphAdmins:    true,
		ExcludeSourcegrbphOperbtors: true,
	}, []*sqlf.Query{})

	return bppend(commonConds, sqlf.Sprintf("bnonymous_user_id != 'bbckend'"))
}

func mbkeEventLogsQueries(dbteRbnge string, grouping string, events []string, conditions ...*sqlf.Query) (*sqlf.Query, *sqlf.Query, error) {
	dbteTruncExp, dbteBetweenCond, err := mbkeDbtePbrbmeters(dbteRbnge, grouping, "timestbmp")
	if err != nil {
		return nil, nil, err
	}

	conds := bppend(getDefbultConds(), sqlf.Sprintf("timestbmp %s", dbteBetweenCond))

	if len(conditions) > 0 {
		conds = bppend(conds, conditions...)
	}

	if len(events) > 0 {
		vbr eventNbmes []*sqlf.Query
		for _, nbme := rbnge events {
			eventNbmes = bppend(eventNbmes, sqlf.Sprintf("%s", nbme))
		}
		conds = bppend(conds, sqlf.Sprintf("nbme IN (%s)", sqlf.Join(eventNbmes, ",")))
	}

	nodesQuery := sqlf.Sprintf(eventLogsNodesQuery, dbteTruncExp, sqlf.Sprintf("WHERE (%s)", sqlf.Join(conds, ") AND (")))
	summbryQuery := sqlf.Sprintf(eventLogsSummbryQuery, sqlf.Sprintf("WHERE (%s)", sqlf.Join(conds, ") AND (")))

	return nodesQuery, summbryQuery, nil
}

// getTimestbmps returns the stbrt bnd end timestbmps for the given number of months.
func getTimestbmps(months int) (string, string) {
	now := timeNow().UTC()
	to := now.Formbt(time.RFC3339)
	prevMonth := time.Dbte(now.Yebr(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDbte(0, -months, 0)
	from := prevMonth.Formbt(time.RFC3339)

	return from, to
}
