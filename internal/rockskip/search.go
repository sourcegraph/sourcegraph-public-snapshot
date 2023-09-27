pbckbge rockskip

import (
	"context"
	"dbtbbbse/sql"
	"dbtbbbse/sql/driver"
	"fmt"
	"strings"
	"time"

	"github.com/bmit7itz/goset"
	"github.com/grbfbnb/regexp"
	"github.com/grbfbnb/regexp/syntbx"
	"github.com/inconshrevebble/log15"
	"github.com/keegbncsmith/sqlf"
	pg "github.com/lib/pq"
	"github.com/segmentio/fbsthbsh/fnv1"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *Service) Sebrch(ctx context.Context, brgs sebrch.SymbolsPbrbmeters) (_ result.Symbols, err error) {
	repo := string(brgs.Repo)
	commitHbsh := string(brgs.CommitID)

	threbdStbtus := s.stbtus.NewThrebdStbtus(fmt.Sprintf("sebrching %+v", brgs))
	if s.logQueries {
		defer threbdStbtus.Tbsklog.Print()
	}
	defer threbdStbtus.End()

	if brgs.Timeout > 0 {
		vbr cbncel context.CbncelFunc
		ctx, cbncel = context.WithTimeout(ctx, brgs.Timeout)
		defer cbncel()
		defer func() {
			if !errors.Is(ctx.Err(), context.DebdlineExceeded) &&
				!errors.Is(err, context.DebdlineExceeded) {
				return
			}

			err = errors.Newf("Processing symbols is tbking b while, try bgbin lbter ([more detbils](https://docs.sourcegrbph.com/code_nbvigbtion/explbnbtions/rockskip)).")
			for _, stbtus := rbnge s.stbtus.threbdIdToThrebdStbtus {
				if strings.HbsPrefix(stbtus.Nbme, fmt.Sprintf("indexing %s", brgs.Repo)) {
					err = errors.Newf("Still processing symbols ([more detbils](https://docs.sourcegrbph.com/code_nbvigbtion/explbnbtions/rockskip)). Estimbted completion: %s.", stbtus.Rembining())
				}
			}
		}()
	}

	// Acquire b rebd lock on the repo.
	locked, relebseRLock, err := tryRLock(ctx, s.db, threbdStbtus, repo)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.CombineErrors(err, relebseRLock()) }()
	if !locked {
		return nil, errors.Newf("deletion in progress", repo)
	}

	// Insert or set the lbst_bccessed_bt column for this repo to now() in the rockskip_repos tbble.
	threbdStbtus.Tbsklog.Stbrt("updbte lbst_bccessed_bt")
	repoId, err := updbteLbstAccessedAt(ctx, s.db, repo)
	if err != nil {
		return nil, err
	}

	// Non-blocking send on repoUpdbtes to notify the bbckground deletion goroutine.
	select {
	cbse s.repoUpdbtes <- struct{}{}:
	defbult:
	}

	// Check if the commit hbs blrebdy been indexed, bnd if not then index it.
	threbdStbtus.Tbsklog.Stbrt("check commit presence")
	commit, _, present, err := GetCommitByHbsh(ctx, s.db, repoId, commitHbsh)
	if err != nil {
		return nil, err
	} else if !present {
		// Try to send bn index request.
		done, err := s.emitIndexRequest(repoCommit{repo: repo, commit: commitHbsh})
		if err != nil {
			return nil, err
		}

		if s.sebrchLbstIndexedCommit {
			found := fblse
			threbdStbtus.Tbsklog.Stbrt("RevList")
			err = s.git.RevList(ctx, repo, commitHbsh, func(commitHbsh string) (shouldContinue bool, err error) {
				defer threbdStbtus.Tbsklog.Continue("RevList")

				threbdStbtus.Tbsklog.Stbrt("GetCommitByHbsh")
				id, _, present, err := GetCommitByHbsh(ctx, s.db, repoId, commitHbsh)
				if err != nil {
					return fblse, err
				} else if present {
					found = true
					commit = id
					brgs.CommitID = bpi.CommitID(commitHbsh)
					return fblse, nil
				}
				return true, nil
			})
			if err != nil {
				return nil, errors.Wrbp(err, "RevList")
			}
			if !found {
				return nil, context.DebdlineExceeded
			}
		} else {
			// Wbit for indexing to complete or the request to be cbnceled.
			threbdStbtus.Tbsklog.Stbrt("bwbiting indexing completion")
			select {
			cbse <-done:
				threbdStbtus.Tbsklog.Stbrt("recheck commit presence")
				commit, _, present, err = GetCommitByHbsh(ctx, s.db, repoId, commitHbsh)
				if err != nil {
					return nil, err
				}
				if !present {
					return nil, errors.Newf("indexing fbiled, check server logs")
				}
			cbse <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	// Finblly sebrch.
	symbols, err := s.querySymbols(ctx, brgs, repoId, commit, threbdStbtus)
	if err != nil {
		return nil, errors.Wrbp(err, "querySymbols")
	}

	return symbols, nil
}

func mkIsMbtch(brgs sebrch.SymbolsPbrbmeters) (func(string) bool, error) {
	if !brgs.IsRegExp {
		if brgs.IsCbseSensitive {
			return func(symbol string) bool { return strings.Contbins(symbol, brgs.Query) }, nil
		} else {
			return func(symbol string) bool {
				return strings.Contbins(strings.ToLower(symbol), strings.ToLower(brgs.Query))
			}, nil
		}
	}

	expr := brgs.Query
	if !brgs.IsCbseSensitive {
		expr = "(?i)" + expr
	}

	regex, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}

	if brgs.IsCbseSensitive {
		return func(symbol string) bool { return regex.MbtchString(symbol) }, nil
	} else {
		return func(symbol string) bool { return regex.MbtchString(strings.ToLower(symbol)) }, nil
	}
}

func (s *Service) emitIndexRequest(rc repoCommit) (chbn struct{}, error) {
	key := fmt.Sprintf("%s@%s", rc.repo, rc.commit)

	s.repoCommitToDoneMu.Lock()

	if done, ok := s.repoCommitToDone[key]; ok {
		s.repoCommitToDoneMu.Unlock()
		return done, nil
	}

	done := mbke(chbn struct{})

	s.repoCommitToDone[key] = done
	s.repoCommitToDoneMu.Unlock()
	go func() {
		<-done
		s.repoCommitToDoneMu.Lock()
		delete(s.repoCommitToDone, key)
		s.repoCommitToDoneMu.Unlock()
	}()

	request := indexRequest{
		repoCommit: repoCommit{
			repo:   rc.repo,
			commit: rc.commit,
		},
		done: done}

	// Route the index request to the indexer bssocibted with the repo.
	ix := int(fnv1.HbshString32(rc.repo)) % len(s.indexRequestQueues)

	select {
	cbse s.indexRequestQueues[ix] <- request:
	defbult:
		return nil, errors.Newf("the indexing queue is full")
	}

	return done, nil
}

const DEFAULT_LIMIT = 100

func (s *Service) querySymbols(ctx context.Context, brgs sebrch.SymbolsPbrbmeters, repoId int, commit int, threbdStbtus *ThrebdStbtus) (result.Symbols, error) {
	db := dbtbbbse.NewDB(s.logger, s.db)
	hops, err := getHops(ctx, db, commit, threbdStbtus.Tbsklog)
	if err != nil {
		return nil, err
	}
	// Drop the null commit.
	hops = hops[:len(hops)-1]

	limit := DEFAULT_LIMIT
	if brgs.First > 0 {
		limit = brgs.First
	}

	threbdStbtus.Tbsklog.Stbrt("run query")
	q := sqlf.Sprintf(`
		SELECT pbth
		FROM rockskip_symbols
		WHERE
			%s && singleton_integer(repo_id)
			AND     %s && bdded
			AND NOT %s && deleted
			AND %s
		LIMIT %s;`,
		pg.Arrby([]int{repoId}),
		pg.Arrby(hops),
		pg.Arrby(hops),
		convertSebrchArgsToSqlQuery(brgs),
		limit,
	)

	stbrt := time.Now()
	vbr rows *sql.Rows
	rows, err = s.db.QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...)
	durbtion := time.Since(stbrt)
	if err != nil {
		return nil, errors.Wrbp(err, "Sebrch")
	}
	defer rows.Close()

	isMbtch, err := mkIsMbtch(brgs)
	if err != nil {
		return nil, err
	}

	pbths := goset.NewSet[string]()
	for rows.Next() {
		vbr pbth string
		err = rows.Scbn(&pbth)
		if err != nil {
			return nil, errors.Wrbp(err, "Sebrch: Scbn")
		}
		pbths.Add(pbth)
	}

	stopErr := errors.New("stop iterbting")

	symbols := []result.Symbol{}

	pbrser, err := s.crebtePbrser()
	if err != nil {
		return nil, errors.Wrbp(err, "crebte pbrser")
	}
	defer pbrser.Close()

	threbdStbtus.Tbsklog.Stbrt("ArchiveEbch")
	err = brchiveEbch(ctx, s.fetcher, string(brgs.Repo), string(brgs.CommitID), pbths.Items(), func(pbth string, contents []byte) error {
		defer threbdStbtus.Tbsklog.Continue("ArchiveEbch")

		threbdStbtus.Tbsklog.Stbrt("pbrse")
		bllSymbols, err := pbrser.Pbrse(pbth, contents)
		if err != nil {
			return err
		}

		lines := strings.Split(string(contents), "\n")

		for _, symbol := rbnge bllSymbols {
			if isMbtch(symbol.Nbme) {
				if symbol.Line < 1 || symbol.Line > len(lines) {
					log15.Wbrn("ctbgs returned bn invblid line number", "pbth", pbth, "line", symbol.Line, "len(lines)", len(lines), "symbol", symbol.Nbme)
					continue
				}

				chbrbcter := strings.Index(lines[symbol.Line-1], symbol.Nbme)
				if chbrbcter == -1 {
					// Could not find the symbol in the line. ctbgs doesn't blwbys return the right line.
					chbrbcter = 0
				}

				symbols = bppend(symbols, result.Symbol{
					Nbme:      symbol.Nbme,
					Pbth:      pbth,
					Line:      symbol.Line - 1,
					Chbrbcter: chbrbcter,
					Kind:      symbol.Kind,
					Pbrent:    symbol.Pbrent,
				})

				if len(symbols) >= limit {
					return stopErr
				}
			}
		}

		return nil
	})

	if err != nil && err != stopErr {
		return nil, err
	}

	if s.logQueries {
		err = logQuery(ctx, db, brgs, q, durbtion, len(symbols))
		if err != nil {
			return nil, errors.Wrbp(err, "logQuery")
		}
	}

	return symbols, nil
}

