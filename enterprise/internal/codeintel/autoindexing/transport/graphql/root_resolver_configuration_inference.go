package graphql

import (
	"context"

	"go.opentelemetry.io/otel/attribute"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func (r *rootResolver) CodeIntelligenceInferenceScript(ctx context.Context) (script string, err error) {
	ctx, _, endObservation := r.operations.codeIntelligenceInferenceScript.With(ctx, &err, observation.Args{})
	defer endObservation(1, observation.Args{})

	return r.autoindexSvc.GetInferenceScript(ctx)
}

func (r *rootResolver) UpdateCodeIntelligenceInferenceScript(ctx context.Context, args *resolverstubs.UpdateCodeIntelligenceInferenceScriptArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateCodeIntelligenceInferenceScript.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("script", args.Script),
	}})
	defer endObservation(1, observation.Args{})

	if err := r.autoindexSvc.SetInferenceScript(ctx, args.Script); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, nil
}
