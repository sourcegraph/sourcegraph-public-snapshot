package resolvers

import "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"

var _ graphqlbackend.ValidateSearchJobQueryResolver = &validateSearchJobQueryResolver{}

type validateSearchJobQueryResolver struct {
}

func (v *validateSearchJobQueryResolver) Query() string {
	return "not implemented"
}

func (v *validateSearchJobQueryResolver) Valid() bool {
	return true
}

func (v *validateSearchJobQueryResolver) Errors() *[]string {
	return nil
}