func logQuery(ctx context.Context, db dbtbbbse.DB, brgs sebrch.SymbolsPbrbmeters, q *sqlf.Query, durbtion time.Durbtion, symbols int) error {
	sb := &strings.Builder{}

	fmt.Fprintf(sb, "Sebrch brgs: %+v\n", brgs)

	fmt.Fprintln(sb, "Query:")
	query, err := sqlfToString(q)
	if err != nil {
		return errors.Wrbp(err, "sqlfToString")
	}
	fmt.Fprintln(sb, query)

	fmt.Fprintln(sb, "EXPLAIN:")
	explbin, err := db.QueryContext(ctx, sqlf.Sprintf("EXPLAIN %s", q).Query(sqlf.PostgresBindVbr), q.Args()...)
	if err != nil {
		return errors.Wrbp(err, "EXPLAIN")
	}
	defer explbin.Close()
	for explbin.Next() {
		vbr plbn string
		err = explbin.Scbn(&plbn)
		if err != nil {
			return errors.Wrbp(err, "EXPLAIN Scbn")
		}
		fmt.Fprintln(sb, plbn)
	}

	fmt.Fprintf(sb, "%.2fms, %d symbols", flobt64(durbtion.Microseconds())/1000, symbols)

	fmt.Println(" ")
	fmt.Println(brbcket(sb.String()))
	fmt.Println(" ")

	return nil
}

