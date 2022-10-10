package ranking

import (
	"archive/tar"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"sort"
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

	// TODO - improve symbol extraction (this is major weak sauce)
	pathsBySymbols := map[string][]string{}
	extractSymbols := func(h *tar.Header, content []byte) error {
		matches := symbolPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			pathsBySymbols[match[1]] = append(pathsBySymbols[match[1]], h.Name)
		}

		return nil
	}
	if err := s.forEachGoFileForDemo(ctx, repoName, extractSymbols); err != nil {
		return err
	}

	// a value contains an occurrence defined by its key
	graph := map[string]map[string]struct{}{}
	extractGraphEdges := func(h *tar.Header, content []byte) error {
		for k, ps := range pathsBySymbols {
			if !bytes.Contains(content, []byte(k)) {
				continue
			}

			for _, p := range ps {
				if _, ok := graph[p]; !ok {
					graph[p] = map[string]struct{}{}
				}

				graph[p][h.Name] = struct{}{}
			}
		}

		return nil
	}
	if err := s.forEachGoFileForDemo(ctx, repoName, extractGraphEdges); err != nil {
		return err
	}

	paths := make([]string, 0, len(graph))
	for p := range graph {
		paths = append(paths, p)
	}
	sort.Slice(paths, func(i, j int) bool { return len(graph[paths[i]]) < len(graph[paths[j]]) })

	ranks := map[string][]float64{}
	n := float64(len(paths))
	for i, path := range paths {
		ranks[path] = []float64{1 - float64(i)/n}
	}

	return s.store.SetDocumentRanks(ctx, repoName, ranks)
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
