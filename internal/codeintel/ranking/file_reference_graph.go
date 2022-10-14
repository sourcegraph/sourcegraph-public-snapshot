package ranking

import (
	"archive/tar"
	"context"
	"io"
	"path/filepath"
	"strings"

	"github.com/cloudflare/ahocorasick"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

func (s *Service) buildFileReferenceGraph(ctx context.Context, repoName api.RepoName) (streamingGraph, error) {
	symbolsByPath, err := s.extractSymbols(ctx, repoName)
	if err != nil {
		return nil, err
	}

	ch := make(chan streamedEdge)

	type searchTermAndPath struct {
		symbol string
		paths  []string
	}

	searchTermsAndPathsByTerm := map[string]*searchTermAndPath{}
	for path, symbols := range symbolsByPath {
		for _, symbol := range symbols {
			searchTermAndPaths, ok := searchTermsAndPathsByTerm[symbol]
			if !ok {
				searchTermAndPaths = &searchTermAndPath{symbol: symbol}
				searchTermsAndPathsByTerm[symbol] = searchTermAndPaths
			}

			searchTermAndPaths.paths = append(searchTermAndPaths.paths, path)
		}
	}

	searchTermsAndPathsByExtension := map[string][]*searchTermAndPath{}
	for _, searchTermAndPaths := range searchTermsAndPathsByTerm {
		pathsByExtension := map[string][]string{}
		for _, path := range searchTermAndPaths.paths {
			if extension := filepath.Ext(path); extension != "" {
				if _, ok := pathsByExtension[extension]; !ok {
					pathsByExtension[extension] = nil
				}

				pathsByExtension[extension] = append(pathsByExtension[extension], path)
			}
		}

		for extension, paths := range pathsByExtension {
			if _, ok := searchTermsAndPathsByExtension[extension]; !ok {
				searchTermsAndPathsByExtension[extension] = nil
			}

			searchTermsAndPathsByExtension[extension] = append(searchTermsAndPathsByExtension[extension], &searchTermAndPath{
				symbol: searchTermAndPaths.symbol,
				paths:  paths,
			})
		}
	}

	symbolsMatcherByExtension := map[string]*ahocorasick.Matcher{}
	for extension, searchTermsAndPaths := range searchTermsAndPathsByExtension {
		var symbols []string
		for _, searchTermAndPaths := range searchTermsAndPaths {
			symbols = append(symbols, searchTermAndPaths.symbol)
		}

		symbolsMatcherByExtension[extension] = ahocorasick.NewStringMatcher(symbols)
	}

	extensions := make([]string, 0, len(searchTermsAndPathsByExtension))
	for extension := range searchTermsAndPathsByExtension {
		extensions = append(extensions, extension)
	}

	extractGraphEdges := func(h *tar.Header, content []byte) error {
		extension := filepath.Ext(h.Name)
		searchTermsAndPaths := searchTermsAndPathsByExtension[extension]
		matcher, ok := symbolsMatcherByExtension[extension]
		if !ok {
			// WHY?
			return nil
		}

		emitted := map[string]struct{}{}
		for _, matchingIndex := range matcher.Match(content) {
			for _, path := range searchTermsAndPaths[matchingIndex].paths {
				if _, ok := emitted[path]; ok {
					continue
				}
				emitted[path] = struct{}{}

				select {
				case ch <- streamedEdge{from: h.Name, to: path}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
		}

		return nil
	}

	go func() {
		defer close(ch)

		if err := s.forEachFileInArchive(ctx, repoName, extensions, extractGraphEdges); err != nil {
			select {
			case ch <- streamedEdge{err: err}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return &graphStreamer{ch: ch}, nil
}

func (s *Service) extractSymbols(ctx context.Context, repoName api.RepoName) (map[string][]string, error) {
	headCommit, ok, err := s.gitserverClient.HeadFromName(ctx, repoName)
	if err != nil {
		return nil, err
	}
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
		if result.Name == "_" {
			continue
		}

		if _, ok := symbolsByPath[result.Path]; !ok {
			symbolsByPath[result.Path] = nil
		}

		symbolsByPath[result.Path] = append(symbolsByPath[result.Path], result.Name)
	}

	return symbolsByPath, nil
}

const maxFileSize = int64(10e4)

// forEachFileInArchive invokes the given callback with a tar header and a byte buffer representing
// that file's contents for each file in the repository with one of the given extensions. The byte
// buffer is re-used on each invocation of the callback, so the use of the buffer must be finished
// prior to the callback's return to ensure a stable read.
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

	buf := make([]byte, maxFileSize)

	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}
		if h.FileInfo().IsDir() || strings.Contains(h.Name, "vendor") || h.Size >= maxFileSize {
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