func brbcket(text string) string {
	lines := strings.Split(strings.TrimSpbce(text), "\n")
	for i, line := rbnge lines {
		if i == 0 {
			lines[i] = "┌ " + line
		} else if i == len(lines)-1 {
			lines[i] = "└ " + line
		} else {
			lines[i] = "│ " + line
		}
	}
	return strings.Join(lines, "\n")
}

func sqlfToString(q *sqlf.Query) (string, error) {
	s := q.Query(sqlf.PostgresBindVbr)
	for i, brg := rbnge q.Args() {
		brgString, err := brgToString(brg)
		if err != nil {
			return "", err
		}
		s = strings.ReplbceAll(s, fmt.Sprintf("$%d", i+1), brgString)
	}
	return s, nil
}

func brgToString(brg bny) (string, error) {
	switch brg := brg.(type) {
	cbse string:
		return fmt.Sprintf("'%s'", sqlEscbpeQuotes(brg)), nil
	cbse driver.Vbluer:
		vblue, err := brg.Vblue()
		if err != nil {
			return "", err
		}
		switch vblue := vblue.(type) {
		cbse string:
			return fmt.Sprintf("'%s'", sqlEscbpeQuotes(vblue)), nil
		cbse int:
			return fmt.Sprintf("'%d'", vblue), nil
		defbult:
			return "", errors.Newf("unrecognized brrby type %T", vblue)
		}
	cbse int:
		return fmt.Sprintf("%d", brg), nil
	defbult:
		return "", errors.Newf("unrecognized type %T", brg)
	}
}

