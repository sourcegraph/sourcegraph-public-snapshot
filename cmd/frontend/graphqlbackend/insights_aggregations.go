package graphqlbackend

import "context"

type InsightsAggregationResolver interface {
	SearchQueryAggregate(ctx context.Context, args SearchQueryArgs) (SearchQueryAggregateResolver, error)
}

type SearchQueryArgs struct {
	Query       string `json:"query"`
	PatternType string `json:"patternType"`
}

type SearchQueryAggregateResolver interface {
	ModeAvailability(ctx context.Context) []AggregationModeAvailabilityResolver
	Aggregations(ctx context.Context, args AggregationsArgs) (SearchAggregationResultResolver, error)
}

type AggregationModeAvailabilityResolver interface {
	Mode() string //ENUM
	Available() (bool, error)
	ReasonUnavailable() (*string, error)
}

type ExhaustiveSearchAggregationResultResolver interface {
	Groups() ([]AggregationGroup, error)
	SupportsPersistence() (*bool, error)
	OtherResultCount() (*int32, error)
	OtherGroupCount() (*int32, error)
	Mode() (string, error)
}

type NonExhaustiveSearchAggregationResultResolver interface {
	Groups() ([]AggregationGroup, error)
	SupportsPersistence() (*bool, error)
	OtherResultCount() (*int32, error)
	ApproximateOtherGroupCount() (*int32, error)
	Mode() (string, error)
}

type AggregationGroup interface {
	Label() string
	Count() int32
	Query() (*string, error)
}

type SearchAggregationNotAvailable interface {
	Reason() string
	ReasonType() string //enum
	Mode() string
}

type SearchAggregationResultResolver interface {
	ToExhaustiveSearchAggregationResult() (ExhaustiveSearchAggregationResultResolver, bool)
	ToNonExhaustiveSearchAggregationResult() (NonExhaustiveSearchAggregationResultResolver, bool)
	ToSearchAggregationNotAvailable() (SearchAggregationNotAvailable, bool)
}

type AggregationsArgs struct {
	Mode            *string `json:"mode"` //enum
	Limit           int32   `json:"limit"`
	ExtendedTimeout bool    `json:"extendedTimeout"`
}
