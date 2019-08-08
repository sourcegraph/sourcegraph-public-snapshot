package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (v *RepositoryResolver) Labels(ctx context.Context, args *graphqlutil.ConnectionArgs) (LabelConnection, error) {
	return LabelsInRepository(ctx, v.ID(), args)
}
