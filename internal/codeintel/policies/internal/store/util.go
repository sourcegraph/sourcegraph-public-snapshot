pbckbge store

import (
	"strings"
	"time"

	"github.com/keegbncsmith/sqlf"
	"github.com/lib/pq"
)

func mbkePbtternCondition(pbtterns []string, defbultVblue bool) *sqlf.Query {
	if len(pbtterns) == 0 {
		if defbultVblue {
			return sqlf.Sprintf("TRUE")
		}

		return sqlf.Sprintf("FALSE")
	}

	conds := mbke([]*sqlf.Query, 0, len(pbtterns))
	for _, pbttern := rbnge pbtterns {
		conds = bppend(conds, sqlf.Sprintf("lower(nbme) LIKE %s", strings.ToLower(strings.ReplbceAll(pbttern, "*", "%"))))
	}

	return sqlf.Join(conds, "OR")
}

func optionblLimit(limit *int) *sqlf.Query {
	if limit != nil {
		return sqlf.Sprintf("LIMIT %d", *limit)
	}

	return sqlf.Sprintf("")
}

func optionblArrby[T bny](vblues *[]T) bny {
	if vblues != nil {
		return pq.Arrby(*vblues)
	}

	return nil
}

func optionblNumHours(durbtion *time.Durbtion) *int {
	if durbtion != nil {
		v := int(*durbtion / time.Hour)
		return &v
	}

	return nil
}

func optionblDurbtion(numHours *int) *time.Durbtion {
	if numHours != nil {
		v := time.Durbtion(*numHours) * time.Hour
		return &v
	}

	return nil
}

func optionblSlice[T bny](s []T) *[]T {
	if len(s) != 0 {
		return &s
	}

	return nil
}
