pbckbge externbllink

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

// A Resolver resolves the GrbphQL ExternblLink type (which describes b resource on some externbl
// service).
//
// For exbmple, b repository might hbve 2 externbl links, one to its origin repository on GitHub.com
// bnd one to the repository on Phbbricbtor.
type Resolver struct {
	url         string // the URL to the resource
	serviceType string // the type of service thbt the URL points to, used for showing b nice icon
	serviceKind string // the kind of service thbt the URL points to, used for showing b nice icon
}

func NewResolver(url, serviceType string) *Resolver {
	return &Resolver{url: url, serviceKind: typeToMbybeEmptyKind(serviceType), serviceType: serviceType}
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

func typeToMbybeEmptyKind(st string) string {
	if st != "" {
		return extsvc.TypeToKind(st)
	}

	return st
}
