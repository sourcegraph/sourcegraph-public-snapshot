package uploads

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

func (s *Service) SerializeRankingGraph(ctx context.Context, numRankingRoutines int) error {
	if s.rankingBucket == nil {
		return nil
	}

	uploads, err := s.store.GetUploadsForRanking(ctx, rankingGraphKey, "ranking", rankingGraphBatchSize)
	if err != nil {
		return err
	}

	g := group.New().WithContext(ctx)

	sharedUploads := make(chan store.ExportedUpload, len(uploads))
	for _, upload := range uploads {
		sharedUploads <- upload
	}
	close(sharedUploads)

	for i := 0; i < numRankingRoutines; i++ {
		g.Go(func(ctx context.Context) error {
			for upload := range sharedUploads {
				if err := s.serializeAndPersistRankingGraphForUpload(ctx, upload.ID, upload.Repo, upload.Root, upload.ObjectPrefix); err != nil {
					s.logger.Error(
						"Failed to process upload for ranking graph",
						log.Int("id", upload.ID),
						log.String("repo", upload.Repo),
						log.String("root", upload.Root),
						log.Error(err),
					)

					return err
				}

				s.logger.Info(
					"Processed upload for ranking graph",
					log.Int("id", upload.ID),
					log.String("repo", upload.Repo),
					log.String("root", upload.Root),
				)
				s.operations.numUploadsRead.Inc()
			}

			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (s *Service) VacuumRankingGraph(ctx context.Context) error {
	if s.rankingBucket == nil {
		return nil
	}

	numDeleted, err := s.store.ProcessStaleExportedUploads(ctx, rankingGraphKey, rankingGraphDeleteBatchSize, func(ctx context.Context, objectPrefix string) error {
		if objectPrefix == "" {
			// Special case: we haven't backfilled some data on dotcom yet
			return nil
		}

		objects := s.rankingBucket.Objects(ctx, &storage.Query{
			Prefix: objectPrefix,
		})
		for {
			attrs, err := objects.Next()
			if err != nil {
				if err == iterator.Done {
					break
				}

				return err
			}

			if err := s.rankingBucket.Object(attrs.Name).Delete(ctx); err != nil {
				return err
			}

			s.operations.numBytesDeleted.Add(float64(attrs.Size))
		}

		return nil
	})
	if err != nil {
		return err
	}

	s.operations.numStaleRecordsDeleted.Add(float64(numDeleted))
	return nil
}

const maxBytesPerObject = 1024 * 1024 * 1024 // 1GB

func (s *Service) serializeAndPersistRankingGraphForUpload(
	ctx context.Context,
	id int,
	repo string,
	root string,
	objectPrefix string,
) (err error) {
	writers := map[string]*gcsObjectWriter{}
	defer func() {
		for _, wc := range writers {
			if closeErr := wc.Close(); closeErr != nil {
				err = errors.Append(err, closeErr)
			}
		}
	}()

	return s.serializeRankingGraphForUpload(ctx, id, repo, root, func(filename string, format string, args ...any) error {
		path := fmt.Sprintf("%s/%s", objectPrefix, filename)

		ow, ok := writers[path]
		if !ok {
			handle := s.rankingBucket.Object(path)
			if err := handle.Delete(ctx); err != nil && err != storage.ErrObjectNotExist {
				return err
			}

			wc := handle.NewWriter(ctx)
			ow = &gcsObjectWriter{
				Writer:  bufio.NewWriter(wc),
				c:       wc,
				written: 0,
			}
			writers[path] = ow
		}

		if n, err := io.Copy(ow, strings.NewReader(fmt.Sprintf(format, args...))); err != nil {
			return err
		} else {
			ow.written += n
			s.operations.numBytesUploaded.Add(float64(n))

			if ow.written > maxBytesPerObject {
				return errors.Newf("CSV output exceeds max bytes (%d)", maxBytesPerObject)
			}
		}

		return nil
	})
}

type gcsObjectWriter struct {
	*bufio.Writer
	c       io.Closer
	written int64
}

func (b *gcsObjectWriter) Close() error {
	return errors.Append(b.Flush(), b.c.Close())
}

type countAndPath struct {
	documentID int64
	count      int
}

func (s *Service) serializeRankingGraphForUpload(
	ctx context.Context,
	id int,
	repo string,
	root string,
	write func(filename string, format string, args ...any) error,
) error {
	localPathLookup := map[string]int64{}
	definitionIDs := map[precise.ID]countAndPath{}

	if err := s.lsifstore.ScanDocuments(ctx, id, func(path string, ranges map[precise.ID]precise.RangeData) error {
		id := hash(strings.Join([]string{repo, root, path}, ":"))
		localPathLookup[path] = id
		for _, r := range ranges {
			definitionIDs[r.DefinitionResultID] = countAndPath{id, definitionIDs[r.DefinitionResultID].count + 1}
		}

		return write("documents.csv", "%d,%s,%s\n", id, repo, filepath.Join(root, path))
	}); err != nil {
		return err
	}

	if err := s.lsifstore.ScanResultChunks(ctx, id, func(idx int, resultChunk precise.ResultChunkData) error {
		for id, countAndPath := range definitionIDs {
			if documentAndRanges, ok := resultChunk.DocumentIDRangeIDs[id]; ok {
				for _, documentAndRange := range documentAndRanges {
					pathID, ok := localPathLookup[resultChunk.DocumentPaths[documentAndRange.DocumentID]]
					if !ok {
						continue
					}

					for i := 0; i < countAndPath.count; i++ {
						if err := write("references.csv", "%d,%d\n", countAndPath.documentID, pathID); err != nil {
							return err
						}
					}
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := s.lsifstore.ScanLocations(ctx, id, func(scheme, identifier, monikerType string, locations []precise.LocationData) error {
		for _, location := range locations {
			pathID, ok := localPathLookup[location.URI]
			if !ok {
				continue
			}

			if err := write("monikers.csv", "%d,%s,%s:%s\n", pathID, monikerType, scheme, identifier); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	if err := write("done", "%s\n", time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	return nil
}

func hash(v string) (h int64) {
	if len(v) == 0 {
		return 0
	}
	for _, r := range v {
		h = 31*h + int64(r)
	}

	return h
}
