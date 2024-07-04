package enterpriseportal

import "github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

func ReadScopes() scopes.Scopes {
	return scopes.Scopes{
		scopes.ToScope(
			scopes.ServiceEnterprisePortal,
			scopes.PermissionEnterprisePortalSubscription,
			scopes.ActionRead,
		),
		scopes.ToScope(
			scopes.ServiceEnterprisePortal,
			scopes.PermissionEnterprisePortalCodyAccess,
			scopes.ActionRead,
		),
	}
}

func WriteScopes() scopes.Scopes {
	return scopes.Scopes{
		scopes.ToScope(
			scopes.ServiceEnterprisePortal,
			scopes.PermissionEnterprisePortalSubscription,
			scopes.ActionWrite,
		),
		scopes.ToScope(
			scopes.ServiceEnterprisePortal,
			scopes.PermissionEnterprisePortalCodyAccess,
			scopes.ActionWrite,
		),
	}
}
