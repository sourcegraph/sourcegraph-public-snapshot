package ldap

import "src.sourcegraph.com/sourcegraph/sgx/cli"

// Flags defines settings (in the form of CLI flags) for auth via LDAP.
// LDAP auth must be turned on via "--auth.source=ldap" for
// these settings to take effect.
type Flags struct {
	Host string `long:"ldap.host" description:"host name of the LDAP server to connect to for authentication" default:"localhost"`

	Port int `long:"ldap.port" description:"port on LDAP server to connect to" default:"389"`

	TLS bool `long:"ldap.tls" description:"connect to LDAP server using TLS"`

	TLSSkipVerify bool `long:"ldap.tls-skip-verify" description:"skip verification of TLS certificate (not recommended for production)"`

	SearchUser string `long:"ldap.search-user" description:"The LDAP user that performs user lookups. (eg. cn=sourcegraph_ldap,ou=People,dc=mycompany,dc=com)"`

	SearchPassword string `long:"ldap.search-password" description:"The password for the search user."`

	DomainBase string `long:"ldap.domain-base" description:"The point in the LDAP tree where users are searched from. (eg. ou=People,dc=mycompany,dc=com)"`

	Filter string `long:"ldap.filter" description:"Filter the search query (eg. objectClass=users)"`

	UserIDField string `long:"ldap.user-id" description:"The LDAP field mapped to user ID" default:"uid"`

	ProfileNameField string `long:"ldap.profile-name" description:"The LDAP field mapped to profile name" default:"cn"`

	EmailField string `long:"ldap.email" description:"The LDAP field mapped to user email" default:"mail"`
}

// Config is the currently active LDAP config (as set by the CLI flags).
var Config Flags

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		cli.Serve.AddGroup("LDAP", "LDAP", &Config)
	})
}
