package eval

import (
	"context"
	"fmt"
	"strconv"
	"time"

	openfgav1 "github.com/openfga/api/proto/openfga/v1"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/openfga/openfga/internal/condition"
	"github.com/openfga/openfga/internal/condition/metrics"
	"github.com/openfga/openfga/pkg/telemetry"
	"github.com/openfga/openfga/pkg/typesystem"
)

var tracer = otel.Tracer("openfga/internal/condition/eval")

// EvaluateTupleCondition looks at the given tuple's condition and returns an evaluation result for the given context.
// If the tuple doesn't have a condition, it exits early and doesn't create a span.
// If the tuple's condition isn't found in the model it returns an EvaluationError.
func EvaluateTupleCondition(
	ctx context.Context,
	tupleKey *openfgav1.TupleKey,
	typesys *typesystem.TypeSystem,
	context *structpb.Struct,
) (*condition.EvaluationResult, error) {
	tupleCondition := tupleKey.GetCondition()
	conditionName := tupleCondition.GetName()
	if conditionName == "" {
		return &condition.EvaluationResult{
			ConditionMet: true,
		}, nil
	}

	ctx, span := tracer.Start(ctx, "EvaluateTupleCondition", trace.WithAttributes(
		attribute.String("tuple_key", tupleKey.String()),
		attribute.String("condition_name", conditionName)))
	defer span.End()

	start := time.Now()

	evaluableCondition, ok := typesys.GetCondition(conditionName)
	if !ok {
		err := condition.NewEvaluationError(conditionName, fmt.Errorf("condition was not found"))
		telemetry.TraceError(span, err)
		return nil, err
	}

	span.SetAttributes(attribute.String("condition_expression", evaluableCondition.GetExpression()))

	// merge both contexts
	contextFields := []map[string]*structpb.Value{
		{},
	}
	if context != nil {
		contextFields = []map[string]*structpb.Value{context.GetFields()}
	}

	tupleContext := tupleCondition.GetContext()
	if tupleContext != nil {
		contextFields = append(contextFields, tupleContext.GetFields())
	}

	conditionResult, err := evaluableCondition.Evaluate(ctx, contextFields...)
	if err != nil {
		telemetry.TraceError(span, err)
		return nil, err
	}

	metrics.Metrics.ObserveEvaluationDuration(time.Since(start))
	metrics.Metrics.ObserveEvaluationCost(conditionResult.Cost)

	span.SetAttributes(attribute.Bool("condition_met", conditionResult.ConditionMet),
		attribute.String("condition_cost", strconv.FormatUint(conditionResult.Cost, 10)),
		attribute.StringSlice("condition_missing_params", conditionResult.MissingParameters),
	)
	return &conditionResult, nil
}
