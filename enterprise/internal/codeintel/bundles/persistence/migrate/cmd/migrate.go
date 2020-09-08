package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/migrate"
	"github.com/sourcegraph/sourcegraph/internal/sqliteutil"
)

const dbsDir = "/mnt/lsif-storage/dbs"

var numWorkers = runtime.GOMAXPROCS(0)

func init() {
	sqliteutil.MustRegisterSqlite3WithPcre()
}

func main() {
	if err := mainErr(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	infos, err := getInfos()
	if err != nil {
		return err
	}

	db, err := sql.Open("postgres", "postgres://sg:sg@localhost:5432")
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	dumpIDs := make(chan int)
	go func() {
		defer close(dumpIDs)

		for _, info := range infos {
			if dumpID, err := strconv.Atoi(info.Name()); err == nil {
				dumpIDs <- dumpID
			}
		}
	}()

	var wg sync.WaitGroup
	errs := make(chan error)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for dumpID := range dumpIDs {
				errs <- migrateWithinTransaction(context.Background(), dumpID, makeDBFilename(dumpID), db)
			}
		}()
	}

	go func() {
		defer close(errs)
		wg.Wait()
	}()

	total := len(infos)
	finished := 0
	for err := range errs {
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}

		finished++
		fmt.Printf("%d of %d complete (%.2f%%)\n", finished, total, float64(finished)/float64(total)*100)
	}

	return nil
}

func getInfos() ([]os.FileInfo, error) {
	infos, err := ioutil.ReadDir(dbsDir)
	if err != nil {
		return nil, err
	}

	sizes := make([]int64, len(infos))
	for i, info := range infos {
		dumpID, err := strconv.Atoi(info.Name())
		if err != nil {
			continue
		}

		stat, err := os.Stat(makeDBFilename(dumpID))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			return nil, err
		}

		sizes[i] = stat.Size()
	}

	sort.Slice(infos, func(i, j int) bool { return sizes[i] < sizes[j] })
	return infos, nil
}

func makeDBFilename(dumpID int) string {
	return filepath.Join(dbsDir, fmt.Sprintf("%d", dumpID), "sqlite.db")
}

func migrateWithinTransaction(ctx context.Context, dumpID int, filename string, db *sql.DB) (err error) {
	start := time.Now()
	fmt.Printf("migrating %s\n", filename)
	defer func() {
		stat, _ := os.Stat(filename)
		fmt.Printf("finished in %s (%.2f MB)\n", time.Since(start), float64(stat.Size())/1024/1024)
	}()

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
