// Package idf computes and stores the inverse document frequency (IDF) of a set of repositories.
package idf

import (
	"archive/tar"
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"path"
	"strings"
	"unicode"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/rcache"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const featureFlagName = "enhanced-index"

var redisCache = rcache.NewWithTTL(redispool.Cache, "idf-index", 10*24*60*60)

func Update(ctx context.Context, logger log.Logger, repoName api.RepoName) error {
	if !featureflag.FromContext(ctx).GetBoolOr(featureFlagName, false) {
		return nil
	}

	stats := NewStatsAggregator()

	git := gitserver.NewClient("idf-indexer")
	r, err := git.ArchiveReader(ctx, repoName, gitserver.ArchiveOptions{Treeish: "HEAD", Format: gitserver.ArchiveFormatTar})
	if err != nil {
		return nil
	}

	permissibleExtensions := map[string]bool{
		".py": true, ".js": true, ".ts": true, ".java": true, ".cpp": true,
		".c": true, ".cs": true, ".go": true, ".rb": true, ".rs": true,
		".php": true, ".html": true, ".css": true, ".scss": true, ".md": true,
		".sh": true, ".swift": true, ".kt": true, ".m": true,
	}

	tr := tar.NewReader(r)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			logger.Error("Error reading next tar header", log.Error(err))
			continue
		}

		// Skip directories
		if header.Typeflag == tar.TypeDir {
			continue
		}

		// Check if the file has a permissible extension
		ext := strings.ToLower(path.Ext(header.Name))

		if !permissibleExtensions[ext] {
			continue
		}

		// Read the first line of the file
		scanner := bufio.NewScanner(tr)
		if scanner.Scan() {
			stats.ProcessDoc(scanner.Text())
		} else if err := scanner.Err(); err != nil {
			logger.Error("Error reading file content", log.Error(err))
		}
	}

	statsP := stats.EvalProvider()
	statsBytes, err := json.Marshal(statsP)
	if err != nil {
		return errors.Wrap(err, "idf.Update: failed to marshal IDF table")
	}

	logger.Info("Storing enhanced index", log.Int("numWords", len(statsP.IDF)), log.Int("numBytes", len(statsBytes)))
	redisCache.Set(fmt.Sprintf("repo:%v", repoName), statsBytes)
	return nil
}

func Get(ctx context.Context, logger log.Logger, repoName api.RepoName) (*StatsProvider, error) {
	b, ok := redisCache.Get(fmt.Sprintf("repo:%v", repoName))
	if !ok {
		return nil, nil
	}

	var stats StatsProvider
	if err := json.Unmarshal(b, &stats); err != nil {
		return nil, errors.Wrap(err, "idf.Get: failed to unmarshal IDF table")
	}

	logger.Info("Fetched enhanced index", log.Int("numWords", len(stats.IDF)))
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

func isValidWord(word string) bool {
	if len(word) < 3 || len(word) > 20 {
		return false
	}
	hasLetter := false
	for _, char := range word {
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			return false
		}
		if unicode.IsLetter(char) {
			hasLetter = true
		}
	}
	return hasLetter
}

func (s *StatsAggregator) ProcessDoc(text string) {
	words := strings.Fields(text)
	for _, word := range words {
		// word = strings.ToLower(word)
		if isValidWord(word) {
			s.TermToDocCt[word]++
		}
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

func (s *StatsProvider) GetTerms() map[string]float32 {
	return s.IDF
}
