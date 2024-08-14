// Package idf computes and stores the inverse document frequency (IDF) of a set of repositories.
//
// TODO(beyang): should probably move this elsewhere
package idf

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var redisCache = rcache.NewWithTTL(redispool.Cache, "idf-index", 10*24*60*60)

func Update(ctx context.Context, repoName api.RepoName) error {
	fmt.Printf("# idf.Update(%v)\n", repoName)

	stats := NewStatsAggregator()

	git := gitserver.NewClient("idf-indexer")
	r, err := git.ArchiveReader(ctx, repoName, gitserver.ArchiveOptions{Treeish: "HEAD", Format: gitserver.ArchiveFormatTar})
	if err != nil {
		return nil
	}

	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Printf("Error reading next tar header: %v", err)
			continue
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Read the first line of the file
		scanner := bufio.NewScanner(tr)
		if scanner.Scan() {
			stats.ProcessDoc(scanner.Text())
		} else if err := scanner.Err(); err != nil {
			log.Printf("Error reading file content: %v", err)
		}
	}

	statsP := stats.EvalProvider()
	statsBytes, err := json.Marshal(statsP)

	log.Printf("# storing stats: %s", string(statsBytes))

	if err != nil {
		return errors.Wrap(err, "idf.Update: failed to marshal IDF table")
	}

	redisCache.Set(fmt.Sprintf("repo:%v", repoName), statsBytes)
	return nil
}

func Get(ctx context.Context, repoName api.RepoName) (*StatsProvider, error) {
	fmt.Printf("# idf.Get(%v)", repoName)
	b, ok := redisCache.Get(fmt.Sprintf("repo:%v", repoName))
	if !ok {
		return nil, nil
	}

	var stats StatsProvider
	if err := json.Unmarshal(b, &stats); err != nil {
		return nil, errors.Wrap(err, "idf.Get: failed to unmarshal IDF table")
	}

	log.Printf("# fetching stats: %v", stats)

	return &stats, nil
}

type StatsAggregator struct {
	TermToDocCt map[string]int
	DoctCt      int
}

func NewStatsAggregator() *StatsAggregator {
	return &StatsAggregator{
		TermToDocCt: make(map[string]int),
	}
}

func (s *StatsAggregator) ProcessDoc(text string) {
	for _, tok := range Tokenize(text) {
		term := strings.ToLower((tok))
		s.TermToDocCt[term]++
	}
	s.DoctCt++
}

func (s *StatsAggregator) EvalProvider() StatsProvider {
	idf := make(map[string]float32)
	for term, docCt := range s.TermToDocCt {
		idf[term] = float32(math.Log(float64(s.DoctCt) / (1.0 + float64(docCt))))
	}
	return StatsProvider{IDF: idf}
}

type StatsProvider struct {
	IDF map[string]float32
}

func (s *StatsProvider) GetIDF(term string) float32 {
	return s.IDF[strings.ToLower(term)]
}
