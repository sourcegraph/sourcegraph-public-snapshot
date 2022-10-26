package codenav

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/precise"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (s *Service) SerializeRankingGraph(
	ctx context.Context,
) error {
	if !envvar.SourcegraphDotComMode() && os.Getenv("ENABLE_EXPERIMENTAL_RANKING") == "" {
		return nil
	}
	if s.rankingBucket == nil {
		s.logger.Warn("No ranking bucket is configured")
		return nil
	}

	uploads, err := s.store.GetUploadsForRanking(ctx, rankingGraphKey, rankingGraphBatchSize)
	if err != nil {
		return err
	}

	for _, upload := range uploads {
		if err := s.serializeAndPersistRankingGraphForUpload(ctx, upload.ID, upload.Repo, upload.Root); err != nil {
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
}

const maxBytesPerObject = 1024 * 1024 * 1024 // 1GB

func (s *Service) serializeAndPersistRankingGraphForUpload(
	ctx context.Context,
	id int,
	repo string,
	root string,
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
		path := filepath.Join("ranking", rankingGraphKey, strconv.Itoa(id), filename)

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
