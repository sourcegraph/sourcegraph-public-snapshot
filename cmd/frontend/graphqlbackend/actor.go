package graphqlbackend

// Actor implements the GraphQL union Actor. Exactly 1 of the fields must be non-nil.
type Actor struct {
	User          *UserResolver
	Org           *OrgResolver
	ExternalActor *ExternalActor
}

func (v Actor) ToUser() (*UserResolver, bool) {
	return v.User, v.User != nil
}

func (v Actor) ToOrg() (*OrgResolver, bool) {
	return v.Org, v.Org != nil
}

func (v Actor) ToExternalActor() (*ExternalActor, bool) {
	return v.ExternalActor, v.ExternalActor != nil
}

// ExternalActor implements the GraphQL type ExternalActor.
type ExternalActor struct {
	Username_    string
	DisplayName_ *string
	URL_         string
}

func (v ExternalActor) Username() string     { return v.Username_ }
func (v ExternalActor) DisplayName() *string { return v.DisplayName_ }
func (v ExternalActor) URL() string          { return v.URL_ }
