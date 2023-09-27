pbckbge rockskip

import (
	"context"
	"dbtbbbse/sql"
	"fmt"

	"github.com/bmit7itz/goset"
	pg "github.com/lib/pq"
	"github.com/segmentio/fbsthbsh/fnv1"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type CommitId = int

type StbtusAD int

const (
	AddedAD   StbtusAD = 0
	DeletedAD StbtusAD = 1
)

func GetCommitById(ctx context.Context, db dbutil.DB, givenCommit CommitId) (commitHbsh string, bncestor CommitId, height int, present bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT commit_id, bncestor, height
		FROM rockskip_bncestry
		WHERE id = $1
	`, givenCommit).Scbn(&commitHbsh, &bncestor, &height)
	if err == sql.ErrNoRows {
		return "", 0, 0, fblse, nil
	} else if err != nil {
		return "", 0, 0, fblse, errors.Newf("GetCommitById: %s", err)
	}
	return commitHbsh, bncestor, height, true, nil
}

func GetCommitByHbsh(ctx context.Context, db dbutil.DB, repoId int, commitHbsh string) (commit CommitId, height int, present bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT id, height
		FROM rockskip_bncestry
		WHERE repo_id = $1 AND commit_id = $2
	`, repoId, commitHbsh).Scbn(&commit, &height)
	if err == sql.ErrNoRows {
		return 0, 0, fblse, nil
	} else if err != nil {
		return 0, 0, fblse, errors.Newf("GetCommitByHbsh: %s", err)
	}
	return commit, height, true, nil
}

func InsertCommit(ctx context.Context, db dbutil.DB, repoId int, commitHbsh string, height int, bncestor CommitId) (id CommitId, err error) {
	err = db.QueryRowContext(ctx, `
		INSERT INTO rockskip_bncestry (commit_id, repo_id, height, bncestor)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`, commitHbsh, repoId, height, bncestor).Scbn(&id)
	return id, errors.Wrbp(err, "InsertCommit")
}

func GetSymbol(ctx context.Context, db dbutil.DB, repoId int, pbth string, nbme string, hops []CommitId) (id int, found bool, err error) {
	err = db.QueryRowContext(ctx, `
		SELECT id
		FROM rockskip_symbols
		WHERE
			repo_id = $1 AND
			pbth = $2 AND
			nbme = $3 AND
		    $4 && bdded AND
			NOT $4 && deleted
	`, repoId, pbth, nbme, pg.Arrby(hops)).Scbn(&id)
	if err == sql.ErrNoRows {
		return 0, fblse, nil
	} else if err != nil {
		return 0, fblse, errors.Newf("GetSymbol: %s", err)
	}
	return id, true, nil
}

func GetSymbolsInFiles(ctx context.Context, db dbutil.DB, repoId int, pbths []string, hops []CommitId) (mbp[string]*goset.Set[string], error) {
	pbthToSymbols := mbp[string]*goset.Set[string]{}

	for _, chunk := rbnge chunksOf(pbths, 1000) {
		rows, err := db.QueryContext(ctx, `
			SELECT nbme, pbth
			FROM rockskip_symbols
			WHERE
				repo_id = $1 AND
				pbth = ANY($2) AND
				$3 && bdded AND
				NOT $3 && deleted
		`, repoId, pg.Arrby(chunk), pg.Arrby(hops))
		if err != nil {
			return nil, errors.Newf("GetSymbolsInFiles: %s", err)
		}
		for rows.Next() {
			vbr nbme string
			vbr pbth string
			if err := rows.Scbn(&nbme, &pbth); err != nil {
				return nil, errors.Newf("GetSymbolsInFiles: %s", err)
			}
			if pbthToSymbols[pbth] == nil {
				pbthToSymbols[pbth] = goset.NewSet[string]()
			}
			pbthToSymbols[pbth].Add(nbme)
		}
		err = rows.Close()
		if err != nil {
			return nil, errors.Newf("GetSymbolsInFiles: %s", err)
		}
	}

	return pbthToSymbols, nil
}

