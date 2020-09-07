package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/migrate"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

const dbsDir = "/Users/efritz/.sourcegraph/lsif-storage/dbs"

func init() {
	sqliteutil.MustRegisterSqlite3WithPcre()
}

func main() {
	if err := mainErr(); err != nil {
		panic(err.Error()) // TODO
	}
}

func mainErr() error {
	infos, err := ioutil.ReadDir(dbsDir)
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", "postgres://sg:sg@localhost:5432")
	if err != nil {
		return err
	}

	for _, info := range infos {
		dumpID, err := strconv.Atoi(info.Name())
		if err != nil {
			continue
		}

		if err := migrateWithinTransaction(context.Background(), dumpID, filepath.Join(dbsDir, info.Name(), "sqlite.db"), db); err != nil {
			return err
		}
	}

	return nil
}

func migrateWithinTransaction(ctx context.Context, dumpID int, filename string, db *sql.DB) (err error) {
	fmt.Printf("migrating %s\n", filename)

	tx, err := db.BeginTx(context.Background(), &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		} else if commitErr := tx.Commit(); commitErr != nil {
			err = multierror.Append(err, commitErr)
		}
	}()

	return migrate.Migrate(
		context.Background(),
		dumpID,
		filepath.Join(dbsDir, fmt.Sprintf("%d", dumpID), "sqlite.db"),
		tx,
	)
}
