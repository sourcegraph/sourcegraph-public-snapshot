package ranking

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

func (s *Service) buildFileReferenceGraph(ctx context.Context, repoName api.RepoName) (streamingGraph, error) {
	symbolsByPath, err := s.extractSymbols(ctx, repoName)
	if err != nil {
		return nil, err
	}

	ch := make(chan streamedEdge)

	type searchTermAndPath struct {
		Pattern string
		Paths   []string
	}

	//
	// Current focus: Rotate shit around until it's fast
	//

	searchTermsAndPathsByTerm := map[string]*searchTermAndPath{}
	for path, symbols := range symbolsByPath {
		for _, symbol := range symbols {
			searchTermAndPaths, ok := searchTermsAndPathsByTerm[symbol]
			if !ok {
				pattern := symbol
				searchTermAndPaths = &searchTermAndPath{Pattern: pattern}
				searchTermsAndPathsByTerm[symbol] = searchTermAndPaths
			}

			searchTermAndPaths.Paths = append(searchTermAndPaths.Paths, path)
		}
	}

	searchTermsByPathByExtension := map[string]map[string][]string{}
	for _, searchTermAndPaths := range searchTermsAndPathsByTerm {
		pathsByExtension := map[string][]string{}
		for _, path := range searchTermAndPaths.Paths {
			if extension := filepath.Ext(path); extension != "" {
				if _, ok := pathsByExtension[extension]; !ok {
					pathsByExtension[extension] = nil
				}

				pathsByExtension[extension] = append(pathsByExtension[extension], path)
			}
		}

		for extension, paths := range pathsByExtension {
			if _, ok := searchTermsByPathByExtension[extension]; !ok {
				searchTermsByPathByExtension[extension] = map[string][]string{}
			}

			for _, path := range paths {
				if _, ok := searchTermsByPathByExtension[extension][path]; !ok {
					searchTermsByPathByExtension[extension][path] = nil
				}

				searchTermsByPathByExtension[extension][path] = append(searchTermsByPathByExtension[extension][path], searchTermAndPaths.Pattern)
			}
		}
	}

	pathBySearchTermsByExtension := map[string]map[string]*lazyregexp.Regexp{}
	for extension, searchTermsByPath := range searchTermsByPathByExtension {
		pathBySearchTermsByExtension[extension] = map[string]*lazyregexp.Regexp{}

		for path, searchTerms := range searchTermsByPath {
			quotedSearchTerms := make([]string, 0, len(searchTerms))
			for _, searchTerm := range searchTerms {
				quotedSearchTerms = append(quotedSearchTerms, `(`+regexp.QuoteMeta(searchTerm)+`)`)
			}

			pathBySearchTermsByExtension[extension][path] = lazyregexp.New(`(^|\b)` + strings.Join(quotedSearchTerms, `|`) + `($|\b)`)
		}
	}

	extensions := make([]string, 0, len(searchTermsByPathByExtension))
	for extension := range searchTermsByPathByExtension {
		extensions = append(extensions, extension)
	}

	extractGraphEdges := func(h *tar.Header, content []byte) error {
		extension := filepath.Ext(h.Name)
		pathBySearchTerms := pathBySearchTermsByExtension[extension]

		toSet := map[string]struct{}{}
		for path, terms := range pathBySearchTerms {
			if terms.Match(content) {
				toSet[path] = struct{}{}
			}
		}

		for to := range toSet {
			select {
			case ch <- streamedEdge{from: h.Name, to: to}:
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		return nil
	}

	go func() {
		defer close(ch)

		now := time.Now()
		if err := s.forEachFileInArchive(ctx, repoName, extensions, extractGraphEdges); err != nil {
			select {
			case ch <- streamedEdge{err: err}:
			case <-ctx.Done():
				return
			}
		}

		fmt.Printf("Done for %s in %s\n", repoName, time.Since(now))
	}()

	return &graphStreamer{ch: ch}, nil
}

func (s *Service) extractSymbols(ctx context.Context, repoName api.RepoName) (map[string][]string, error) {
	headCommit, ok, err := s.gitserverClient.HeadFromName(ctx, repoName)
	if !ok {
		return nil, nil
	}

	results, err := s.symbolsClient.Search(ctx, search.SymbolsParameters{
		Repo:     repoName,
		CommitID: api.CommitID(headCommit),
		Query:    "",
		First:    10000,
	})
	if err != nil {
		return nil, err
	}

	symbolsByPath := map[string][]string{}
	for _, result := range results {
		if _, ok := symbolsByPath[result.Path]; !ok {
			symbolsByPath[result.Path] = nil
		}

		symbolsByPath[result.Path] = append(symbolsByPath[result.Path], result.Name)
	}

	return symbolsByPath, nil
}

// NOTE: we only look at files under 10,000 characters for demo
func (s *Service) forEachFileInArchive(ctx context.Context, repoName api.RepoName, extensions []string, callback func(h *tar.Header, content []byte) error) error {
	pathspecs := make([]gitdomain.Pathspec, 0, len(extensions))
	for _, extension := range extensions {
		pathspecs = append(pathspecs, gitdomain.PathspecSuffix(extension))
	}

	r, err := s.gitserverClient.ArchiveReader(ctx, authz.DefaultSubRepoPermsChecker, repoName, gitserver.ArchiveOptions{
		Treeish:   "HEAD",
		Format:    "tar",
		Pathspecs: pathspecs,
	})
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()

	cap := int64(10e4)
	buf := make([]byte, cap)

	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}
		if h.FileInfo().IsDir() || strings.Contains(h.Name, "vendor") || h.Size >= cap {
			continue
		}

		if _, err := io.ReadFull(tr, buf[:h.Size]); err != nil {
			return err
		}
		if err := callback(h, buf[:h.Size]); err != nil {
			return err
		}
	}

	return nil
}
