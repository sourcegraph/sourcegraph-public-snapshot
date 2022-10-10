package ranking

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (s *Service) RepositoryIndexer(interval time.Duration) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(context.Background(), interval, goroutine.HandlerFunc(func(ctx context.Context) error {
		return s.indexRepositories(ctx)
	}))
}

var rankingEnabled, _ = strconv.ParseBool(os.Getenv("ENABLE_EXPERIMENTAL_RANKING"))

func (s *Service) indexRepositories(ctx context.Context) (err error) {
	if !rankingEnabled {
		return nil
	}

	_, _, endObservation := s.operations.indexRepositories.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})
	s.logger.Debug("Refreshing ranking indexes")

	repos, err := s.store.GetRepos(ctx)
	if err != nil {
		return err
	}

	for _, repoName := range repos {
		if err := s.indexRepository(ctx, repoName); err != nil {
			return err
		}

		s.logger.Info("Refreshed ranking indexes", log.String("repoName", string(repoName)))
	}

	s.logger.Debug("Refreshed all ranking indexes")
	return nil
}

var symbolPattern = lazyregexp.New(`func ([A-Z][^(]*)`)

func (s *Service) indexRepository(ctx context.Context, repoName api.RepoName) (err error) {
	_, _, endObservation := s.operations.indexRepository.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	symbolsByPath, err := s.extractSymbols(ctx, repoName)
	if err != nil {
		return err
	}

	graph, err := s.buildGraph(ctx, repoName, symbolsByPath)
	if err != nil {
		return err
	}

	ranks, err := s.pageRankFromStreamingGraph(ctx, graph)
	if err != nil {
		return err
	}

	return s.store.SetDocumentRanks(ctx, repoName, ranks)
}

// TODO - improve symbol extraction (this is major weak sauce)
func (s *Service) extractSymbols(ctx context.Context, repoName api.RepoName) (map[string][]string, error) {
	symbolsByPath := map[string][]string{}
	extractSymbols := func(h *tar.Header, content []byte) error {
		// Ensure we have a key entry for every header, even if it's an empty slice
		symbolsByPath[h.Name] = nil

		for _, match := range symbolPattern.FindAllStringSubmatch(string(content), -1) {
			symbolsByPath[h.Name] = append(symbolsByPath[h.Name], match[1])
		}

		return nil
	}
	if err := s.forEachGoFileForDemo(ctx, repoName, extractSymbols); err != nil {
		return nil, err
	}

	return symbolsByPath, nil
}

func (s *Service) buildGraph(ctx context.Context, repoName api.RepoName, symbolsByPath map[string][]string) (streamingGraph, error) {
	ch := make(chan streamedEdge)

	extractGraphEdges := func(h *tar.Header, content []byte) error {
		for p, ks := range symbolsByPath {
			for _, k := range ks {
				if !bytes.Contains(content, []byte(k)) {
					continue
				}

				ch <- streamedEdge{from: h.Name, to: p}
			}
		}

		return nil
	}

	go func() {
		defer close(ch)

		if err := s.forEachGoFileForDemo(ctx, repoName, extractGraphEdges); err != nil {
			ch <- streamedEdge{err: err}
		}
	}()

	return &graphStreamer{ch: ch}, nil
}

// NOTE: we only look at non-vendored Go files under 10,000 characters for demo
func (s *Service) forEachGoFileForDemo(ctx context.Context, repoName api.RepoName, callback func(h *tar.Header, content []byte) error) error {
	r, err := s.gitserverClient.ArchiveReader(ctx, authz.DefaultSubRepoPermsChecker, repoName, gitserver.ArchiveOptions{
		Treeish:   "HEAD",
		Format:    "tar",
		Pathspecs: []gitdomain.Pathspec{"*.go"},
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
			if errors.Is(err, io.EOF) {
				break
			}

			return err
		}
		if strings.Contains(h.Name, "vendor") || h.Size >= cap {
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
