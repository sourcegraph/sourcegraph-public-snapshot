package conf

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/schema"
)

const defaultHTTPStrictTransportSecurity = "max-age=31536000" // 1 year

// HTTPStrictTransportSecurity returns the value of the Strict-Transport-Security HTTP header to set.
func HTTPStrictTransportSecurity() string {
	switch v := Get().HttpStrictTransportSecurity.(type) {
	case string:
		return v
	case bool:
		if !v {
			return ""
		}
		return defaultHTTPStrictTransportSecurity
	default:
		return defaultHTTPStrictTransportSecurity
	}
}

// JumpToDefOSSIndexEnabled returns true if JumpToDefOSSIndex experiment is enabled.
func JumpToDefOSSIndexEnabled() bool {
	p := Get().ExperimentalFeatures.JumpToDefOSSIndex
	// default is disabled
	return p == "enabled"
}

// MultipleAuthProvidersEnabled reports whether the "multipleAuthProviders" experiment is enabled.
func MultipleAuthProvidersEnabled() bool { return MultipleAuthProvidersEnabledFromConfig(Get()) }

// MultipleAuthProvidersEnabledFromConfig is like MultipleAuthProvidersEnabled, except it accepts a
// site configuration input value instead of using the current global value.
func MultipleAuthProvidersEnabledFromConfig(c *schema.SiteConfiguration) bool {
	// default is enabled
	return c.ExperimentalFeatures == nil || c.ExperimentalFeatures.MultipleAuthProviders != "disabled"
}

type AccessTokAllow string

const (
	AccessTokensNone  AccessTokAllow = "none"
	AccessTokensAll   AccessTokAllow = "all-users-create"
	AccessTokensAdmin AccessTokAllow = "site-admin-create"
)

// AccessTokensAllow returns whether access tokens are enabled, disabled, or restricted to creation by admin users.
func AccessTokensAllow() AccessTokAllow {
	if Get().AuthDisableAccessTokens {
		return AccessTokensNone
	}

	cfg := Get().AuthAccessTokens
	if cfg == nil {
		return AccessTokensAll
	}
	switch cfg.Allow {
	case "":
		return AccessTokensAll
	case string(AccessTokensAll):
		return AccessTokensAll
	case string(AccessTokensNone):
		return AccessTokensNone
	case string(AccessTokensAdmin):
		return AccessTokensAdmin
	default:
		return AccessTokensNone
	}
}

// EmailVerificationRequired returns whether users must verify an email address before they
// can perform most actions on this site.
//
// It's false for sites that do not have an email sending API key set up.
func EmailVerificationRequired() bool {
	return Get().EmailSmtp != nil
}

// CanSendEmail returns whether the site can send emails (e.g., to reset a password or
// invite a user to an org).
//
// It's false for sites that do not have an email sending API key set up.
func CanSendEmail() bool {
	return Get().EmailSmtp != nil
}

// CanReadEmail tells if an IMAP server is configured and reading email is possible.
func CanReadEmail() bool {
	return Get().EmailImap != nil
}

// HasGitHubDotComToken reports whether there are any personal access tokens configured for
// github.com.
func HasGitHubDotComToken() bool {
	for _, c := range Get().Github {
		u, err := url.Parse(c.Url)
		if err != nil {
			continue
		}
		hostname := strings.ToLower(u.Hostname())
		if (hostname == "github.com" || hostname == "api.github.com") && c.Token != "" {
			return true
		}
	}
	return false
}

// HasGitLabDotComToken reports whether there are any personal access tokens configured for
// github.com.
func HasGitLabDotComToken() bool {
	for _, c := range Get().Gitlab {
		u, err := url.Parse(c.Url)
		if err != nil {
			continue
		}
		hostname := strings.ToLower(u.Hostname())
		if hostname == "gitlab.com" && c.Token != "" {
			return true
		}
	}
	return false
}

// EnabledLangservers returns the langservers that are not disabled.
func EnabledLangservers() []*schema.Langservers {
	all := Get().Langservers
	results := make([]*schema.Langservers, 0, len(all))
	for _, langserver := range all {
		if langserver.Disabled {
			continue
		}
		results = append(results, langserver)
	}
	return results
}

