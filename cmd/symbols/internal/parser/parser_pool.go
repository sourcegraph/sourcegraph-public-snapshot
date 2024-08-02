package parser

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/go-ctags"

	"github.com/sourcegraph/sourcegraph/internal/ctags_config"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/languages"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type poolMetrics struct {
	parseQueueSize     prometheus.Gauge
	parseQueueTimeouts prometheus.Counter
}

func newPoolMetrics(observationCtx *observation.Context, metricsNamespace string) *poolMetrics {

	parseQueueSize := prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: metricsNamespace,
		Name:      "codeintel_symbols_parse_queue_size",
		Help:      "The number of parse jobs enqueued.",
	})
	observationCtx.Registerer.MustRegister(parseQueueSize)

	parseQueueTimeouts := prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: metricsNamespace,
		Name:      "codeintel_symbols_parse_queue_timeouts_total",
		Help:      "The total number of parse jobs that timed out while enqueued.",
	})
	observationCtx.Registerer.MustRegister(parseQueueTimeouts)

	return &poolMetrics{
		parseQueueSize:     parseQueueSize,
		parseQueueTimeouts: parseQueueTimeouts,
	}
}

type ParserFactory func(ctags_config.ParserType) (ctags.Parser, error)

type ParserPool struct {
	newParser ParserFactory
	pool      map[ctags_config.ParserType]chan ctags.Parser
	metrics   *poolMetrics
}

var DefaultParserTypes = []ctags_config.ParserType{ctags_config.UniversalCtags, ctags_config.ScipCtags}

func NewParserPool(observationCtx *observation.Context, metricsNamespace string, newParser ParserFactory, numParserProcesses int, parserTypes []ctags_config.ParserType) (*ParserPool, error) {
	pool := make(map[ctags_config.ParserType]chan ctags.Parser)

	if len(parserTypes) == 0 {
		parserTypes = DefaultParserTypes
	}

	// NOTE: We obviously don't make `NoCtags` available in the pool.
	for _, parserType := range parserTypes {
		pool[parserType] = make(chan ctags.Parser, numParserProcesses)
		for range numParserProcesses {
			parser, err := newParser(parserType)
			if err != nil {
				return nil, err
			}
			pool[parserType] <- parser
		}
	}

	parserPool := &ParserPool{
		metrics:   newPoolMetrics(observationCtx, metricsNamespace),
		newParser: newParser,
		pool:      pool,
	}

	return parserPool, nil
}

// Get a parser from the pool. Once this parser is no longer in use, the Done method
// MUST be called with either the parser or a nil value (when countering an error).
// Nil values will be recreated on-demand via the factory supplied when constructing
// the pool. This method always returns a non-nil parser with a nil error value.
//
// This method blocks until a parser is available or the given context is canceled.
func (p *ParserPool) GetParser(ctx context.Context, path string, content []byte) (ctags.Parser, ctags_config.ParserType, error) {

	parserType := p.getParserType(path, content)
	if ctags_config.ParserIsNoop(parserType) {
		return nil, parserType, nil
	}

	parser, err := p.parserFromPool(ctx, parserType)
	if err != nil {
		return nil, parserType, err
	}

	return parser, parserType, nil
}

func (p *ParserPool) Done(parser ctags.Parser, source ctags_config.ParserType) {
	pool := p.pool[source]
	pool <- parser
}

func (p *ParserPool) parserFromPool(ctx context.Context, source ctags_config.ParserType) (ctags.Parser, error) {
	if ctags_config.ParserIsNoop(source) {
		return nil, errors.New("Should not pass Noop ParserType to parserFromPool")
	}

	p.metrics.parseQueueSize.Inc()
	defer p.metrics.parseQueueSize.Dec()

	parser, err := p.get(ctx, source)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			p.metrics.parseQueueTimeouts.Inc()
		}
		// ignore err if context has expired since err is likely due to that
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, errors.Wrap(err, "failed to create parser")
	}

	return parser, err
}

func (p *ParserPool) getParserType(path string, contents []byte) ctags_config.ParserType {
	language, found := languages.GetMostLikelyLanguage(path, string(contents))
	if !found {
		return ctags_config.UnknownCtags
	}

	parserType := GetParserType(language)
	return parserType
}

func (p *ParserPool) get(ctx context.Context, source ctags_config.ParserType) (ctags.Parser, error) {
	if ctags_config.ParserIsNoop(source) {
		return nil, errors.New("NoCtags is not a valid ParserType")
	}

	pool := p.pool[source]

	select {
	case parser := <-pool:
		if parser != nil {
			return parser, nil
		}

		return p.newParser(source)

	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
