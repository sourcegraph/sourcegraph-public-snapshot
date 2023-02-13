package uploads

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"google.golang.org/api/iterator"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/redis"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/group"
)

const (
	rankingDefinitionHash = "graphDefs:dev"
	rankingReferenceHash  = "graphRefs:dev"
	batchNumber           = 1000
)

func (s *Service) MapperRankingGraph(ctx context.Context, numRankingRoutines int) error {
	fmt.Println("RUNNING RUNNING RUNNING RUNNING well this is running RUNNING RUNNING RUNNING RUNNING")

	// First time running. There are no uploads
	uploadsExists, err := redisStore.Exists("ranking:uploads:processed").Int()
	if err != nil {
		return nil
	}
	if uploadsExists == 0 {
		return nil
	}

	// First time running. There are uploads so make sure we have a graph
	graphExists, err := redisStore.Exists("ranking:graph:processed").Int()
	if err != nil {
		return nil
	}
	if graphExists == 0 {
		err = redisStore.Set("ranking:graph:processed", 0)
		if err != nil {
			return err
		}
	}

	uploadsProcessed, err := redisStore.Get("ranking:uploads:processed").String()
	if err != nil {
		return err
	}
	fmt.Printf("DEBUG: uploadsProcessed: %s \n", uploadsProcessed)

	graphProcessed, err := redisStore.Get("ranking:graph:processed").String()
	if err != nil {
		return err
	}
	fmt.Printf("DEBUG: graphProcessed: %s \n", graphProcessed)

	if uploadsProcessed == graphProcessed {
		return nil
	}

	referencesByUploadID, err := redisStore.LRange("ranking:references:gold", 0, -1).Strings()
	if err != nil {
		return err
	}

	for _, uid := range referencesByUploadID {
		symbolNames := make([]string, 0)
		start := 0
		stop := batchNumber
		for {
			key := fmt.Sprintf("graph:references:%s", uid)
			references, err := redisStore.LRange(key, start, stop).Strings()
			if err != nil {
				return err
			}
			if len(references) == 0 {
				break
			}

			for _, r := range references {
				s := strings.Split(r, "{!@@!}")
				symbolNames = append(symbolNames, s[1])
			}

			fmt.Printf("DEBUG: batching: %d \n", len(symbolNames))

			definitionPath, err := redisStore.HMGet(rankingDefinitionHash, symbolNames...).Strings()
			if err != nil {
				return err
			}

			countMap := make(map[string]int)

			for _, d := range definitionPath {
				if d != "" {
					countMap[d] = countMap[d] + 1
				}
			}

			fields := make([]string, 0, batchNumber)
			for k := range countMap {
				fields = append(fields, k)
			}
			f, err := redisStore.HMGet("graph:globalfiles:counts", fields...).Strings()
			if err != nil {
				return err
			}

			for i, v := range fields {
				value := f[i]
				if value != "" {
					vx, err := strconv.Atoi(value)
					if err != nil {
						return nil
					}
					countMap[v] = countMap[v] + vx
				}
			}

			redisCountMap := make(map[string]interface{})
			for k, v := range countMap {
				redisCountMap[k] = v
			}

			err = redisStore.HMSet("graph:globalfiles:counts", redisCountMap).Err()
			if err != nil {
				return err
			}

			start = stop + 1
			stop = stop + batchNumber
		}
	}

	countMap, err := redisStore.HGetAll("graph:globalfiles:counts").StringMap()
	if err != nil {
		return err
	}

	err = redisStore.Del("graph:globalfiles:counts")
	if err != nil {
		return err
	}

	for k, v := range countMap {
		fmt.Printf("OHBOY OHBOY OHBOY %s: %s \n", k, v)
	}

	redisStore.Incr("ranking:graph:processed")

	err = s.store.SetGlobalRanks(ctx, countMap)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) SerializeRankingGraph(ctx context.Context, numRankingRoutines int) error {
	// if s.rankingBucket == nil {
	// 	return nil
	// }

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
				// if err := s.serializeAndPersistRankingGraphForUpload(ctx, upload.ID, upload.Repo, upload.Root, upload.ObjectPrefix); err != nil {
				// 	s.logger.Error(
				// 		"Failed to process upload for ranking graph",
				// 		log.Int("id", upload.ID),
				// 		log.String("repo", upload.Repo),
				// 		log.String("root", upload.Root),
				// 		log.Error(err),
				// 	)

				// 	return err
				// }

				if err := s.lsifstore.InsertDefinitionsAndReferencesForRanking(ctx, upload, s.setDefinitionsAndReferencesForUpload); err != nil {
					s.logger.Error(
						"Failed to process upload for ranking graph",
						log.Int("id", upload.ID),
						log.String("repo", upload.Repo),
						log.String("root", upload.Root),
						log.Error(err),
					)
				}

				redisStore.LPush("ranking:references:gold", upload.ID)

				// increment the number of processed uploads
				redisStore.Incr("ranking:uploads:processed")

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

func (s *Service) setDefinitionsAndReferencesForUpload(ctx context.Context, upload store.ExportedUpload, path string, document *scip.Document) error {
	definitions := map[string]interface{}{}
	for _, occ := range document.Occurrences {
		if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) {
			continue
		}

		if scip.SymbolRole_Definition.Matches(occ) {
			fullPath := fmt.Sprintf("%s@@%s@@%s", upload.Repo, upload.Root, path)
			definitions[occ.Symbol] = fullPath
		}
	}

	references := []interface{}{}
	for _, occ := range document.Occurrences {
		if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) {
			continue
		}

		if _, ok := definitions[occ.Symbol]; ok {
			continue
		}
		if !scip.SymbolRole_Definition.Matches(occ) {
			references = append(references, fmt.Sprintf("%s{!@@!}%s", path, occ.Symbol))
		}
	}

	if len(definitions) > 0 {
		_, err := redisStore.HMSet(rankingDefinitionHash, definitions).Int()
		if err != nil {
			return err
		}
	}

	if len(references) > 0 {
		hashKey := fmt.Sprintf("graph:references:%d", upload.ID)
		err := redisStore.LPush(hashKey, references...)
		if err != nil {
			return err
		}
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

var redisStore = redis.RedisKeyValue(redispool.Pool)

func (s *Service) serializeRankingGraphForUpload(
	ctx context.Context,
	id int,
	repo string,
	root string,
	write func(filename string, format string, args ...any) error,
) error {
	documentsDefiningSymbols := map[string][]int64{}
	documentsReferencingSymbols := map[string][]int64{}

	if err := s.lsifstore.ScanDocuments(ctx, id, func(path string, document *scip.Document) error {
		documentID := hash(strings.Join([]string{repo, root, path}, ":"))
		documentMonikers := map[string]map[string]struct{}{}

		for _, occurrence := range document.Occurrences {
			if occurrence.Symbol == "" || scip.IsLocalSymbol(occurrence.Symbol) {
				continue
			}

			if scip.SymbolRole_Definition.Matches(occurrence) {
				if _, ok := documentMonikers[occurrence.Symbol]; !ok {
					documentMonikers[occurrence.Symbol] = map[string]struct{}{}
				}
				documentMonikers[occurrence.Symbol]["definition"] = struct{}{}
				documentsDefiningSymbols[occurrence.Symbol] = append(documentsDefiningSymbols[occurrence.Symbol], documentID)
			} else {
				if _, ok := documentMonikers[occurrence.Symbol]; !ok {
					documentMonikers[occurrence.Symbol] = map[string]struct{}{}
				}
				documentMonikers[occurrence.Symbol]["reference"] = struct{}{}
				documentsReferencingSymbols[occurrence.Symbol] = append(documentsReferencingSymbols[occurrence.Symbol], documentID)
			}
		}

		if err := write("documents.csv", "%d,%s,%s\n", documentID, repo, filepath.Join(root, path)); err != nil {
			return err
		}

		for identifier, monikerTypes := range documentMonikers {
			for monikerType := range monikerTypes {
				if err := write("monikers.csv", "%d,%s,%s:%s\n", documentID, monikerType, "scip", identifier); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return err
	}

	for symbolName, referencingDocumentIDs := range documentsReferencingSymbols {
		for _, definingDocumentID := range documentsDefiningSymbols[symbolName] {
			for _, referencingDocumentID := range referencingDocumentIDs {
				if referencingDocumentID == definingDocumentID {
					continue
				}

				if err := write("references.csv", "%d,%d\n", referencingDocumentID, definingDocumentID); err != nil {
					return err
				}
			}
		}
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
