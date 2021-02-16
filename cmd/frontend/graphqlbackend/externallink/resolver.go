package externallink

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// A Resolver resolves the GraphQL ExternalLink type (which describes a resource on some external
// service).
//
// For example, a repository might have 2 external links, one to its origin repository on GitHub.com
// and one to the repository on Phabricator.
type Resolver struct {
	url         string // the URL to the resource
	serviceType string // the type of service that the URL points to, used for showing a nice icon
	serviceKind string // the kind of service that the URL points to, used for showing a nice icon
}

func NewResolver(url, serviceType string) *Resolver {
	return &Resolver{url: url, serviceKind: typeToMaybeEmptyKind(serviceType), serviceType: serviceType}
}

func (r *Resolver) URL() string { return r.url }

func (r *Resolver) ServiceKind() *string {
	if r.serviceKind == "" {
		return nil
	}
	return &r.serviceKind
}

func (r *Resolver) ServiceType() *string {
	if r.serviceType == "" {
		return nil
	}
	return &r.serviceType
}

func (r *Resolver) String() string { return fmt.Sprintf("%s@%s", r.serviceKind, r.url) }

func typeToMaybeEmptyKind(st string) string {
	if st != "" {
		return extsvc.TypeToKind(st)
	}

	return st
}
