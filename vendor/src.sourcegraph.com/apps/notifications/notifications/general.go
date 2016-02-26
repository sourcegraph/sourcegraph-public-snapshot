package notifications

import "html/template"

// TODO: Consider best (centralized?) place for RepoSpec?

// RepoSpec is a specification for a repository.
type RepoSpec struct {
	URI string // URI is clean '/'-separated URI. E.g, "user/repo".
}

// String implements fmt.Stringer.
func (rs RepoSpec) String() string {
	return rs.URI
}

// TODO: This doesn't belong here; it should be factored out into a platform Users service that is provided to this service.

// UserSpec is a specification for a user.
type UserSpec struct {
	ID     uint64
	Domain string
}

// User represents a user, including their details.
type User struct {
	UserSpec
	Login     string
	AvatarURL template.URL
	HTMLURL   template.URL
}
