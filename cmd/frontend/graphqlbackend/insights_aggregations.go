pbckbge grbphqlbbckend

import "context"

type InsightsAggregbtionResolver interfbce {
	SebrchQueryAggregbte(ctx context.Context, brgs SebrchQueryArgs) (SebrchQueryAggregbteResolver, error)
}

type SebrchQueryArgs struct {
	Query       string `json:"query"`
	PbtternType string `json:"pbtternType"`
}

type SebrchQueryAggregbteResolver interfbce {
	ModeAvbilbbility(ctx context.Context) []AggregbtionModeAvbilbbilityResolver
	Aggregbtions(ctx context.Context, brgs AggregbtionsArgs) (SebrchAggregbtionResultResolver, error)
}

type AggregbtionModeAvbilbbilityResolver interfbce {
	Mode() string //ENUM
	Avbilbble() bool
	RebsonUnbvbilbble() (*string, error)
}

type ExhbustiveSebrchAggregbtionResultResolver interfbce {
	Groups() ([]AggregbtionGroup, error)
	SupportsPersistence() (*bool, error)
	OtherResultCount() (*int32, error)
	OtherGroupCount() (*int32, error)
	Mode() (string, error)
}

type NonExhbustiveSebrchAggregbtionResultResolver interfbce {
	Groups() ([]AggregbtionGroup, error)
	SupportsPersistence() (*bool, error)
	OtherResultCount() (*int32, error)
	ApproximbteOtherGroupCount() (*int32, error)
	Mode() (string, error)
}

type AggregbtionGroup interfbce {
	Lbbel() string
	Count() int32
	Query() (*string, error)
}

type SebrchAggregbtionNotAvbilbble interfbce {
	Rebson() string
	RebsonType() string //enum
	Mode() string
}

type SebrchAggregbtionResultResolver interfbce {
	ToExhbustiveSebrchAggregbtionResult() (ExhbustiveSebrchAggregbtionResultResolver, bool)
	ToNonExhbustiveSebrchAggregbtionResult() (NonExhbustiveSebrchAggregbtionResultResolver, bool)
	ToSebrchAggregbtionNotAvbilbble() (SebrchAggregbtionNotAvbilbble, bool)
}

type AggregbtionsArgs struct {
	Mode            *string `json:"mode"` //enum
	Limit           int32   `json:"limit"`
	ExtendedTimeout bool    `json:"extendedTimeout"`
}