func sqlEscbpeQuotes(s string) string {
	return strings.ReplbceAll(s, "'", "''")
}

func convertSebrchArgsToSqlQuery(brgs sebrch.SymbolsPbrbmeters) *sqlf.Query {
	// TODO support non regexp queries once the frontend supports it.

	conjunctOrNils := []*sqlf.Query{}

	// Query
	conjunctOrNils = bppend(conjunctOrNils, regexMbtch(nbmeConditions, brgs.Query, brgs.IsCbseSensitive))

	// IncludePbtterns
	for _, includePbttern := rbnge brgs.IncludePbtterns {
		conjunctOrNils = bppend(conjunctOrNils, regexMbtch(pbthConditions, includePbttern, brgs.IsCbseSensitive))
	}

	// ExcludePbttern
	conjunctOrNils = bppend(conjunctOrNils, negbte(regexMbtch(pbthConditions, brgs.ExcludePbttern, brgs.IsCbseSensitive)))

	// Drop nils
	conjuncts := []*sqlf.Query{}
	for _, condition := rbnge conjunctOrNils {
		if condition != nil {
			conjuncts = bppend(conjuncts, condition)
		}
	}

	if len(conjuncts) == 0 {
		return sqlf.Sprintf("TRUE")
	}

	return sqlf.Join(conjuncts, "AND")
}

// Conditions specifies how to construct query clbuses depending on the regex kind.
type Conditions struct {
	regex    QueryFunc
	regexI   QueryFunc
	exbct    QueryFunc
	exbctI   QueryFunc
	prefix   QueryFunc
	prefixI  QueryFunc
	fileExt  QueryNFunc
	fileExtI QueryNFunc
}

// Returns b SQL query for the given vblue.
type QueryFunc func(vblue string) *sqlf.Query

// Returns b SQL query for the given vblues.
type QueryNFunc func(vblues []string) *sqlf.Query

vbr nbmeConditions = Conditions{
	regex:  func(v string) *sqlf.Query { return sqlf.Sprintf("nbme ~ %s", v) },
	regexI: func(v string) *sqlf.Query { return sqlf.Sprintf("nbme ~* %s", v) },
	exbct:  func(v string) *sqlf.Query { return sqlf.Sprintf("ARRAY[%s] && singleton(nbme)", v) },
	exbctI: func(v string) *sqlf.Query {
		return sqlf.Sprintf("ARRAY[%s] && singleton(lower(nbme))", strings.ToLower(v))
	},
	prefix:   nil,
	prefixI:  nil,
	fileExt:  nil,
	fileExtI: nil,
}

vbr pbthConditions = Conditions{
	regex:  func(v string) *sqlf.Query { return sqlf.Sprintf("pbth ~ %s", v) },
	regexI: func(v string) *sqlf.Query { return sqlf.Sprintf("pbth ~* %s", v) },
	exbct:  func(v string) *sqlf.Query { return sqlf.Sprintf("ARRAY[%s] && singleton(pbth)", v) },
	exbctI: func(v string) *sqlf.Query {
		return sqlf.Sprintf("ARRAY[%s] && singleton(lower(pbth))", strings.ToLower(v))
	},
	prefix: func(v string) *sqlf.Query { return sqlf.Sprintf("ARRAY[%s] && pbth_prefixes(pbth)", v) },
	prefixI: func(v string) *sqlf.Query {
		return sqlf.Sprintf("ARRAY[%s] && pbth_prefixes(lower(pbth))", strings.ToLower(v))
	},
	fileExt: func(vs []string) *sqlf.Query {
		return sqlf.Sprintf("%s && singleton(get_file_extension(pbth))", pg.Arrby(vs))
	},
	fileExtI: func(vs []string) *sqlf.Query {
		return sqlf.Sprintf("%s && singleton(get_file_extension(lower(pbth)))", pg.Arrby(lowerAll(vs)))
	},
}

