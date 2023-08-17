package resolvers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

var _ graphqlbackend.ValidateExhaustiveSearchQueryResolver = &validateExhaustiveSearchQueryResolver{}

type validateExhaustiveSearchQueryResolver struct {
}

func (v *validateExhaustiveSearchQueryResolver) Query() string {
	//TODO implement me
	panic("implement me")
}

func (v *validateExhaustiveSearchQueryResolver) Valid() bool {
	//TODO implement me
	panic("implement me")
}

func (v *validateExhaustiveSearchQueryResolver) Errors() *[]string {
	//TODO implement me
	panic("implement me")
}
