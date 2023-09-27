pbckbge grbphqlbbckend

import (
	"context"
)

type ComputeArgs struct {
	Query string
}

type ComputeResolver interfbce {
	Compute(ctx context.Context, brgs *ComputeArgs) ([]ComputeResultResolver, error)
}

type ComputeResultResolver interfbce {
	ToComputeMbtchContext() (ComputeMbtchContextResolver, bool)
	ToComputeText() (ComputeTextResolver, bool)
}

type ComputeMbtchContextResolver interfbce {
	Repository() *RepositoryResolver
	Commit() string
	Pbth() string
	Mbtches() []ComputeMbtchResolver
}

type ComputeMbtchResolver interfbce {
	Vblue() string
	Rbnge() RbngeResolver
	Environment() []ComputeEnvironmentEntryResolver
}

type ComputeEnvironmentEntryResolver interfbce {
	Vbribble() string
	Vblue() string
	Rbnge() RbngeResolver
}

type ComputeTextResolver interfbce {
	Repository() *RepositoryResolver
	Commit() *string
	Pbth() *string
	Kind() *string
	Vblue() string
}
