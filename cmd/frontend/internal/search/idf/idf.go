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
	"time"
	"unicode"

	"github.com/grafana/regexp"

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
var permissibleExtensions = map[string]struct{}{
	".py":      {},
	".js":      {},
	".ts":      {},
	".java":    {},
	".cpp":     {},
	".c":       {},
	".cs":      {},
	".go":      {},
	".rb":      {},
	".rs":      {},
	".php":     {},
	".html":    {},
	".css":     {},
	".scss":    {},
	".md":      {},
	".sh":      {},
	".swift":   {},
	".kt":      {},
	".m":       {},
	".awk":     {},
	".bash":    {},
	".bat":     {},
	".bazel":   {},
	".bzl":     {},
	".cc":      {},
	".cjs":     {},
	".cue":     {},
	".dart":    {},
	".gradle":  {},
	".graphql": {},
	".groovy":  {},
	".hack":    {},
	".hcl":     {},
	".jsx":     {},
	".lua":     {},
	".scala":   {},
	".sql":     {},
	".svelte":  {},
	".tsx":     {},
	".zig":     {},
}

func Update(ctx context.Context, logger log.Logger, repoName api.RepoName) error {
	if !featureflag.FromContext(ctx).GetBoolOr(featureFlagName, false) {
		return nil
	}

	stats := NewStatsAggregator()

	git := gitserver.NewClient("idf-indexer")
	r, err := git.ArchiveReader(ctx, repoName, gitserver.ArchiveOptions{Treeish: "HEAD", Format: gitserver.ArchiveFormatTar, Paths: []string{""}})
	if err != nil {
		return nil
	}

	numFilesProcessed := 0
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
		if _, ok := permissibleExtensions[ext]; !ok {
			continue
		}

		// Read the first line of the file
		scanner := bufio.NewScanner(tr)
		buf := make([]byte, 4*1024)
		scanner.Buffer(buf, 10*1024*1024) // max 10MB file size
		if scanner.Scan() {
			stats.ProcessDoc(scanner.Text())
			numFilesProcessed++
		} else if err := scanner.Err(); err == bufio.ErrTooLong {
			logger.Info("Ignoring file because it was too long", log.String("filename", header.Name))
		} else if err != nil {
			logger.Error("Error reading file content", log.Error(err))
		}
	}

	logger.Info("Processed files for enhanced index", log.String("repoName", string(repoName)), log.Int("numFiles", numFilesProcessed))

	statsP := stats.EvalProvider()
	statsBytes, err := json.Marshal(statsP)
	if err != nil {
		return errors.Wrap(err, "idf.Update: failed to marshal IDF table")
	}

	logger.Info("Storing enhanced index", log.Int("numWords", len(statsP.IDF)), log.Int("numBytes", len(statsBytes)))
	redisCache.Set(fmt.Sprintf("repo:%v", repoName), statsBytes)
	return nil
}

func Get(ctx context.Context, logger log.Logger, repoName api.RepoName) (_ *StatsProvider, err error) {
	if !featureflag.FromContext(ctx).GetBoolOr(featureFlagName, false) {
		return nil, nil
	}

	start := time.Now()
	defer func() {
		if err == nil {
			logger.Info("idf.Get", log.Duration("duration", time.Since(start)))
		} else {
			logger.Error("idf.Get failure", log.Error(err), log.Duration("duration", time.Since(start)))
		}
	}()

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
	if len(word) < 3 || len(word) > 50 {
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
	words := getKeywords(text)
	for _, word := range words {
		if isValidWord(word) {
			s.TermToDocCt[word]++
		}
	}
	s.DoctCt++
}

var keywordRe = regexp.MustCompile(`[^a-zA-Z0-9_]+`)

func getKeywords(s string) []string {

	// Split the string using the regular expression
	fields := keywordRe.Split(s, -1)

	// Filter out empty strings
	var result []string
	for _, field := range fields {
		if field != "" {
			result = append(result, field)
		}
	}

	return result
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
