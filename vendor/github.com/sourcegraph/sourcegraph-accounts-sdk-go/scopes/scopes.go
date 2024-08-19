package scopes

import (
	"regexp"
	"slices"
	"strings"
)

// The list of concrete scopes that can be requested by a client.
const (
	OpenID        Scope = "openid"
	Profile       Scope = "profile"
	Email         Scope = "email"
	OfflineAccess Scope = "offline_access"

	// The list of scopes for governing access of a client to a service. For
	// example, "client.ssc" should only be granted to clients that can retrieve SSC
	// data, etc.
	ClientSSC    Scope = "client.ssc"
	ClientDotcom Scope = "client.dotcom"
)

// Ths list of regular expressions to make sure each part of a scope is
// spec-compliant.
var (
	ServiceRegex    = regexp.MustCompile(`^[a-z_]{1,30}$`)
	PermissionRegex = regexp.MustCompile(`^[a-z_.]{1,215}$`)
	ActionRegex     = regexp.MustCompile(`^(read|write|delete)$`)
)

// Service is a type for the service part of a scope.
type Service string

// The list of registered services that publishes scopes.
const (
	ServiceCodyGateway      Service = "cody_gateway"
	ServiceSAMS             Service = "sams"
	ServiceTelemetryGateway Service = "telemetry_gateway"
	ServiceEnterprisePortal Service = "enterprise_portal"
)

// Action is a type for the action part of a scope.
type Action string

const (
	ActionRead   Action = "read"
	ActionWrite  Action = "write"
	ActionDelete Action = "delete"
)

// Scope is the string literal of a scope.
type Scope string

// ToStrings converts a list of scopes to a list of strings.
func ToStrings(scopes []Scope) []string {
	ss := make([]string, len(scopes))
	for i, scope := range scopes {
		ss[i] = string(scope)
	}
	return ss
}

// ToScopes converts a list of strings to a list of scopes.
func ToScopes(strings []string) Scopes {
	scopes := make([]Scope, len(strings))
	for i, s := range strings {
		scopes[i] = Scope(s)
	}
	return scopes
}

// AllowedScopes is a concrete list of allowed scopes that can be registered by
// a client.
type AllowedScopes []Scope

// Contains returns true if the scope is in the list of allowed scopes. It DOES
// NOT do prefix matching like Strategy to prevent clients registering free-form
// and nonsense scopes.
func (s AllowedScopes) Contains(scope Scope) bool {
	return slices.Contains(s, scope)
}

// ToScope returns a scope string in the format of
// "service::permission::action".
func ToScope(service Service, permission Permission, action Action) Scope {
	return Scope(string(service) + "::" + string(permission) + "::" + string(action))
}

// Permission is a type for the permission part of a scope.
type Permission string

var (
	codyGatewayPermissions = []Permission{
		"flaggedprompts",
	}
	samsPermissions = []Permission{
		"user",
		"user.profile",
		"user.roles",
		"user.metadata",
		"user.metadata.cody",
		"user.metadata.dotcom",

		"session",
	}
	telemetryGatewayPermissions = []Permission{
		"events",
	}
	enterprisePortalPermissions = []Permission{
		PermissionEnterprisePortalSubscription,
		PermissionEnterprisePortalSubscriptionPermission,
		PermissionEnterprisePortalCodyAccess,
	}
)

const (
	// Permissions for Enterprise Portal gRPC service:
	// enterpriseportal.subscriptions.v1.SubscriptionsService

	// PermissionEnterprisePortalSubscription designates permissions for
	// Enteprrise subscriptions.
	PermissionEnterprisePortalSubscription Permission = "subscription"
	// PermissionEnterprisePortalSubscriptionPermission designates permissions
	// for managing permissions on Enterprise subscriptions.
	PermissionEnterprisePortalSubscriptionPermission Permission = "permission.subscription"

	// Permissions for Enterprise Portal gRPC service:
	// enterpriseportal.codyaccess.v1.CodyAccessService

	// PermissionEnterprisePortalCodyAccess designates permissions for Enterprise
	// Cody Access for managed Cody features.
	PermissionEnterprisePortalCodyAccess Permission = "codyaccess"
)

// Allowed returns all allowed scopes for a client. The caller should use
// AllowedScopes.Contains for matching requested scopes.
func Allowed() AllowedScopes {
	// Start with the scopes that are defined in the OAuth and OIDC spec.
	allowed := []Scope{
		OpenID, Profile, Email, OfflineAccess,
		// Legacy scopes that will be replaced in future iterations, e.g.
		//	- "client.ssc" -> "ssc::<permission>::<action>"
		//	- "client.dotcom" -> "dotcom::<permission>::<action>"
		ClientSSC, ClientDotcom,
	}

	// Add full { read, write, delete } actions for all permissions for the given service.
	appendScopes := func(service Service, permissions []Permission) {
		for _, permission := range permissions {
			allowed = append(
				allowed,
				[]Scope{
					ToScope(service, permission, ActionRead),
					ToScope(service, permission, ActionWrite),
					ToScope(service, permission, ActionDelete),
				}...,
			)
		}
	}

	appendScopes(ServiceCodyGateway, codyGatewayPermissions)
	appendScopes(ServiceSAMS, samsPermissions)
	appendScopes(ServiceTelemetryGateway, telemetryGatewayPermissions)
	appendScopes(ServiceEnterprisePortal, enterprisePortalPermissions)
	// 👉 ADD YOUR SCOPES HERE
	return allowed
}

