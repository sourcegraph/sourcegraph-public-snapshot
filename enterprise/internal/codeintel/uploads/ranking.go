package uploads

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

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
)

func (s *Service) ReduceRankingGraph(ctx context.Context, numRankingRoutines int) error {
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

	fmt.Println("=================== WE ARE IN ==========================")

	referencesByUploadID, err := redisStore.LRange("ranking:references:gold", 0, -1).Strings()
	if err != nil {
		return err
	}

	for _, uid := range referencesByUploadID { //"graph:references:consumable": [006, 007] - 100 for now
		symbolNames := make([]string, 0)
		start := 0
		stop := 1000
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
					// _, err := redisStore.HIncrBy("graph:globalfiles:counts", d, 1).Int()
					// if err != nil {
					// 	return err
					// }
				}
			}

			fields := make([]string, 0, 1000)
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
			stop = stop + 1000
		}
	}

	// for _, r := range references {
	// 	s := strings.Split(r, "{!@@!}")
	// 	symbol := s[1]
	// 	definitionPath, err := redisStore.HGet(rankingDefinitionHash, symbol).String()
	// 	fmt.Printf("DEBUG: definitionPath: %s \n", symbol)
	// 	if err == nil {
	// 		countMap[definitionPath] = countMap[definitionPath] + 1
	// 	} else if err != redis.ErrNil {
	// 		return err
	// 	}
	// }

	// countMap, err := redisStore.HGetAll("graph:globalfiles:counts").StringMap()
	// if err != nil {
	// 	return err
	// }

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

	// dirty goes up by 1
	// set ranking:graph:processed to the one above

	// adjacencyList := make(map[string][]string)
	// referenceEdges, err := redisStore.HGetAll(rankingReferenceHash).StringMap()
	// if err != nil {
	// 	return err
	// }

	// for symbolName, edgePath := range referenceEdges {
	// 	vertexPath, err := redisStore.HGet(rankingDefinitionHash, symbolName).String()
	// 	if err == redis.ErrNil || err == nil {
	// 		adjacencyList[vertexPath] = append(adjacencyList[vertexPath], edgePath)
	// 	} else if err != nil {
	// 		return err
	// 	}
	// }

	// fmt.Print("==========START========== \n")
	// for v, e := range adjacencyList {
	// 	fmt.Printf("HERE HERE vertex: %s, totalRefsByPath: %v \n", v, len(e))
	// }
	// fmt.Print("=========END=========== \n")

	return nil
}

