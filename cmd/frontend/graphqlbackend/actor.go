package graphqlbackend

// Actor implements the GraphQL union Actor. Exactly 1 of the fields must be non-nil.
type Actor struct {
	User *UserResolver
	Org  *OrgResolver
}

func (v Actor) ToUser() (*UserResolver, bool) {
	return v.User, v.User != nil
}

func (v Actor) ToOrg() (*OrgResolver, bool) {
	return v.Org, v.Org != nil
}
