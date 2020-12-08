package externallink

import "fmt"

// A Resolver resolves the GraphQL ExternalLink type (which describes a resource on some external
// service).
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
type Resolver struct {
	url         string // the URL to the resource
	serviceKind string // the kind of service that the URL points to, used for showing a nice icon
}

func NewResolver(url, serviceKind string) *Resolver {
	return &Resolver{url: url, serviceKind: serviceKind}
}

func (r *Resolver) URL() string { return r.url }

func (r *Resolver) ServiceKind() *string {
	if r.serviceKind == "" {
		return nil
	} else {
		return &r.serviceKind
	}

}

func (r *Resolver) String() string { return fmt.Sprintf("%s@%s", r.serviceKind, r.url) }