func (s *Service) SerializeRankingGraph(ctx context.Context, numRankingRoutines int) error {
	// fmt.Println("well this is running")
	// if s.rankingBucket == nil {
	// 	return nil
	// }

	uploads, err := s.store.GetUploadsForRanking(ctx, rankingGraphKey, "ranking", rankingGraphBatchSize)
	if err != nil {
		return err
	}

	fmt.Printf("HERE HERE HERE uploads: %v", uploads)

	g := group.New().WithContext(ctx)

	sharedUploads := make(chan store.ExportedUpload, len(uploads))
	for _, upload := range uploads {
		sharedUploads <- upload
	}
	close(sharedUploads)

	for i := 0; i < numRankingRoutines; i++ {
		g.Go(func(ctx context.Context) error {
			for upload := range sharedUploads {
				fmt.Printf("well this is running and I have uploads in here %v \n", upload)
				// if err := s.serializeAndPersistRankingGraphForUpload(ctx, upload.ID, upload.Repo, upload.Root, upload.ObjectPrefix); err != nil {
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
	// objectPrefix string,
) (err error) {
	// fmt.Println("well this is running")
	// writers := map[string]*gcsObjectWriter{}
	// defer func() {
	// 	for _, wc := range writers {
	// 		if closeErr := wc.Close(); closeErr != nil {
	// 			err = errors.Append(err, closeErr)
	// 		}
	// 	}
	// }()

	return s.serializeRankingGraphForUpload(ctx, id, repo, root)
	// return s.serializeRankingGraphForUpload(ctx, id, repo, root, func(filename string, format string, args ...any) error {
	// 	path := fmt.Sprintf("%s/%s", objectPrefix, filename)

	// 	ow, ok := writers[path]
	// 	if !ok {
	// 		handle := s.rankingBucket.Object(path)
	// 		if err := handle.Delete(ctx); err != nil && err != storage.ErrObjectNotExist {
	// 			return err
	// 		}

	// 		wc := handle.NewWriter(ctx)
	// 		ow = &gcsObjectWriter{
	// 			Writer:  bufio.NewWriter(wc),
	// 			c:       wc,
	// 			written: 0,
	// 		}
	// 		writers[path] = ow
	// 	}

	// 	if n, err := io.Copy(ow, strings.NewReader(fmt.Sprintf(format, args...))); err != nil {
	// 		return err
	// 	} else {
	// 		ow.written += n
	// 		s.operations.numBytesUploaded.Add(float64(n))

	// 		if ow.written > maxBytesPerObject {
	// 			return errors.Newf("CSV output exceeds max bytes (%d)", maxBytesPerObject)
	// 		}
	// 	}

	// 	return nil
	// })
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

func (s *Service) populateDefsAndRefs(ctx context.Context, uploadID int, repo, root, path string, document *scip.Document) error {
	definitions := map[string]interface{}{}
	for _, occ := range document.Occurrences {
		if occ.Symbol == "" || scip.IsLocalSymbol(occ.Symbol) {
			continue
		}

		if scip.SymbolRole_Definition.Matches(occ) {
			fullPath := fmt.Sprintf("%s@@%s@@%s", repo, root, path)
			definitions[occ.Symbol] = fullPath
			// err := redisStore.HSet(rankingDefinitionHash, occ.Symbol, fullPath)
			// if err != nil {
			// 	return err
			// }
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
			// hashKey := fmt.Sprintf("graph:references:%d", uploadID)
			// item :=
			references = append(references, fmt.Sprintf("%s{!@@!}%s", path, occ.Symbol))
			// err := redisStore.LPush(hashKey, item)
			// if err != nil {
			// 	return err
			// }
		}
	}

	if len(definitions) > 0 {
		_, err := redisStore.HMSet(rankingDefinitionHash, definitions).Int()
		if err != nil {
			return err
		}
	}

	if len(references) > 0 {
		hashKey := fmt.Sprintf("graph:references:%d", uploadID)
		err := redisStore.LPush(hashKey, references...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) serializeRankingGraphForUpload(
	ctx context.Context,
	id int,
	repo string,
	root string,
	// write func(filename string, format string, args ...any) error,
) error {
	// documentsDefiningSymbols := map[string][]int64{}
	// documentsReferencingSymbols := map[string][]int64{}

	// _ = repo
	// _ = root
	// _ = write

	// fmt.Printf("we've started with this many definitions: %v \n", definitions)
	// fmt.Printf("we've started with this many references: %v \n", references)

	if err := s.lsifstore.ScanDocuments(ctx, id, repo, root, s.populateDefsAndRefs); err != nil {
		return err
	}
	redisStore.LPush("ranking:references:gold", id)
	// redisStore.LPush("ranking:references:consumable", id)

	// increment the number of processed uploads
	redisStore.Incr("ranking:uploads:processed")

	// fmt.Printf("we've ENDED up with this many definitions: %v \n", definitions)
	// fmt.Printf("we've ENDED up with this many references: %v \n", references)

	// for symbolName, referencingDocumentIDs := range documentsReferencingSymbols {
	// 	for _, definingDocumentID := range documentsDefiningSymbols[symbolName] {
	// 		for _, referencingDocumentID := range referencingDocumentIDs {
	// 			if referencingDocumentID == definingDocumentID {
	// 				continue
	// 			}

	// 			if err := write("references.csv", "%d,%d\n", referencingDocumentID, definingDocumentID); err != nil {
	// 				return err
	// 			}
	// 		}
	// 	}
	// }

	// if err := write("done", "%s\n", time.Now().Format(time.RFC3339)); err != nil {
	// 	return err
	// }

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
