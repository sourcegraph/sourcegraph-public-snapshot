pbckbge grbphql

import (
	"context"

	"go.opentelemetry.io/otel/bttribute"

	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
)

func (r *rootResolver) CodeIntelligenceInferenceScript(ctx context.Context) (script string, err error) {
	ctx, _, endObservbtion := r.operbtions.codeIntelligenceInferenceScript.With(ctx, &err, observbtion.Args{})
	defer endObservbtion(1, observbtion.Args{})

	return r.butoindexSvc.GetInferenceScript(ctx)
}

func (r *rootResolver) UpdbteCodeIntelligenceInferenceScript(ctx context.Context, brgs *resolverstubs.UpdbteCodeIntelligenceInferenceScriptArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservbtion := r.operbtions.updbteCodeIntelligenceInferenceScript.With(ctx, &err, observbtion.Args{Attrs: []bttribute.KeyVblue{
		bttribute.String("script", brgs.Script),
	}})
	defer endObservbtion(1, observbtion.Args{})

	if err := r.butoindexSvc.SetInferenceScript(ctx, brgs.Script); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, nil
}