type ParsedScope struct {
	Service    Service
	Permission Permission
	Action     Action
}

// ParseScope parses a scope into its parts. It returns the service, permission,
// action, and a boolean indicating if the scope is valid.
//
// Not using strings.Split and returning a non-pointer type to achieve "0 allocs/op" based on benchmarks:
//
// go test -bench=. -benchmem -cpu=4
//
//	BenchmarkStrategy_Match-4     	 6745492	       156.6 ns/op	       0 B/op	       0 allocs/op
//	BenchmarkStrategy_NoMatch-4   	 7670725	       155.6 ns/op	       0 B/op	       0 allocs/op
func ParseScope(scope Scope) (_ ParsedScope, valid bool) {
	// Special case for builtin OAuth and OIDC scopes that has no alias.
	if scope == OpenID || scope == Email || scope == OfflineAccess {
		return ParsedScope{
			Service:    "",
			Permission: Permission(scope),
			Action:     "",
		}, true
	}

	// Backward compatibility for legacy scopes.
	if scope == ClientSSC || scope == ClientDotcom {
		return ParsedScope{
			Service:    "",
			Permission: Permission(scope),
			Action:     "",
		}, true
	}

	i := strings.Index(string(scope), "::")
	if i == -1 {
		return ParsedScope{}, false
	}
	service := scope[:i]

	i += 2
	j := strings.Index(string(scope[i:]), "::")
	if j == -1 {
		return ParsedScope{}, false
	}
	permission := scope[i : i+j]
	action := scope[i+j+2:]
	if service == "" || permission == "" || action == "" {
		return ParsedScope{}, false // Any of the parts of the scope can't be empty.
	}
	return ParsedScope{
		Service:    Service(service),
		Permission: Permission(permission),
		Action:     Action(action),
	}, true
}

var aliases = map[Scope]Scope{
	Profile: ToScope(ServiceSAMS, "user.profile", ActionRead),
}

// Strategy is a custom scope strategy that matches scopes based on the following rules:
//   - Builtin scopes ("openid", "email", "offline_access") without alias are
//     matched by their exact name.
//   - Any matcher or needle that must have the desired the format,
//     "service::permission::action". Otherwise consider not match (returns false).
//   - A overall match is considered when all "service", "permission", and
//     "action" match (returns true).
//   - The "permission" part of the scope is (conceptually) prefix matching, i.e.
//     "user" matches "user" as well as "user.roles" and "user.metadata".
//
// Full specification of the token scope is available at
// https://handbook.sourcegraph.com/departments/engineering/teams/core-services/sams/token_scope_specification/
//
// NOTE: This function must accept strings to have the type of
// `fosite.ScopeStrategy`.
func Strategy(matcherLiterals []string, needleLiteral string) bool {
	needle := Scope(needleLiteral)
	// Canonicalize some scopes that are being searched for with an alias, e.g. only
	// search for "sams::user.profile::read" instead of worrying about both
	// "sams::user.profile::read" AND "profile".
	if alias, ok := aliases[needle]; ok {
		needle = alias
	}
	needleScope, valid := ParseScope(needle)
	if !valid {
		return false
	}

	for _, matcherLiteral := range matcherLiterals {
		matcher := Scope(matcherLiteral)
		if alias, ok := aliases[matcher]; ok {
			matcher = alias
		}

		// If the matcher is longer than the needle, it is impossible to have a match
		// because the permission is prefixing matching, i.e. the needle needs to be at
		// least the same length as the matcher.
		if len(matcher) > len(needle) {
			continue
		}

		scope, valid := ParseScope(matcher)
		if !valid {
			continue
		}

		// If the service or action do not match, it is pointless to check the
		// permission.
		if scope.Service != needleScope.Service || scope.Action != needleScope.Action {
			continue
		}

		if scope.Permission == needleScope.Permission || strings.HasPrefix(string(needleScope.Permission), string(scope.Permission)+".") {
			return true
		}
	}
	return false
}

// Scopes is a list of scopes.
type Scopes []Scope

// Match returns true if any of the scope in the list matches the target scope
// using Strategy.
func (s Scopes) Match(target Scope) bool {
	return Strategy(ToStrings(s), string(target))
}