// DeployType tells the deployment type.
func DeployType() string {
	if e := os.Getenv("DEPLOY_TYPE"); e != "" {
		return e
	}
	// Default to datacenter so that every Kubernetes deployment
	// doesn't need to be configured with DEPLOY_TYPE.
	return "datacenter"
}

// IsDataCenter tells if the given deployment type is "datacenter".
func IsDataCenter(deployType string) bool {
	return deployType == "datacenter"
}

// IsServer tells if the given deployment type is "server".
func IsServer(deployType string) bool {
	return deployType == "server"
}

// IsDev tells if the given deployment type is "dev".
func IsDev(deployType string) bool {
	return deployType == "dev"
}

// ProductName reports the name of the Sourcegraph product that is currently running ("Sourcegraph
// Server", "Sourcegraph Data Center", etc.).
func ProductName() string {
	deployType := DeployType()
	switch {
	case IsDataCenter(deployType):
		return "Sourcegraph Data Center"
	case IsServer(deployType):
		return "Sourcegraph Server"
	case IsDev(deployType):
		return "Sourcegraph Dev"
	default:
		// Should not reach here, but return a reasonable value just in case.
		return "Sourcegraph"
	}
}

// DebugManageDocker tells if Docker language servers should be managed or not.
//
// This only exists for dev mode / debugging purposes, and should never be used
// in a production setting. It panics if the deploy type is not "dev".
func DebugManageDocker() bool {
	if deployType := DeployType(); !IsDev(deployType) {
		panic(fmt.Sprintf("DebugManageDocker cannot be called except when DEPLOY_TYPE=dev (found %q)", deployType))
	}
	v, err := strconv.ParseBool(os.Getenv("DEBUG_MANAGE_DOCKER"))
	if err != nil {
		return true // use managed docker by default in dev mode
	}
	return v
}

// DebugNoDockerSocket returns true if the application should pretend that
// there is no Docker socket present, regardless of what the filesystem says.
//
// This is useful for testing that state in a dev instance, which must be
// supported as users can remove the Docker socket pass-through for e.g.
// security reasons.
//
// This only exists for dev mode / debugging purposes, and should never be used
// in a production setting. It panics if the deploy type is not "dev".
func DebugNoDockerSocket() bool {
	if deployType := DeployType(); !IsDev(deployType) {
		panic(fmt.Sprintf("DebugManageDocker cannot be called except when DEPLOY_TYPE=dev (found %q)", deployType))
	}
	v, _ := strconv.ParseBool(os.Getenv("DEBUG_NO_DOCKER_SOCKET"))
	return v
}

// UpdateChannel tells the update channel. Default is "release".
func UpdateChannel() string {
	channel := GetTODO().UpdateChannel
	if channel == "" {
		return "release"
	}
	return channel
}

// SupportsManagingLanguageServers reports, by consulting *only* the site configuration and deploy
// type, whether language servers can be managed (enabled/disabled/restarted/updated) from the
// Sourcegraph API without requiring users to take manual steps.
//
// Callers needing to know whether the capability is actually present in the current environment
// must use langservers.CanManage instead.
//
// If no, the boolean is false and the reason (which is always non-empty in this case) describes why
// not. Otherwise the boolean is true and the reason is empty.
func SupportsManagingLanguageServers() (reason string, ok bool) {
	deployType := DeployType()
	if IsDataCenter(deployType) {
		// Do not run in Data Center, or else we would print log messages below
		// about not finding the docker socket.
		return "Managing language servers automatically is not supported in Sourcegraph Data Center. See https://github.com/sourcegraph/deploy-sourcegraph/blob/master/docs/install.md#add-language-servers-for-code-intelligence for help.", false
	}
	if IsDev(deployType) && !DebugManageDocker() {
		// Running in dev mode with managed docker disabled.
		return "Managing language servers automatically is disabled in your local dev instance due to the value of DEBUG_MANAGE_DOCKER.", false
	}
	return "", true
}
