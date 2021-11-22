package sqlite

import (
	"regexp/syntax"
	"strings"

	"github.com/keegancsmith/sqlf"
)

func makeCondition(column string, regex string, isCaseSensitive bool) []*sqlf.Query {
	conditions := []*sqlf.Query{}

	if regex == "" {
		return conditions
	}

	if isExact, symbolName, err := isLiteralEquality(regex); isExact && err == nil {
		// It looks like the user is asking for exact matches, so use `=` to
		// get the speed boost from the index on the column.
		if isCaseSensitive {
			conditions = append(conditions, sqlf.Sprintf(column+" = %s", symbolName))
		} else {
			conditions = append(conditions, sqlf.Sprintf(column+"lowercase = %s", strings.ToLower(symbolName)))
		}
	} else {
		if !isCaseSensitive {
			regex = "(?i:" + regex + ")"
		}
		conditions = append(conditions, sqlf.Sprintf(column+" REGEXP %s", regex))
	}

	return conditions
}

func negateAll(oldConditions []*sqlf.Query) []*sqlf.Query {
	newConditions := []*sqlf.Query{}

	for _, oldCondition := range oldConditions {
		newConditions = append(newConditions, sqlf.Sprintf("NOT %s", oldCondition))
	}

	return newConditions
}

// isLiteralEquality checks if the given regex matches literal strings exactly.
// Returns whether or not the regex is exact, along with the literal string if
// so.
func isLiteralEquality(expr string) (ok bool, lit string, err error) {
	r, err := syntax.Parse(expr, syntax.Perl)
	if err != nil {
		return false, "", err
	}
	// Want a Concat of size 3 which is [Begin, Literal, End]
	if r.Op != syntax.OpConcat || len(r.Sub) != 3 || // size 3 concat
		!(r.Sub[0].Op == syntax.OpBeginLine || r.Sub[0].Op == syntax.OpBeginText) || // Starts with ^
		!(r.Sub[2].Op == syntax.OpEndLine || r.Sub[2].Op == syntax.OpEndText) || // Ends with $
		r.Sub[1].Op != syntax.OpLiteral { // is a literal
		return false, "", nil
	}
	return true, string(r.Sub[1].Rune), nil
}
