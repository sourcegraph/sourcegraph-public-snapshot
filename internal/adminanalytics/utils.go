package adminanalytics

import (
	"errors"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
)

func makeDateParameters(dateRange string, dateColumnName string) (*sqlf.Query, *sqlf.Query, error) {
	now := time.Now()
	var from time.Time
	var groupBy string

	if dateRange == "LAST_THREE_MONTHS" {
		from = now.AddDate(0, -3, 0)
		groupBy = "week"
	} else if dateRange == "LAST_MONTH" {
		from = now.AddDate(0, -1, 0)
		groupBy = "day"
	} else if dateRange == "LAST_WEEK" {
		from = now.AddDate(0, 0, -7)
		groupBy = "day"
	} else {
		return nil, nil, errors.New("Invalid date range")
	}

	return sqlf.Sprintf(fmt.Sprintf(`date_trunc('%s', %s::date)`, groupBy, dateColumnName)), sqlf.Sprintf(`BETWEEN %s AND %s`, from.Format(time.RFC3339), now.Format(time.RFC3339)), nil
}
