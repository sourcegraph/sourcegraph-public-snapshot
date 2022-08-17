package insights

import "github.com/keegancsmith/sqlf"

func grantQuery(tableName string, userID, orgID *int32) *sqlf.Query {
	table := sqlf.Sprintf(tableName)

	if userID != nil {
		return sqlf.Sprintf("%s.user_id = %s", table, *userID)
	}

	if orgID != nil {
		return sqlf.Sprintf("%s.org_id = %s", table, *orgID)
	}

	return sqlf.Sprintf("%s.global IS TRUE", table)
}

func grantValues2(userID, orgID *int32) []any {
	if userID != nil {
		return []any{*userID, nil, nil}
	}

	if orgID != nil {
		return []any{nil, *orgID, nil}
	}

	return []any{nil, nil, true}
}

func grantValues(userID int, orgIDs []int) []any {
	if userID != 0 {
		return []any{userID, nil, nil}
	}

	if len(orgIDs) != 0 {
		return []any{nil, orgIDs[0], nil}
	}

	return []any{nil, nil, true}
}