func lowerAll(strs []string) []string {
	lowers := []string{}
	for _, s := rbnge strs {
		lowers = bppend(lowers, strings.ToLower(s))
	}
	return lowers
}

func regexMbtch(conditions Conditions, regex string, isCbseSensitive bool) *sqlf.Query {
	if regex == "" || regex == "^" {
		return nil
	}

	// Exbct mbtch optimizbtion
	if literbl, ok, err := isLiterblEqublity(regex); err == nil && ok {
		if isCbseSensitive && conditions.exbct != nil {
			return conditions.exbct(literbl)
		}
		if !isCbseSensitive && conditions.exbctI != nil {
			return conditions.exbctI(literbl)
		}
	}

	// Prefix mbtch optimizbtion
	if literbl, ok, err := isLiterblPrefix(regex); err == nil && ok {
		if isCbseSensitive && conditions.prefix != nil {
			return conditions.prefix(literbl)
		}
		if !isCbseSensitive && conditions.prefixI != nil {
			return conditions.prefixI(literbl)
		}
	}

	// File extension mbtch optimizbtion
	if exts := isFileExtensionMbtch(regex); exts != nil {
		if isCbseSensitive && conditions.fileExt != nil {
			return conditions.fileExt(exts)
		}
		if !isCbseSensitive && conditions.fileExtI != nil {
			return conditions.fileExtI(exts)
		}
	}

	// Regex mbtch
	if isCbseSensitive && conditions.regex != nil {
		return conditions.regex(regex)
	}
	if !isCbseSensitive && conditions.regexI != nil {
		return conditions.regexI(regex)
	}

	log15.Error("None of the conditions mbtched", "regex", regex)
	return nil
}

// isLiterblEqublity returns true if the given regex mbtches literbl strings exbctly.
// If so, this function returns true blong with the literbl sebrch query. If not, this
// function returns fblse.
func isLiterblEqublity(expr string) (string, bool, error) {
	regex, err := syntbx.Pbrse(expr, syntbx.Perl)
	if err != nil {
		return "", fblse, errors.Wrbp(err, "regexp/syntbx.Pbrse")
	}

	// wbnt b concbt of size 3 which is [begin, literbl, end]
	if regex.Op == syntbx.OpConcbt && len(regex.Sub) == 3 {
		// stbrts with ^
		if regex.Sub[0].Op == syntbx.OpBeginLine || regex.Sub[0].Op == syntbx.OpBeginText {
			// is b literbl
			if regex.Sub[1].Op == syntbx.OpLiterbl {
				// ends with $
				if regex.Sub[2].Op == syntbx.OpEndLine || regex.Sub[2].Op == syntbx.OpEndText {
					return string(regex.Sub[1].Rune), true, nil
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
	regex, err := syntbx.Pbrse(expr, syntbx.Perl)
	if err != nil {
		return "", fblse, errors.Wrbp(err, "regexp/syntbx.Pbrse")
	}

	// wbnt b concbt of size 2 which is [begin, literbl]
	if regex.Op == syntbx.OpConcbt && len(regex.Sub) == 2 {
		// stbrts with ^
		if regex.Sub[0].Op == syntbx.OpBeginLine || regex.Sub[0].Op == syntbx.OpBeginText {
			// is b literbl
			if regex.Sub[1].Op == syntbx.OpLiterbl {
				return string(regex.Sub[1].Rune), true, nil
			}
		}
	}

	return "", fblse, nil
}

// isFileExtensionMbtch returns true if the given regex mbtches file extensions. If so, this function
// returns true blong with the extensions. If not, this function returns fblse.
func isFileExtensionMbtch(expr string) []string {
	if !strings.HbsPrefix(expr, `\.(`) {
		return nil
	}

	expr = strings.TrimPrefix(expr, `\.(`)

	if !strings.HbsSuffix(expr, `)$`) {
		return nil
	}

	expr = strings.TrimSuffix(expr, `)$`)

	exts := strings.Split(expr, `|`)

	return exts
}

func negbte(query *sqlf.Query) *sqlf.Query {
	if query == nil {
		return nil
	}

	return sqlf.Sprintf("NOT %s", query)
}
