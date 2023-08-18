package resolvers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

var _ graphqlbackend.ValidateSearchJobQueryResolver = &validateSearchJobQueryResolver{}

type validateSearchJobQueryResolver struct {
}

func (v *validateSearchJobQueryResolver) Query() string {
	//TODO implement me
	panic("implement me")
}

func (v *validateSearchJobQueryResolver) Valid() bool {
	//TODO implement me
	panic("implement me")
}

func (v *validateSearchJobQueryResolver) Errors() *[]string {
	//TODO implement me
	panic("implement me")
}