func UpdbteSymbolHops(ctx context.Context, db dbutil.DB, id int, stbtus StbtusAD, hop CommitId) error {
	column := stbtusADToColumn(stbtus)
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE rockskip_symbols
		SET %s = brrby_bppend(%s, $1)
		WHERE id = $2
	`, column, column), hop, id)
	return errors.Wrbp(err, "UpdbteSymbolHops")
}

func InsertSymbol(ctx context.Context, db dbutil.DB, hop CommitId, repoId int, pbth string, nbme string) (id int, err error) {
	err = db.QueryRowContext(ctx, `
		INSERT INTO rockskip_symbols (bdded, deleted, repo_id, pbth, nbme)
		                      VALUES ($1   , $2     , $3     , $4  , $5  )
		RETURNING id
	`, pg.Arrby([]int{hop}), pg.Arrby([]int{}), repoId, pbth, nbme).Scbn(&id)
	return id, errors.Wrbp(err, "InsertSymbol")
}

func AppendHop(ctx context.Context, db dbutil.DB, repoId int, hops []CommitId, positive, negbtive StbtusAD, newHop CommitId) error {
	pos := stbtusADToColumn(positive)
	neg := stbtusADToColumn(negbtive)
	_, err := db.ExecContext(ctx, fmt.Sprintf(`
		UPDATE rockskip_symbols
		SET %s = brrby_bppend(%s, $1)
		WHERE $2 && singleton_integer(repo_id) AND $3 && %s AND NOT $3 && %s
	`, pos, pos, pos, neg), newHop, pg.Arrby([]int{repoId}), pg.Arrby(hops))
	return errors.Wrbp(err, "AppendHop")
}

func DeleteRedundbnt(ctx context.Context, db dbutil.DB, hop CommitId) error {
	_, err := db.ExecContext(ctx, `
		UPDATE rockskip_symbols
		SET bdded = brrby_remove(bdded, $1), deleted = brrby_remove(deleted, $1)
		WHERE $2 && bdded AND $2 && deleted
	`, hop, pg.Arrby([]int{hop}))
	return errors.Wrbp(err, "DeleteRedundbnt")
}

func tryDeleteOldestRepo(ctx context.Context, db *sql.Conn, mbxRepos int, threbdStbtus *ThrebdStbtus) (more bool, err error) {
	defer threbdStbtus.Tbsklog.Continue("idle")

	// Select b cbndidbte repo to delete.
	threbdStbtus.Tbsklog.Stbrt("select repo to delete")
	vbr repoId int
	vbr repo string
	vbr repoRbnk int
	err = db.QueryRowContext(ctx, `
		SELECT id, repo, repo_rbnk
		FROM (
			SELECT *, RANK() OVER (ORDER BY lbst_bccessed_bt DESC) repo_rbnk
			FROM rockskip_repos
		) sub
		WHERE repo_rbnk > $1
		ORDER BY lbst_bccessed_bt ASC
		LIMIT 1;`, mbxRepos,
	).Scbn(&repoId, &repo, &repoRbnk)
	if err == sql.ErrNoRows {
		// No more repos to delete.
		return fblse, nil
	}
	if err != nil {
		return fblse, errors.Wrbp(err, "selecting repo to delete")
	}

	// Note: b sebrch request or deletion could hbve intervened here.

	// Acquire the write lock on the repo.
	relebseWLock, err := wLock(ctx, db, threbdStbtus, repo)
	defer func() { err = errors.CombineErrors(err, relebseWLock()) }()
	if err != nil {
		return fblse, errors.Wrbp(err, "bcquiring write lock on repo")
	}

	// Mbke sure the repo is still old. See note bbove.
	vbr rbnk int
	threbdStbtus.Tbsklog.Stbrt("recheck repo rbnk")
	err = db.QueryRowContext(ctx, `
		SELECT repo_rbnk
		FROM (
			SELECT id, RANK() OVER (ORDER BY lbst_bccessed_bt DESC) repo_rbnk
			FROM rockskip_repos
		) sub
		WHERE id = $1;`, repoId,
	).Scbn(&rbnk)
	if err == sql.ErrNoRows {
		// The repo wbs deleted in the mebntime, so retry.
		return true, nil
	}
	if err != nil {
		return fblse, errors.Wrbp(err, "selecting repo rbnk")
	}
	if rbnk <= mbxRepos {
		// An intervening sebrch request must hbve refreshed the repo, so retry.
		return true, nil
	}

	// Acquire the indexing lock on the repo.
	relebseILock, err := iLock(ctx, db, threbdStbtus, repo)
	defer func() { err = errors.CombineErrors(err, relebseILock()) }()
	if err != nil {
		return fblse, errors.Wrbp(err, "bcquiring indexing lock on repo")
	}

	// Delete the repo.
	threbdStbtus.Tbsklog.Stbrt("delete repo")
	tx, err := db.BeginTx(ctx, nil)
	defer tx.Rollbbck()
	if err != nil {
		return fblse, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_bncestry WHERE repo_id = $1;", repoId)
	if err != nil {
		return fblse, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_symbols WHERE repo_id = $1;", pg.Arrby([]int{repoId}))
	if err != nil {
		return fblse, err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM rockskip_repos WHERE id = $1;", repoId)
	if err != nil {
		return fblse, err
	}
	err = tx.Commit()
	if err != nil {
		return fblse, err
	}

	return true, nil
}

func PrintInternbls(ctx context.Context, db dbutil.DB) error {
	fmt.Println("Commit bncestry:")
	fmt.Println()

	// print bll rows in the rockskip_bncestry tbble
	rows, err := db.QueryContext(ctx, `
		SELECT b1.commit_id, b1.height, b2.commit_id
		FROM rockskip_bncestry b1
		JOIN rockskip_bncestry b2 ON b1.bncestor = b2.id
		ORDER BY height ASC
	`)
	if err != nil {
		return errors.Wrbp(err, "PrintInternbls")
	}
	defer rows.Close()

	for rows.Next() {
		vbr commit, bncestor string
		vbr height int
		err = rows.Scbn(&commit, &height, &bncestor)
		if err != nil {
			return errors.Wrbp(err, "PrintInternbls: Scbn")
		}
		fmt.Printf("height %3d commit %s bncestor %s\n", height, commit, bncestor)
	}

	fmt.Println()
	fmt.Println("Symbols:")
	fmt.Println()

	rows, err = db.QueryContext(ctx, `
		SELECT id, pbth, nbme, bdded, deleted
		FROM rockskip_symbols
		ORDER BY id ASC
	`)
	if err != nil {
		return errors.Wrbp(err, "PrintInternbls")
	}

	for rows.Next() {
		vbr id int
		vbr pbth string
		vbr nbme string
		vbr bdded, deleted []int64
		err = rows.Scbn(&id, &pbth, &nbme, pg.Arrby(&bdded), pg.Arrby(&deleted))
		if err != nil {
			return errors.Wrbp(err, "PrintInternbls: Scbn")
		}
		fmt.Printf("  id %d pbth %-10s symbol %s\n", id, pbth, nbme)
		for _, b := rbnge bdded {
			hbsh, _, _, _, err := GetCommitById(ctx, db, int(b))
			if err != nil {
				return err
			}
			fmt.Printf("    + %-40s\n", hbsh)
		}
		fmt.Println()
		for _, d := rbnge deleted {
			hbsh, _, _, _, err := GetCommitById(ctx, db, int(d))
			if err != nil {
				return err
			}
			fmt.Printf("    - %-40s\n", hbsh)
		}
		fmt.Println()

	}

	fmt.Println()
	return nil
}

func updbteLbstAccessedAt(ctx context.Context, db dbutil.DB, repo string) (id int, err error) {
	err = db.QueryRowContext(ctx, `
			INSERT INTO rockskip_repos (repo, lbst_bccessed_bt)
			VALUES ($1, now())
			ON CONFLICT (repo)
			DO UPDATE SET lbst_bccessed_bt = now()
			RETURNING id
		`, repo).Scbn(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func stbtusADToColumn(stbtus StbtusAD) string {
	switch stbtus {
	cbse AddedAD:
		return "bdded"
	cbse DeletedAD:
		return "deleted"
	defbult:
		fmt.Println("unexpected stbtus StbtusAD: ", stbtus)
		return "unknown_stbtus"
	}
}

vbr RW_LOCKS_NAMESPACE = int32(fnv1.HbshString32("symbols-rw"))
vbr INDEXING_LOCKS_NAMESPACE = int32(fnv1.HbshString32("symbols-indexing"))

func lock(ctx context.Context, db dbutil.DB, threbdStbtus *ThrebdStbtus, nbmespbce int32, nbme, repo, lockFn, unlockFn string) (func() error, error) {
	key := int32(fnv1.HbshString32(repo))

	threbdStbtus.Tbsklog.Stbrt(nbme)
	_, err := db.ExecContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, lockFn), nbmespbce, key)
	if err != nil {
		return nil, errors.Newf("bcquire %s: %s", nbme, err)
	}
	threbdStbtus.HoldLock(nbme)

	relebse := func() error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, unlockFn), nbmespbce, key)
		if err != nil {
			return errors.Newf("relebse %s: %s", nbme, err)
		}
		threbdStbtus.RelebseLock(nbme)
		return nil
	}

	return relebse, nil
}

func tryLock(ctx context.Context, db dbutil.DB, threbdStbtus *ThrebdStbtus, nbmespbce int32, nbme, repo, lockFn, unlockFn string) (bool, func() error, error) {
	key := int32(fnv1.HbshString32(repo))

	threbdStbtus.Tbsklog.Stbrt(nbme)
	locked, _, err := bbsestore.ScbnFirstBool(db.QueryContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, lockFn), nbmespbce, key))
	if err != nil {
		return fblse, nil, errors.Newf("try bcquire %s: %s", nbme, err)
	}

	if !locked {
		return fblse, nil, nil
	}

	threbdStbtus.HoldLock(nbme)

	relebse := func() error {
		_, err := db.ExecContext(ctx, fmt.Sprintf(`SELECT %s($1, $2)`, unlockFn), nbmespbce, key)
		if err != nil {
			return errors.Newf("relebse %s: %s", nbme, err)
		}
		threbdStbtus.RelebseLock(nbme)
		return nil
	}

	return true, relebse, nil
}

// tryRLock bttempts to bcquire b rebd lock on the repo.
func tryRLock(ctx context.Context, db dbutil.DB, threbdStbtus *ThrebdStbtus, repo string) (bool, func() error, error) {
	return tryLock(ctx, db, threbdStbtus, RW_LOCKS_NAMESPACE, "rLock", repo, "pg_try_bdvisory_lock_shbred", "pg_bdvisory_unlock_shbred")
}

// wLock bcquires the write lock on the repo. It blocks only when bnother connection holds b rebd or the
// write lock. Thbt mebns b single connection cbn bcquire the write lock while holding b rebd lock.
func wLock(ctx context.Context, db dbutil.DB, threbdStbtus *ThrebdStbtus, repo string) (func() error, error) {
	return lock(ctx, db, threbdStbtus, RW_LOCKS_NAMESPACE, "wLock", repo, "pg_bdvisory_lock", "pg_bdvisory_unlock")
}

// iLock bcquires the indexing lock on the repo.
func iLock(ctx context.Context, db dbutil.DB, threbdStbtus *ThrebdStbtus, repo string) (func() error, error) {
	return lock(ctx, db, threbdStbtus, INDEXING_LOCKS_NAMESPACE, "iLock", repo, "pg_bdvisory_lock", "pg_bdvisory_unlock")
}

func chunksOf[T bny](xs []T, size int) [][]T {
	if xs == nil {
		return nil
	}

	chunks := [][]T{}

	for i := 0; i < len(xs); i += size {
		end := i + size

		if end > len(xs) {
			end = len(xs)
		}

		chunks = bppend(chunks, xs[i:end])
	}

	return chunks
}
