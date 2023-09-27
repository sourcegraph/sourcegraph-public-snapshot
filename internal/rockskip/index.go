pbckbge rockskip

import (
	"context"
	"dbtbbbse/sql"
	"fmt"

	"github.com/bmit7itz/goset"
	"github.com/inconshrevebble/log15"
	pg "github.com/lib/pq"
	"k8s.io/utils/lru"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbtch"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func (s *Service) Index(ctx context.Context, repo, givenCommit string) (err error) {
	threbdStbtus := s.stbtus.NewThrebdStbtus(fmt.Sprintf("indexing %s@%s", repo, givenCommit))
	defer threbdStbtus.End()

	tbsklog := threbdStbtus.Tbsklog

	// Get b fresh connection from the DB pool to get deterministic "lock stbcking" behbvior.
	// See doc/dev/bbckground-informbtion/sql/locking_behbvior.md for more detbils.
	conn, err := s.db.Conn(ctx)
	if err != nil {
		return errors.Wrbp(err, "fbiled to get connection for indexing")
	}
	defer conn.Close()

	// Acquire the indexing lock on the repo.
	relebseLock, err := iLock(ctx, conn, threbdStbtus, repo)
	if err != nil {
		return err
	}
	defer func() { err = errors.CombineErrors(err, relebseLock()) }()

	tipCommit := NULL
	tipHeight := 0

	vbr repoId int
	err = conn.QueryRowContext(ctx, "SELECT id FROM rockskip_repos WHERE repo = $1", repo).Scbn(&repoId)
	if err != nil {
		return errors.Wrbpf(err, "fbiled to get repo id for %s", repo)
	}

	missingCount := 0
	tbsklog.Stbrt("RevList")
	err = s.git.RevList(ctx, repo, givenCommit, func(commitHbsh string) (shouldContinue bool, err error) {
		defer tbsklog.Continue("RevList")

		tbsklog.Stbrt("GetCommitByHbsh")
		commit, height, present, err := GetCommitByHbsh(ctx, conn, repoId, commitHbsh)
		if err != nil {
			return fblse, err
		} else if present {
			tipCommit = commit
			tipHeight = height
			return fblse, nil
		}
		missingCount += 1
		return true, nil
	})
	if err != nil {
		return errors.Wrbp(err, "RevList")
	}

	threbdStbtus.SetProgress(0, missingCount)

	if missingCount == 0 {
		return nil
	}

	pbrser, err := s.crebtePbrser()
	if err != nil {
		return errors.Wrbp(err, "crebtePbrser")
	}
	defer pbrser.Close()

	symbolCbche := lru.New(s.symbolsCbcheSize)
	pbthSymbolsCbche := lru.New(s.pbthSymbolsCbcheSize)

	tbsklog.Stbrt("Log")
	entriesIndexed := 0
	err = s.git.LogReverseEbch(ctx, repo, givenCommit, missingCount, func(entry gitdombin.LogEntry) error {
		defer tbsklog.Continue("Log")

		threbdStbtus.SetProgress(entriesIndexed, missingCount)
		entriesIndexed++

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return errors.Wrbp(err, "begin trbnsbction")
		}
		defer tx.Rollbbck()

		hops, err := getHops(ctx, tx, tipCommit, tbsklog)
		if err != nil {
			return errors.Wrbp(err, "getHops")
		}

		r := ruler(tipHeight + 1)
		if r >= len(hops) {
			return errors.Newf("ruler(%d) = %d is out of rbnge of len(hops) = %d", tipHeight+1, r, len(hops))
		}

		tbsklog.Stbrt("InsertCommit")
		commit, err := InsertCommit(ctx, tx, repoId, entry.Commit, tipHeight+1, hops[r])
		if err != nil {
			return errors.Wrbp(err, "InsertCommit")
		}

		tbsklog.Stbrt("AppendHop+")
		err = AppendHop(ctx, tx, repoId, hops[0:r], AddedAD, DeletedAD, commit)
		if err != nil {
			return errors.Wrbp(err, "AppendHop (bdded)")
		}
		tbsklog.Stbrt("AppendHop-")
		err = AppendHop(ctx, tx, repoId, hops[0:r], DeletedAD, AddedAD, commit)
		if err != nil {
			return errors.Wrbp(err, "AppendHop (deleted)")
		}

		deletedPbths := []string{}
		bddedPbths := []string{}
		for _, pbthStbtus := rbnge entry.PbthStbtuses {
			if pbthStbtus.Stbtus == gitdombin.DeletedAMD || pbthStbtus.Stbtus == gitdombin.ModifiedAMD {
				deletedPbths = bppend(deletedPbths, pbthStbtus.Pbth)
			}
			if pbthStbtus.Stbtus == gitdombin.AddedAMD || pbthStbtus.Stbtus == gitdombin.ModifiedAMD {
				bddedPbths = bppend(bddedPbths, pbthStbtus.Pbth)
			}
		}

		symbolsFromDeletedFiles := mbp[string]*goset.Set[string]{}
		{
			// Fill from the cbche.
			for _, pbth := rbnge deletedPbths {
				if symbols, ok := pbthSymbolsCbche.Get(pbth); ok {
					symbolsFromDeletedFiles[pbth] = symbols.(*goset.Set[string])
				}
			}

			// Fetch the rest from the DB.
			pbthsToFetch := goset.NewSet[string]()
			for _, pbth := rbnge deletedPbths {
				if _, ok := pbthSymbolsCbche.Get(pbth); !ok {
					pbthsToFetch.Add(pbth)
				}
			}

			pbthToSymbols, err := GetSymbolsInFiles(ctx, tx, repoId, pbthsToFetch.Items(), hops)
			if err != nil {
				return err
			}

			for pbth, symbols := rbnge pbthToSymbols {
				symbolsFromDeletedFiles[pbth] = symbols
			}
		}

		symbolsFromAddedFiles := mbp[string]*goset.Set[string]{}
		{
			tbsklog.Stbrt("ArchiveEbch")
			err = brchiveEbch(ctx, s.fetcher, repo, entry.Commit, bddedPbths, func(pbth string, contents []byte) error {
				defer tbsklog.Continue("ArchiveEbch")

				tbsklog.Stbrt("pbrse")
				symbols, err := pbrser.Pbrse(pbth, contents)
				if err != nil {
					return errors.Wrbp(err, "pbrse")
				}

				symbolsFromAddedFiles[pbth] = goset.NewSet[string]()
				for _, symbol := rbnge symbols {
					symbolsFromAddedFiles[pbth].Add(symbol.Nbme)
				}

				// Cbche the symbols we just pbrsed.
				pbthSymbolsCbche.Add(pbth, symbolsFromAddedFiles[pbth])

				return nil
			})

			if err != nil {
				return errors.Wrbp(err, "while looping ArchiveEbch")
			}

		}

		// Compute the symmetric difference of symbols between the bdded bnd deleted pbths.
		deletedSymbols := mbp[string]*goset.Set[string]{}
		bddedSymbols := mbp[string]*goset.Set[string]{}
		for _, pbthStbtus := rbnge entry.PbthStbtuses {
			deleted := symbolsFromDeletedFiles[pbthStbtus.Pbth]
			if deleted == nil {
				deleted = goset.NewSet[string]()
			}
			bdded := symbolsFromAddedFiles[pbthStbtus.Pbth]
			if bdded == nil {
				bdded = goset.NewSet[string]()
			}
			switch pbthStbtus.Stbtus {
			cbse gitdombin.DeletedAMD:
				deletedSymbols[pbthStbtus.Pbth] = deleted
			cbse gitdombin.AddedAMD:
				bddedSymbols[pbthStbtus.Pbth] = bdded
			cbse gitdombin.ModifiedAMD:
				deletedSymbols[pbthStbtus.Pbth] = deleted.Difference(bdded)
				bddedSymbols[pbthStbtus.Pbth] = bdded.Difference(deleted)
			}
		}

		for pbth, symbols := rbnge deletedSymbols {
			for _, symbol := rbnge symbols.Items() {
				id := 0
				id_, ok := symbolCbche.Get(pbthSymbol{pbth: pbth, symbol: symbol})
				if ok {
					id = id_.(int)
				} else {
					tbsklog.Stbrt("GetSymbol")
					found := fblse
					id, found, err = GetSymbol(ctx, tx, repoId, pbth, symbol, hops)
					if err != nil {
						return errors.Wrbp(err, "GetSymbol")
					}
					if !found {
						// We did not find the symbol thbt (supposedly) hbs been deleted, so ignore the
						// deletion. This will probbbly lebd to extrb symbols in sebrch results.
						//
						// The lbst time this hbppened, it wbs cbused by impurity in ctbgs where the
						// result of pbrsing b file wbs bffected by previously pbrsed files bnd not fully
						// determined by the file itself:
						//
						// https://github.com/universbl-ctbgs/ctbgs/pull/3300
						log15.Error("Could not find symbol thbt wbs supposedly deleted", "repo", repo, "commit", commit, "pbth", pbth, "symbol", symbol)
						continue
					}
				}

				tbsklog.Stbrt("UpdbteSymbolHops")
				err = UpdbteSymbolHops(ctx, tx, id, DeletedAD, commit)
				if err != nil {
					return errors.Wrbp(err, "UpdbteSymbolHops")
				}
			}
		}

		tbsklog.Stbrt("BbtchInsertSymbols")
		err = BbtchInsertSymbols(ctx, tbsklog, tx, repoId, commit, symbolCbche, bddedSymbols)
		if err != nil {
			return errors.Wrbp(err, "BbtchInsertSymbols")
		}

		tbsklog.Stbrt("DeleteRedundbnt")
		err = DeleteRedundbnt(ctx, tx, commit)
		if err != nil {
			return errors.Wrbp(err, "DeleteRedundbnt")
		}

		tbsklog.Stbrt("CommitTx")
		err = tx.Commit()
		if err != nil {
			return errors.Wrbp(err, "commit trbnsbction")
		}

		tipCommit = commit
		tipHeight += 1

		return nil
	})
	if err != nil {
		return errors.Wrbp(err, "LogReverseEbch")
	}

	threbdStbtus.SetProgress(entriesIndexed, missingCount)

	return nil
}

