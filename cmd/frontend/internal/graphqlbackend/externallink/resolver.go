package externallink

import "fmt"

// A Resolver resolves the GraphQL ExternalLink type (which describes a resource on some external
// service).
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
type Resolver struct {
	url         string // the URL to the resource
	serviceType string // the type of service that the URL points to, used for showing a nice icon
}

func (r *Resolver) URL() string { return r.url }
func (r *Resolver) ServiceType() *string {
	if r.serviceType == "" {
		return nil
	}
	return &r.serviceType
}

func (r *Resolver) String() string { return fmt.Sprintf("%s@%s", r.serviceType, r.url) }
