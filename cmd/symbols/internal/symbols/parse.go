package symbols

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/pkg/ctags"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/symbols/protocol"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// startParsers starts the parser process pool.
func (s *Service) startParsers() error {
	n := s.NumParserProcesses
	if n == 0 {
		n = runtime.GOMAXPROCS(0)
	}

	s.parsers = make(chan ctags.Parser, n)
	for i := 0; i < n; i++ {
		parser, err := s.NewParser()
		if err != nil {
			return errors.Wrap(err, "NewParser")
		}
		s.parsers <- parser
	}
	return nil
}

func (s *Service) parseUncached(ctx context.Context, repo api.RepoName, commitID api.CommitID) (<-chan protocol.Symbol, <-chan error, error) {
	errChan := make(chan error, 1)
	parseRequests, fetchErrChan, err := s.fetchRepositoryArchive(ctx, repo, commitID)
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	var (
		mu  sync.Mutex // protects err
		wg  sync.WaitGroup
		sem = make(chan struct{}, runtime.NumCPU())
		symbols = make(chan protocol.Symbol, runtime.NumCPU())
	)
	totalParseRequests := 0
	go func() {
		for req := range parseRequests {
			totalParseRequests++
			if ctx.Err() != nil {
				// Drain parseRequests
				go func() {
					for range parseRequests {
					}
				}()
				errChan <- ctx.Err()
				break
			}
			sem <- struct{}{}
			wg.Add(1)
			go func(req parseRequest) {
				defer func() {
					wg.Done()
					<-sem
				}()
				entries, parseErr := s.parse(ctx, req)
				if parseErr != nil && parseErr != context.Canceled && parseErr != context.DeadlineExceeded {
					if err == nil {
						mu.Lock()
						err = errors.Wrap(parseErr, fmt.Sprintf("parse repo %s commit %s path %s", repo, commitID, req.path))
						mu.Unlock()
					}
					cancel()
					log15.Error("Error parsing symbols.", "repo", repo, "commitID", commitID, "path", req.path, "dataSize", len(req.data), "error", parseErr)
				}
				if len(entries) > 0 {
					for _, e := range entries {
						if e.Name == "" || strings.HasPrefix(e.Name, "__anon") || strings.HasPrefix(e.Parent, "__anon") || strings.HasPrefix(e.Name, "AnonymousFunction") || strings.HasPrefix(e.Parent, "AnonymousFunction") {
							continue
						}
						symbols <- entryToSymbol(e)
					}
				}
			}(req)
		}
		wg.Wait()
		close(symbols)
		close(errChan)
	}()

	return symbols, merge(fetchErrChan, errChan), nil
}

func merge(cs ...<-chan error) <-chan error {
	out := make(chan error)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		go func(c <-chan error) {
			for v := range c {
				out <- v
			}
			wg.Done()
		}(c)
	}
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

// parse gets a parser from the pool and uses it to satisfy the parse request.
func (s *Service) parse(ctx context.Context, req parseRequest) (entries []ctags.Entry, err error) {
	parseQueueSize.Inc()

	select {
	case <-ctx.Done():
		parseQueueSize.Dec()
		if ctx.Err() == context.DeadlineExceeded {
			parseQueueTimeouts.Inc()
		}
		return nil, ctx.Err()
	case parser, ok := <-s.parsers:
		parseQueueSize.Dec()

		if !ok {
			return nil, nil
		}

		if parser == nil {
			// The parser failed for some previous receiver (who returned a nil parser to the channel). Try
			// creating a parser.
			var err error
			parser, err = s.NewParser()
			if err != nil {
				return nil, err
			}
		}

		defer func() {
			if err == nil {
				if e := recover(); e != nil {
					err = fmt.Errorf("panic: %s", e)
				}
			}
			if err == nil {
				// Return parser to pool.
				s.parsers <- parser
			} else {
				// Close parser and return nil to pool, indicating that the next receiver should create a new
				// parser.
				log15.Error("Closing failed parser and creating a new one.", "path", req.path, "error", err)
				parseFailed.Inc()
				parser.Close()
				s.parsers <- nil
			}
		}()
		parsing.Inc()
		defer parsing.Dec()
		return parser.Parse(req.path, req.data)
	}
}

func entryToSymbol(e ctags.Entry) protocol.Symbol {
	return protocol.Symbol{
		Name:        e.Name,
		Path:        e.Path,
		Line:        e.Line,
		Kind:        e.Kind,
		Language:    e.Language,
		Parent:      e.Parent,
		ParentKind:  e.ParentKind,
		Signature:   e.Signature,
		Pattern:     e.Pattern,
		FileLimited: e.FileLimited,
	}
}

var (
	parsing = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "symbols",
		Subsystem: "parse",
		Name:      "parsing",
		Help:      "The number of parse jobs currently running.",
	})
	parseQueueSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "symbols",
		Subsystem: "parse",
		Name:      "parse_queue_size",
		Help:      "The number of parse jobs enqueued.",
	})
	parseQueueTimeouts = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "symbols",
		Subsystem: "parse",
		Name:      "parse_queue_timeouts",
		Help:      "The total number of parse jobs that timed out while enqueued.",
	})
	parseFailed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "symbols",
		Subsystem: "parse",
		Name:      "parse_failed",
		Help:      "The total number of parse jobs that failed.",
	})
)

func init() {
	prometheus.MustRegister(parsing)
	prometheus.MustRegister(parseQueueSize)
	prometheus.MustRegister(parseFailed)
}
