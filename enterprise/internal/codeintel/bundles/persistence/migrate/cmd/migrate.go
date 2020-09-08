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
	start := time.Now()
	defer func() {
		fmt.Printf("Finished migration in %s", time.Since(start))
	}()

	if err := mainErr(); err != nil {
		fmt.Printf("error: %s\n", err)
		os.Exit(1)
	}
}

func mainErr() error {
	infos, totalBytes, err := getInfos()
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

	type result struct {
		DumpID int
		Err    error
	}

	var wg sync.WaitGroup
	errs := make(chan result)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for dumpID := range dumpIDs {
				errs <- result{
					DumpID: dumpID,
					Err:    migrateWithinTransaction(context.Background(), dumpID, makeDBFilename(dumpID), db),
				}
			}
		}()
	}

	go func() {
		defer close(errs)
		wg.Wait()
	}()

	total := len(infos)
	finished := 0
	finishedBytes := int64(0)
	for result := range errs {
		if result.Err != nil {
			fmt.Printf("error: %s\n", err)
		}

		finished++
		finishedBytes += getSize(result.DumpID)

		fmt.Printf("%d of %d bundles complete (%.2f%%)\n", finished, total, float64(finished)/float64(total)*100)
		fmt.Printf("%d of %d bytes complete (%.2f%%)\n", finishedBytes, totalBytes, float64(finishedBytes)/float64(totalBytes)*100)
	}

	return nil
}

func getInfos() ([]os.FileInfo, int64, error) {
	infos, err := ioutil.ReadDir(dbsDir)
	if err != nil {
		return nil, 0, err
	}

	var wg sync.WaitGroup
	ch := make(chan int)
	sizes := make([]int64, len(infos))

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := range ch {
				dumpID, err := strconv.Atoi(infos[i].Name())
				if err != nil {
					continue
				}

				sizes[i] = getSize(dumpID)
			}
		}()
	}

	for i := range infos {
		ch <- i
	}
	close(ch)

	wg.Wait()

	totalBytes := int64(0)
	sizeMap := map[string]int64{}
	for i, size := range sizes {
		sizeMap[infos[i].Name()] = size
		totalBytes += size
	}

	sort.Slice(infos, func(i, j int) bool { return sizeMap[infos[j].Name()] < sizeMap[infos[i].Name()] })
	return infos, totalBytes, nil
}

func migrateWithinTransaction(ctx context.Context, dumpID int, filename string, db *sql.DB) (err error) {
	start := time.Now()

	fmt.Printf("Migrating %s\n", filename)
	defer func() {
		fmt.Printf("Finished in %s (%.2f MB)\n", time.Since(start), float64(getSize(dumpID))/1024/1024)
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
		filename,
		tx,
	)
}

func makeDBFilename(dumpID int) string {
	return filepath.Join(dbsDir, fmt.Sprintf("%d", dumpID), "sqlite.db")
}

func getSize(dumpID int) int64 {
	info, err := os.Stat(makeDBFilename(dumpID))
	if err != nil {
		return 0
	}

	return info.Size()
}
