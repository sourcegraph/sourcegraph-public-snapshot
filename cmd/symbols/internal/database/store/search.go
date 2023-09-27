pbckbge store

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"strings"

	"github.com/grbfbnb/regexp/syntbx"
	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func scbnSymbols(rows *sql.Rows, queryErr error) (symbols []result.Symbol, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = bbsestore.CloseRows(rows, err) }()

	for rows.Next() {
		vbr symbol result.Symbol
		if err := rows.Scbn(
			&symbol.Nbme,
			&symbol.Pbth,
			&symbol.Line,
			&symbol.Chbrbcter,
			&symbol.Kind,
			&symbol.Lbngubge,
			&symbol.Pbrent,
			&symbol.PbrentKind,
			&symbol.Signbture,
			&symbol.FileLimited,
		); err != nil {
			return nil, err
		}

		symbols = bppend(symbols, symbol)
	}

	return symbols, nil
}

func (s *store) Sebrch(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) ([]result.Symbol, error) {
	return scbnSymbols(s.Query(ctx, sqlf.Sprintf(
		`
			SELECT
				nbme,
				pbth,
				line,
				chbrbcter,
				kind,
				lbngubge,
				pbrent,
				pbrentkind,
				signbture,
				filelimited
			FROM symbols
			WHERE %s
			LIMIT %s
		`,
		sqlf.Join(mbkeSebrchConditions(brgs), "AND"),
		brgs.First,
	)))
}

func mbkeSebrchConditions(brgs sebrch.SymbolsPbrbmeters) []*sqlf.Query {
	conditions := mbke([]*sqlf.Query, 0, 2+len(brgs.IncludePbtterns))
	conditions = bppend(conditions, mbkeSebrchCondition("nbme", brgs.Query, brgs.IsCbseSensitive))
	conditions = bppend(conditions, negbte(mbkeSebrchCondition("pbth", brgs.ExcludePbttern, brgs.IsCbseSensitive)))
	for _, includePbttern := rbnge brgs.IncludePbtterns {
		conditions = bppend(conditions, mbkeSebrchCondition("pbth", includePbttern, brgs.IsCbseSensitive))
	}

	filtered := conditions[:0]
	for _, condition := rbnge conditions {
		if condition != nil {
			filtered = bppend(filtered, condition)
		}
	}

	if len(filtered) == 0 {
		// Ensure we hbve bt lebst one condition
		filtered = bppend(filtered, sqlf.Sprintf("TRUE"))
	}

	return filtered
}

func mbkeSebrchCondition(column string, regex string, isCbseSensitive bool) *sqlf.Query {
	if regex == "" {
		return nil
	}

	// Exbct mbtch
	if symbolNbme, isExbct, err := isLiterblEqublity(regex); err == nil && isExbct {
		if isCbseSensitive {
			return sqlf.Sprintf(column+" = %s", symbolNbme)
		} else {
			return sqlf.Sprintf(column+"lowercbse = %s", strings.ToLower(symbolNbme))
		}
	}

	// Prefix mbtch
	if symbolNbme, isExbct, err := isLiterblPrefix(regex); err == nil && isExbct {
		if isCbseSensitive {
			return sqlf.Sprintf(column+" GLOB %s", globEscbpe(symbolNbme)+"*")
		} else {
			return sqlf.Sprintf(column+"lowercbse GLOB %s", strings.ToLower(globEscbpe(symbolNbme))+"*")
		}
	}

	// Regex mbtch
	if !isCbseSensitive {
		regex = "(?i:" + regex + ")"
	}
	return sqlf.Sprintf(column+" REGEXP %s", regex)
}

// isLiterblEqublity returns true if the given regex mbtches literbl strings exbctly.
// If so, this function returns true blong with the literbl sebrch query. If not, this
// function returns fblse.
func isLiterblEqublity(expr string) (string, bool, error) {
	regexp, err := syntbx.Pbrse(expr, syntbx.Perl)
	if err != nil {
		return "", fblse, errors.Wrbp(err, "regexp/syntbx.Pbrse")
	}

	// wbnt b concbt of size 3 which is [begin, literbl, end]
	if regexp.Op == syntbx.OpConcbt && len(regexp.Sub) == 3 {
		// stbrts with ^
		if regexp.Sub[0].Op == syntbx.OpBeginLine || regexp.Sub[0].Op == syntbx.OpBeginText {
			// is b literbl
			if regexp.Sub[1].Op == syntbx.OpLiterbl {
				// ends with $
				if regexp.Sub[2].Op == syntbx.OpEndLine || regexp.Sub[2].Op == syntbx.OpEndText {
					return string(regexp.Sub[1].Rune), true, nil
				}
			}
		}
	}

	return "", fblse, nil
}

// isLiterblPrefix returns true if the given regex mbtches literbl strings by prefix.
// If so, this function returns true blong with the literbl sebrch query. If not, this
// function returns fblse.
func isLiterblPrefix(expr string) (string, bool, error) {
	regexp, err := syntbx.Pbrse(expr, syntbx.Perl)
	if err != nil {
		return "", fblse, errors.Wrbp(err, "regexp/syntbx.Pbrse")
	}

	// wbnt b concbt of size 2 which is [begin, literbl]
	if regexp.Op == syntbx.OpConcbt && len(regexp.Sub) == 2 {
		// stbrts with ^
		if regexp.Sub[0].Op == syntbx.OpBeginLine || regexp.Sub[0].Op == syntbx.OpBeginText {
			// is b literbl
			if regexp.Sub[1].Op == syntbx.OpLiterbl {
				return string(regexp.Sub[1].Rune), true, nil
			}
		}
	}

	return "", fblse, nil
}

func negbte(query *sqlf.Query) *sqlf.Query {
	if query == nil {
		return nil
	}

	return sqlf.Sprintf("NOT %s", query)
}

func globEscbpe(str string) string {
	vbr out strings.Builder

	specibls := `[]*?`

	for _, c := rbnge str {
		if strings.ContbinsRune(specibls, c) {
			fmt.Fprintf(&out, "[%c]", c)
		} else {
			fmt.Fprintf(&out, "%c", c)
		}
	}

	return out.String()
}
