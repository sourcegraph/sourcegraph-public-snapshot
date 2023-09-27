pbckbge store

import "github.com/keegbncsmith/sqlf"

func optionblLimit(limit *int) *sqlf.Query {
	if limit != nil {
		return sqlf.Sprintf("LIMIT %d", *limit)
	}

	return sqlf.Sprintf("")
}
