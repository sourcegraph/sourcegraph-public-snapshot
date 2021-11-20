package parser

import (
	"context"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	nettrace "golang.org/x/net/trace"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
)

type Parser interface {
	Parse(ctx context.Context, repo api.RepoName, commitID api.CommitID, paths []string, callback func(symbol result.Symbol) error) error
}

type parser struct {
	gitserverClient GitserverClient
	parserPool      ParserPool
	fetchSem        chan int
}

func NewParser(
	gitserverClient GitserverClient,
	parserPool ParserPool,
	maximumConcurrentFetches int,
) *parser {
	return &parser{
		gitserverClient: gitserverClient,
		parserPool:      parserPool,
		fetchSem:        make(chan int, maximumConcurrentFetches),
	}
}

func (p *parser) Parse(ctx context.Context, repo api.RepoName, commitID api.CommitID, paths []string, callback func(symbol result.Symbol) error) (err error) {
	span, ctx := ot.StartSpanFromContext(ctx, "parseUncached")
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.LogFields(otlog.Error(err))
		}
		span.Finish()
	}()
	span.SetTag("repo", string(repo))
	span.SetTag("commit", string(commitID))

	tr := nettrace.New("parseUncached", string(repo))
	tr.LazyPrintf("commitID: %s", commitID)

	totalSymbols := 0
	defer func() {
		tr.LazyPrintf("symbols=%d", totalSymbols)
		if err != nil {
			tr.LazyPrintf("error: %s", err)
			tr.SetError()
		}
		tr.Finish()
	}()

	tr.LazyPrintf("fetch")
	parseRequests, errChan, err := fetchRepositoryArchive(ctx, p.gitserverClient, p.fetchSem, repo, commitID, paths)
	tr.LazyPrintf("fetch (returned chans)")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu sync.Mutex // protects symbols and err
		wg sync.WaitGroup
	)
	tr.LazyPrintf("parse")
	totalParseRequests := 0
	for req := range parseRequests {
		totalParseRequests++
		if ctx.Err() != nil {
			// Drain parseRequests
			go func() {
				for range parseRequests {
				}
			}()
			return ctx.Err()
		}
		wg.Add(1)
		go func(req parseRequest) {
			defer wg.Done()
			entries, parseErr := parse(ctx, p.parserPool, req)
			if parseErr != nil && parseErr != context.Canceled && parseErr != context.DeadlineExceeded {
				log15.Error("Error parsing symbols.", "repo", repo, "commitID", commitID, "path", req.path, "dataSize", len(req.data), "error", parseErr)
			}
			if len(entries) > 0 {
				mu.Lock()
				defer mu.Unlock()
				for _, e := range entries {
					if e.Name == "" || strings.HasPrefix(e.Name, "__anon") || strings.HasPrefix(e.Parent, "__anon") || strings.HasPrefix(e.Name, "AnonymousFunction") || strings.HasPrefix(e.Parent, "AnonymousFunction") {
						continue
					}
					totalSymbols++
					err = callback(entryToSymbol(e))
					if err != nil {
						log15.Error("Failed to add symbol", "symbol", e, "error", err)
						return
					}
				}
			}
		}(req)
	}
	wg.Wait()
	tr.LazyPrintf("parse (done) totalParseRequests=%d symbols=%d", totalParseRequests, totalSymbols)

	return <-errChan
}

// parse gets a parser from the pool and uses it to satisfy the parse request.
func parse(ctx context.Context, parserPool ParserPool, req parseRequest) (entries []*ctags.Entry, err error) {
	parseQueueSize.Inc()
	parser, err := parserPool.Get(ctx)
	parseQueueSize.Dec()

	if err != nil {
		if err == context.DeadlineExceeded {
			parseQueueTimeouts.Inc()
		}
		return nil, err
	}

	defer func() {
		if err == nil {
			if e := recover(); e != nil {
				err = errors.Errorf("panic: %s", e)
			}
		}
		if err == nil {
			parserPool.Done(parser)
		} else {
			// Close parser and return nil to pool, indicating that the next receiver should create a new
			// parser.
			log15.Error("Closing failed parser and creating a new one.", "path", req.path, "error", err)
			parseFailed.Inc()
			parser.Close()
			parserPool.Done(nil)
		}
	}()

	parsing.Inc()
	defer parsing.Dec()

	return parser.Parse(req.path, req.data)
}

func entryToSymbol(e *ctags.Entry) result.Symbol {
	return result.Symbol{
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
	parsing = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "symbols_parse_parsing",
		Help: "The number of parse jobs currently running.",
	})
	parseQueueSize = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "symbols_parse_parse_queue_size",
		Help: "The number of parse jobs enqueued.",
	})
	parseQueueTimeouts = promauto.NewCounter(prometheus.CounterOpts{
		Name: "symbols_parse_parse_queue_timeouts",
		Help: "The total number of parse jobs that timed out while enqueued.",
	})
	parseFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "symbols_parse_parse_failed",
		Help: "The total number of parse jobs that failed.",
	})
)
