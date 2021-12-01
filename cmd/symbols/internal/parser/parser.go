package parser

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/fetcher"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type Parser interface {
	Parse(ctx context.Context, args types.SearchArgs, paths []string) (<-chan result.Symbol, error)
}

type parser struct {
	parserPool         ParserPool
	repositoryFetcher  fetcher.RepositoryFetcher
	requestBufferSize  int
	numParserProcesses int
	operations         *operations
}

func NewParser(
	parserPool ParserPool,
	repositoryFetcher fetcher.RepositoryFetcher,
	requestBufferSize int,
	numParserProcesses int,
	observationContext *observation.Context,
) Parser {
	return &parser{
		parserPool:         parserPool,
		repositoryFetcher:  repositoryFetcher,
		requestBufferSize:  requestBufferSize,
		numParserProcesses: numParserProcesses,
		operations:         newOperations(observationContext),
	}
}

func (p *parser) Parse(ctx context.Context, args types.SearchArgs, paths []string) (_ <-chan result.Symbol, err error) {
	ctx, traceLog, endObservation := p.operations.parse.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repo", string(args.Repo)),
		log.String("commitID", string(args.CommitID)),
		log.Int("paths", len(paths)),
		log.String("paths", strings.Join(paths, ":")),
	}})
	// NOTE: We call endObservation synchronously within this function when we
	// return an error. Once we get on the success-only path, we install it to
	// run on defer of a background routine, which indicates when the returned
	// symbols channel is closed.

	parseRequestOrErrors := p.repositoryFetcher.FetchRepositoryArchive(ctx, args, paths)
	if err != nil {
		endObservation(1, observation.Args{})
		return nil, errors.Wrap(err, "repositoryFetcher.FetchRepositoryArchive")
	}
	defer func() {
		if err != nil {
			go func() {
				// Drain channel on early exit
				for range parseRequestOrErrors {
				}
			}()
		}
	}()

	var (
		wg                          sync.WaitGroup                                         // concurrency control
		parseRequests               = make(chan fetcher.ParseRequest, p.requestBufferSize) // buffered requests
		symbols                     = make(chan result.Symbol)                             // parsed responses
		totalRequests, totalSymbols uint32                                                 // stats
	)

	defer func() {
		close(parseRequests)

		go func() {
			defer func() {
				endObservation(1, observation.Args{LogFields: []log.Field{
					log.Int("numSymbols", int(totalSymbols)),
				}})
			}()

			wg.Wait()
			close(symbols)
		}()
	}()

	for i := 0; i < p.numParserProcesses; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for parseRequest := range parseRequests {
				_ = p.handleParseRequest(ctx, symbols, parseRequest, &totalSymbols)
			}
		}()
	}

	for v := range parseRequestOrErrors {
		if v.Err != nil {
			return nil, v.Err
		}

		totalRequests++

		select {
		case parseRequests <- v.ParseRequest:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	traceLog(log.Int("numRequests", int(totalRequests)))

	return symbols, nil
}

func (p *parser) handleParseRequest(ctx context.Context, symbols chan<- result.Symbol, parseRequest fetcher.ParseRequest, totalSymbols *uint32) (err error) {
	ctx, traceLog, endObservation := p.operations.handleParseRequest.WithAndLogger(ctx, &err, observation.Args{LogFields: []log.Field{
		log.Int("fileSize", len(parseRequest.Data)),
	}})
	defer endObservation(1, observation.Args{})

	parser, err := p.parserFromPool(ctx)
	if err != nil {
		return err
	}
	traceLog(log.Event("acquired parser from pool"))

	defer func() {
		if err == nil {
			if e := recover(); e != nil {
				err = errors.Errorf("panic: %s", e)
			}
		}

		if err == nil {
			p.parserPool.Done(parser)
		} else {
			// Close parser and return nil to pool, indicating that the next receiver should create a new parser
			log15.Error("Closing failed parser", "error", err)
			parser.Close()
			p.parserPool.Done(nil)
			p.operations.parseFailed.Inc()
		}
	}()

	p.operations.parsing.Inc()
	defer p.operations.parsing.Dec()

	entries, err := parser.Parse(parseRequest.Path, parseRequest.Data)
	if err != nil {
		return errors.Wrap(err, "parser.Parse")
	}
	traceLog(log.Int("numEntries", len(entries)))

	for _, e := range entries {
		if !shouldPersistEntry(e) {
			continue
		}

		symbol := result.Symbol{
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

		select {
		case symbols <- symbol:
			atomic.AddUint32(totalSymbols, 1)

		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}

func (p *parser) parserFromPool(ctx context.Context) (ctags.Parser, error) {
	p.operations.parseQueueSize.Inc()
	defer p.operations.parseQueueSize.Dec()

	parser, err := p.parserPool.Get(ctx)
	if err != nil {
		if err == context.DeadlineExceeded {
			p.operations.parseQueueTimeouts.Inc()
		}
		if err != ctx.Err() {
			err = errors.Wrap(err, "failed to create parser")
		}
	}

	return parser, err
}

func shouldPersistEntry(e *ctags.Entry) bool {
	if e.Name == "" {
		return false
	}

	for _, value := range []string{"__anon", "AnonymousFunction"} {
		if strings.HasPrefix(e.Name, value) || strings.HasPrefix(e.Parent, value) {
			return false
		}
	}

	return true
}
