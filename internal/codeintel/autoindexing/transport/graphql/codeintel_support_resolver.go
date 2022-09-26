package graphql

import (
	"context"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type GitBlobCodeIntelSupportResolver interface {
	SearchBasedSupport(context.Context) (SearchBasedSupportResolver, error)
	PreciseSupport(context.Context) (PreciseSupportResolver, error)
}

type codeIntelSupportResolver struct {
	repo         api.RepoName
	path         string
	autoindexSvc AutoIndexingService
	errTracer    *observation.ErrCollector
}

func NewCodeIntelSupportResolver(autoindexSvc AutoIndexingService, repoName api.RepoName, path string, errTracer *observation.ErrCollector) GitBlobCodeIntelSupportResolver {
	return &codeIntelSupportResolver{
		repo:         repoName,
		path:         path,
		autoindexSvc: autoindexSvc,
		errTracer:    errTracer,
	}
}

func (r *codeIntelSupportResolver) SearchBasedSupport(ctx context.Context) (_ SearchBasedSupportResolver, err error) {
	var (
		ctagsSupported bool
		language       string
	)

	defer func() {
		r.errTracer.Collect(&err,
			log.String("codeIntelSupportResolver.field", "searchBasedSupport"),
			log.String("inferedLanguage", language),
			log.Bool("ctagsSupported", ctagsSupported))
	}()

	ctagsSupported, language, err = r.autoindexSvc.GetSupportedByCtags(ctx, r.path, r.repo)
	if err != nil {
		return nil, err
	}

	if !ctagsSupported {
		return nil, nil
	}

	return NewSearchBasedCodeIntelResolver(language), nil
}

func (r *codeIntelSupportResolver) PreciseSupport(ctx context.Context) (PreciseSupportResolver, error) {
	return NewPreciseCodeIntelSupportResolver(r.path), nil
}
