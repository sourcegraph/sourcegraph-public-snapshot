package authutil

import (
	"log"

	"strings"

	sgxcli "sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
)

func init() {
	sgxcli.PostInit = append(sgxcli.PostInit, func() {
		_, err := sgxcli.Serve.AddGroup("Authentication", "Authentication", &ActiveFlags)
		if err != nil {
			log.Fatal(err)
		}
	})
}

// Flags defines some command-line flags for this package.
type Flags struct {
	AllowAnonymousReaders bool `long:"auth.allow-anon-readers" description:"allow unauthenticated users to perform read operations (viewing repos, etc.)"`

	RestrictWriteAccess bool `long:"auth.restrict-write-access" description:"only allow admin users to perform write operations (create/delete repo, push to repo, etc.)" default:"false"`

	RestrictToUsers string `long:"auth.users" description:"restrict to only users listed here (space-separated logins) and forbid all anonymous access ('<ALL>' means allow all users)" default:"<ALL>"`

	Source string `long:"auth.source" description:"source of authentication to use (none|local|oauth)" default:"oauth"`

	OAuth2AuthServer bool `long:"auth.oauth2-auth-server" description:"enable OAuth2 authentication server (allow users to authenticate via this server)"`

	DisableUserProfiles bool `long:"auth.disable-user-profiles" description:"do not show user profile pages"`

	AllowAllLogins bool `long:"auth.allow-all-logins" description:"do not check access permissions of a user at login. CAUTION: use only for testing."`
}

func (f *Flags) allowAllUsers() bool { return f.RestrictToUsers == "<ALL>" || f.RestrictToUsers == "" }

func (f *Flags) AllowSignUpOrLogInForUser(login string) bool {
	if f.allowAllUsers() {
		return true
	}
	logins := strings.Fields(f.RestrictToUsers)
	for _, l := range logins {
		if l == login {
			return true
		}
	}
	return false
}

// IsLocal returns true if users are stored and authenticated locally.
func (f Flags) IsLocal() bool {
	return f.Source == "local"
}

// HasUserAccounts returns a boolean value indicating whether user
// accounts are enabled. If they are disabled, generally no
// login/signup functionality should be displayed or exposed.
func (f Flags) HasUserAccounts() bool {
	return f.Source != "" && f.Source != "none"
}

// HasLogin returns whether logging in is enabled.
func (f Flags) HasLogin() bool { return f.HasUserAccounts() }

// HasSignup returns whether signing up is enabled.
func (f Flags) HasSignup() bool { return f.IsLocal() }

func (f Flags) HasUserProfiles() bool { return !f.DisableUserProfiles }

// ActiveFlags are the flag values passed from the command line, if
// we're running as a CLI.
var ActiveFlags Flags
