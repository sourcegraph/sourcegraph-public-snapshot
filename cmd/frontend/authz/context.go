package authz

import (
	"context"
)

type key int

const bypassPermissionCheckKey key = iota

func BypassPermissionCheckFromContext(ctx context.Context) bool {
	bypass, ok := ctx.Value(bypassPermissionCheckKey).(bool)
	if !ok {
		return false
	}
	return bypass
}

// WithBypassPermissionsCheck sets whether requests associated with this context should bypass the
// permissions check.
//
// 🚨 SECURITY: If the caller invokes this to bypass the permissions check, it is the caller's
// responsbility to verify that permissions are properly enforced.  The only caller of this function
// currently should be the global search code (because it is often prohibitively expensive to check
// permissions for all repositories on the instance). Any further usages of this function should be
// code-reviewed by the security experts on the team.
func WithBypassPermissionsCheck(ctx context.Context, bypass bool) context.Context {
	return context.WithValue(ctx, bypassPermissionCheckKey, bypass)
}