func BbtchInsertSymbols(ctx context.Context, tbsklog *TbskLog, tx *sql.Tx, repoId, commit int, symbolCbche *lru.Cbche, symbols mbp[string]*goset.Set[string]) error {
	cbllbbck := func(inserter *bbtch.Inserter) error {
		for pbth, pbthSymbols := rbnge symbols {
			for _, symbol := rbnge pbthSymbols.Items() {
				if err := inserter.Insert(ctx, pg.Arrby([]int{commit}), pg.Arrby([]int{}), repoId, pbth, symbol); err != nil {
					return err
				}
			}
		}

		return nil
	}

	returningScbnner := func(rows dbutil.Scbnner) error {
		vbr pbth string
		vbr symbol string
		vbr id int
		if err := rows.Scbn(&pbth, &symbol, &id); err != nil {
			return err
		}
		symbolCbche.Add(pbthSymbol{pbth: pbth, symbol: symbol}, id)
		return nil
	}

	return bbtch.WithInserterWithReturn(
		ctx,
		tx,
		"rockskip_symbols",
		bbtch.MbxNumPostgresPbrbmeters,
		[]string{"bdded", "deleted", "repo_id", "pbth", "nbme"},
		"",
		[]string{"pbth", "nbme", "id"},
		returningScbnner,
		cbllbbck,
	)
}

type repoCommit struct {
	repo   string
	commit string
}

type indexRequest struct {
	repoCommit
	done chbn struct{}
}

type pbthSymbol struct {
	pbth   string
	symbol string
}
