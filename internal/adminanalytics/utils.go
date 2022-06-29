package adminanalytics

import (
	"errors"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
)

var (
	LastThreeMonths = "LAST_THREE_MONTHS"
	LastMonth       = "LAST_MONTH"
	LastWeek        = "LAST_WEEK"
)

func makeStringsInExpression(values []string) *sqlf.Query {
	var conds []*sqlf.Query
	for _, value := range values {
		conds = append(conds, sqlf.Sprintf("%s", value))
	}
	return sqlf.Join(conds, ",")
}

func makeDateParameters(dateRange string, dateColumnName string) (*sqlf.Query, *sqlf.Query, error) {
	now := time.Now()
	var from time.Time
	var groupBy string

	if dateRange == LastThreeMonths {
		from = now.AddDate(0, -3, 0)
		groupBy = "week"
	} else if dateRange == LastMonth {
		from = now.AddDate(0, -1, 0)
		groupBy = "day"
	} else if dateRange == LastWeek {
		from = now.AddDate(0, 0, -7)
		groupBy = "day"
	} else {
		return nil, nil, errors.New("Invalid date range")
	}

	return sqlf.Sprintf(fmt.Sprintf(`DATE_TRUNC('%s', %s::date)`, groupBy, dateColumnName)), sqlf.Sprintf(`BETWEEN %s AND %s`, from.Format(time.RFC3339), now.Format(time.RFC3339)), nil
}
