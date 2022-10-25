package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"database/sql"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func main() {
	if err := mainErr(context.Background()); err != nil {
		panic(err.Error())
	}
}

var (
	frontendDSN  = os.Getenv("SUPER_SECRET_FRONTEND_DSN")
	codeIntelDSN = os.Getenv("SUPER_SECRET_CODEINTEL_DSN")
	scratchPath  = "scratch"
	bucketName   = "lsif-pagerank-experiment"
	progressFile = "progress.json"
)

func mainErr(ctx context.Context) (err error) {
	logger := log.Scoped("", "")
	frontendStore, err := initStore(frontendDSN, "frontend", logger)
	if err != nil {
		return err
	}
	codeIntelStore, err := initStore(codeIntelDSN, "codeintel", logger)
	if err != nil {
		return err
	}
	client, err := storage.NewClient(ctx)
	if err != nil {
		return err
	}
	bucket := client.Bucket(bucketName)

	if err := os.RemoveAll(scratchPath); err != nil {
		return err
	}

	contents, err := os.ReadFile(progressFile)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		contents = []byte(`{}`)
	}
	var progress jsonProgress
	if err := json.Unmarshal(contents, &progress); err != nil {
		return err
	}
	if progress.Done {
		fmt.Printf("Already complete.\n")
		return nil
	}

	n := 1 // runtime.GOMAXPROCS(0)
	g := group.New().WithErrors().WithContext(ctx)
	gcsWriteQueue := make(chan serializableProgress, n)
	progressWriteQueue := make(chan blockingSerializableProgress, n)

	g.Go(func(ctx context.Context) error {
		defer close(gcsWriteQueue)

	loop:
		for !progress.Done {
			id, nextBefore, out, err := processNextUpload(ctx, frontendStore, codeIntelStore, progress.Before)
			if err != nil {
				return err
			}

			progress.Before = nextBefore
			if nextBefore == 0 {
				progress.Done = true
			}

			for {
				select {
				case gcsWriteQueue <- serializableProgress{
					id:       id,
					progress: progress,
					out:      out,
				}:
					continue loop
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				fmt.Printf("Stuck sending to GCS write queue\n")
				time.Sleep(time.Millisecond * 500)
			}
		}

		return nil
	})

	g.Go(func(ctx context.Context) error {
		defer close(progressWriteQueue)

		for serializableProgress := range gcsWriteQueue {
			done := make(chan error, 1)
		selectLoop:
			for {
				select {
				case progressWriteQueue <- blockingSerializableProgress{
					serializableProgress: serializableProgress,
					done:                 done,
				}:
					break selectLoop
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				fmt.Printf("Stuck sending to progress write queue\n")
				time.Sleep(time.Millisecond * 500)
			}

			g.Go(func(ctx context.Context) (err error) {
				defer close(done)
				defer func() { done <- err }()

				for {
					select {
					case err := <-serializableProgress.out:
						if err != nil {
							return err
						}

						return moveToGCS(ctx, bucket, serializableProgress.id)
					case <-ctx.Done():
						return ctx.Err()
					default:
					}
					fmt.Printf("Stuck waiting for progress.out\n")
					time.Sleep(time.Millisecond * 500)
				}
			})
		}

		return nil
	})

	g.Go(func(ctx context.Context) error {
		for serializableProgress := range progressWriteQueue {
			for {
				select {
				case err := <-serializableProgress.done:
					if err != nil {
						return err
					}
					serialized, err := json.Marshal(serializableProgress.progress)
					if err != nil {
						return err
					}

					if err := os.WriteFile(progressFile, serialized, os.ModePerm); err != nil {
						return err
					}
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				fmt.Printf("Stuck waiting for progress.done\n")
				time.Sleep(time.Millisecond * 500)
			}
		}

		return nil
	})

	return g.Wait()
}

func moveToGCS(ctx context.Context, bucket *storage.BucketHandle, id int) (err error) {
	start := time.Now()
	fmt.Printf("Uploading %d...\n", id)
	defer func() { fmt.Printf("Done uploading %d (%s)\n", id, time.Since(start)) }()

	g := group.New().WithErrors().WithContext(ctx)

	for _, filename := range []string{"documents.csv", "references.csv", "monikers.csv"} {
		filename := filename

		g.Go(func(ctx context.Context) error {
			objectName := filepath.Join(fmt.Sprintf("%d", id), filename)
			obj := bucket.Object(objectName)
			if err := obj.Delete(ctx); err != nil {
				if err != storage.ErrObjectNotExist {
					return err
				}
			}
			w := obj.NewWriter(ctx)

			return func() (err error) {
				defer func() {
					if closeErr := w.Close(); closeErr != nil {
						err = errors.Append(err, closeErr)
					}
				}()

				return func() (err error) {
					f, err := os.Open(filepath.Join(scratchPath, objectName))
					if err != nil {
						return err
					}
					defer func() {
						if closeErr := f.Close(); closeErr != nil {
							err = errors.Append(err, closeErr)
						}
					}()

					if _, err := io.Copy(w, f); err != nil {
						return err
					}

					return nil
				}()
			}()
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return os.RemoveAll(filepath.Join(scratchPath, fmt.Sprintf("%d", id)))
}

var sem = make(chan struct{}, runtime.GOMAXPROCS(0))

func processNextUpload(ctx context.Context, frontendStore, codeIntelStore *basestore.Store, before int) (id, nextBefore int, _ <-chan error, _ error) {
	beforeCond := sqlf.Sprintf("")
	if before != 0 {
		beforeCond = sqlf.Sprintf("u.id < %s AND", before)
	}

	uploadMeta, ok, err := scanFirstUploadMeta(frontendStore.Query(ctx, sqlf.Sprintf(`
		SELECT
			u.id,
			r.name,
			u.root
		FROM lsif_uploads u
		JOIN repo r ON r.id = u.repository_id
		WHERE
			%s
			u.id IN (SELECT uvt.upload_id FROM lsif_uploads_visible_at_tip uvt WHERE uvt.is_default_branch) AND
			r.deleted_at IS NULL AND r.blocked IS NULL
		ORDER BY u.id DESC
		LIMIT 1
	`, beforeCond)))
	if err != nil || !ok {
		return 0, 0, nil, err
	}

	root := uploadMeta.root
	if root == "" {
		root = "/"
	}

	select {
	case sem <- struct{}{}:
	case <-ctx.Done():
		return 0, 0, nil, ctx.Err()
	}

	out := make(chan error, 1)
	go func() {
		defer func() { <-sem }()
		defer close(out)
		out <- processUpload(ctx, frontendStore, codeIntelStore, uploadMeta)
	}()

	return uploadMeta.id, uploadMeta.id, out, nil
}

func processUpload(ctx context.Context, frontendStore, codeIntelStore *basestore.Store, uploadMeta uploadMeta) error {
	start := time.Now()
	fmt.Printf("Processing %d...\n", uploadMeta.id)
	defer func() { fmt.Printf("Done processing %d (%s)\n", uploadMeta.id, time.Since(start)) }()

	localPathLookup := map[string]int64{}
	definitionIDs := map[ID]countAndPath{}

	if err := withWriter(uploadMeta.id, "documents.csv", func(f func(format string, args ...any) error) error {
		return runQuery(ctx, codeIntelStore, sqlf.Sprintf(`SELECT path, ranges FROM lsif_data_documents WHERE dump_id = %s`, uploadMeta.id), func(s dbutil.Scanner) error {
			var path string
			var rawRanges []byte
			if err := s.Scan(&path, &rawRanges); err != nil {
				return err
			}

			var document DocumentData
			if err := decode(rawRanges, &document.Ranges); err != nil {
				return err
			}

			id := hash(strings.Join([]string{uploadMeta.repo, uploadMeta.root, path}, ":"))
			localPathLookup[path] = id
			for _, r := range document.Ranges {
				definitionIDs[r.DefinitionResultID] = countAndPath{id, definitionIDs[r.DefinitionResultID].count + 1}
			}

			return f("%d,%s,%s\n", id, uploadMeta.repo, filepath.Join(uploadMeta.root, path))
		})
	}); err != nil {
		return err
	}

	if err := withWriter(uploadMeta.id, "references.csv", func(f func(format string, args ...any) error) error {
		return runQuery(ctx, codeIntelStore, sqlf.Sprintf(`SELECT idx, data FROM lsif_data_result_chunks WHERE dump_id = %s`, uploadMeta.id), func(s dbutil.Scanner) error {
			var idx int
			var rawData []byte
			if err := s.Scan(&idx, &rawData); err != nil {
				return err
			}

			var resultChunk ResultChunkData
			if err := decode(rawData, &resultChunk); err != nil {
				return err
			}

			for id, countAndPath := range definitionIDs {
				if documentAndRanges, ok := resultChunk.DocumentIDRangeIDs[id]; ok {
					for _, documentAndRange := range documentAndRanges {
						pathID, ok := localPathLookup[resultChunk.DocumentPaths[documentAndRange.DocumentID]]
						if !ok {
							continue
						}

						for i := 0; i < countAndPath.count; i++ {
							if err := f("%d,%d\n", countAndPath.documentID, pathID); err != nil {
								return err
							}
						}
					}
				}
			}

			return nil
		})
	}); err != nil {
		return err
	}

	if err := withWriter(uploadMeta.id, "monikers.csv", func(f func(format string, args ...any) error) error {
		return runQuery(ctx, codeIntelStore, sqlf.Sprintf(`SELECT s.scheme, s.identifier, s.data, s.type FROM (SELECT *, 'export' AS type FROM lsif_data_definitions UNION SELECT *, 'import' AS type FROM lsif_data_references) s WHERE s.dump_id = %s`, uploadMeta.id), func(s dbutil.Scanner) error {
			var scheme, identifier, monikerType string
			var rawData []byte
			if err := s.Scan(&scheme, &identifier, &rawData, &monikerType); err != nil {
				return err
			}

			var locations []LocationData
			if err := decode(rawData, &locations); err != nil {
				return err
			}

			for _, location := range locations {
				pathID, ok := localPathLookup[location.URI]
				if !ok {
					continue
				}

				if err := f("%d,%s,%s:%s\n", pathID, monikerType, scheme, identifier); err != nil {
					return err
				}
			}

			return nil
		})
	}); err != nil {
		return err
	}

	return nil
}

type uploadMeta struct {
	id   int
	repo string
	root string
}

type countAndPath struct {
	documentID int64
	count      int
}

type jsonProgress struct {
	Before int
	Done   bool
}

type serializableProgress struct {
	id       int
	progress jsonProgress
	out      <-chan error
}

type blockingSerializableProgress struct {
	serializableProgress
	done <-chan error
}

var scanFirstUploadMeta = basestore.NewFirstScanner(func(s dbutil.Scanner) (meta uploadMeta, err error) {
	err = s.Scan(&meta.id, &meta.repo, &meta.root)
	return meta, err
})

//
// Codeintel-db type deserialization

type ID string

type DocumentData struct {
	Ranges map[ID]RangeData
}
type RangeData struct {
	DefinitionResultID ID
}

type ResultChunkData struct {
	DocumentPaths      map[ID]string
	DocumentIDRangeIDs map[ID][]DocumentIDRangeID
}

type DocumentIDRangeID struct {
	DocumentID ID
	RangeID    ID
}

type LocationData struct {
	URI string
}

var readers sync.Pool

func init() {
	gob.Register(&DocumentData{})
	gob.Register(&ResultChunkData{})
	gob.Register(&LocationData{})

	readers = sync.Pool{New: func() any { return new(gzip.Reader) }}
}

func decode(data []byte, target any) (err error) {
	if len(data) == 0 {
		return nil
	}

	r := readers.Get().(*gzip.Reader)
	defer readers.Put(r)

	if err := r.Reset(bytes.NewReader(data)); err != nil {
		return err
	}
	defer func() {
		if closeErr := r.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	return gob.NewDecoder(r).Decode(target)
}

//
// Database utils

func initStore(dsn, schemaName string, logger log.Logger) (_ *basestore.Store, err error) {
	db, err := dbconn.ConnectInternal(logger, dsn, "rdf-exporter", schemaName)
	if err != nil {
		return nil, err
	}

	return basestore.NewWithHandle(basestore.NewHandleWithDB(db, sql.TxOptions{})), nil
}

func runQuery(ctx context.Context, store *basestore.Store, query *sqlf.Query, f func(dbutil.Scanner) error) (err error) {
	rows, queryErr := store.Query(ctx, query)
	if queryErr != nil {
		return queryErr
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		if err := f(rows); err != nil {
			return err
		}
	}

	return nil
}

//
// Filesystem utils

func withWriter(uploadID int, filename string, f func(f func(format string, args ...any) error) error) (err error) {
	path := filepath.Join(scratchPath, fmt.Sprintf("%d/%s", uploadID, filename))
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}

	wc, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := wc.Close(); closeErr != nil {
			err = errors.Append(err, closeErr)
		}
	}()

	total := int64(0)

	return f(func(format string, args ...any) error {
		n, err := io.Copy(wc, strings.NewReader(fmt.Sprintf(format, args...)))
		if err != nil {
			return err
		}

		total += n
		if total > 1024*1024*1024 {
			return errors.Newf("%d TOO BIG", uploadID)
		}

		return nil
	})
}

//
// Hash utils

func hash(v string) (h int64) {
	if len(v) == 0 {
		return 0
	}
	for _, r := range v {
		h = 31*h + int64(r)
	}

	return h
}
