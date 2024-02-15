package cloud

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"text/template"

	"golang.org/x/crypto/ssh"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	// rawSiteConfig is the base64-encoded string that is signed by the "Sourcegraph
	// Cloud site config singer" private key, which is available at
	// https://team-sourcegraph.1password.com/vaults/dnrhbauihkhjs5ag6vszsme45a/allitems/m4rqoaoujjwesf6twwqyr3lpde.
	rawSiteConfig = env.Get("SRC_CLOUD_SITE_CONFIG", "", "The site configuration specifically for Sourcegraph Cloud")

	defaultNotAllowedErrorMessageTmpl = template.Must(template.New("").Parse("Editing {{.Paths}} in site configuration is not allowed on Sourcegraph Cloud. Please contact support."))
)

// sourcegraphCloudSiteConfigSignerPublicKey is the counterpart of the
// "Sourcegraph Cloud site config singer" private key.
const sourcegraphCloudSiteConfigSignerPublicKey = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIFnVjzARMu+jaSrTgvJCpWEDP503Y3k3DMbs5ghHOkML"

// SignedSiteConfig is the data structure for a site config and its signature.
type SignedSiteConfig struct {
	Signature  *ssh.Signature `json:"signature"`
	SiteConfig []byte         `json:"siteConfig"` // Based64-encoded JSON blob
}

func parseSiteConfig(raw string) (*SchemaSiteConfig, error) {
	publicKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(sourcegraphCloudSiteConfigSignerPublicKey))
	if err != nil {
		return nil, errors.Wrap(err, "parse signer public key")
	}

	signedData, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, errors.Wrap(err, "decode raw site config")
	}

	var signedSiteConfig SignedSiteConfig
	err = json.Unmarshal(signedData, &signedSiteConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal signed data")
	}

	err = publicKey.Verify(signedSiteConfig.SiteConfig, signedSiteConfig.Signature)
	if err != nil {
		return nil, errors.Wrap(err, "verify signed data")
	}

	var siteConfig SchemaSiteConfig
	err = json.Unmarshal(signedSiteConfig.SiteConfig, &siteConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshal verified site config")
	}

	if siteConfig.SiteConfigAllowlistEnabled() {
		if siteConfig.SiteConfigAllowlist.NotAllowedErrorMessage == "" {
			siteConfig.SiteConfigAllowlist.errorMessageTmpl = defaultNotAllowedErrorMessageTmpl
		} else {
			siteConfig.SiteConfigAllowlist.errorMessageTmpl, err = template.New("").Parse(siteConfig.SiteConfigAllowlist.NotAllowedErrorMessage)
			if err != nil {
				return nil, errors.Wrap(err, "parse error message template")
			}
		}
	}
	return &siteConfig, nil
}

var (
	parsedSiteConfigOnce sync.Once
	parsedSiteConfig     *SchemaSiteConfig
)

// MockSiteConfig uses the given mock version to be returned for the subsequent
// calls of SiteConfig function, and restores to the previous version once the
// test suite is finished.
func MockSiteConfig(t *testing.T, mock *SchemaSiteConfig) {
	parsedSiteConfigOnce.Do(func() {}) // Prevent the real "do" to be executed

	parsedSiteConfig = mock
	t.Cleanup(func() {
		parsedSiteConfig = nil
	})
}

// SiteConfig returns the parsed Sourcegraph Cloud site config.
func SiteConfig() *SchemaSiteConfig {
	parsedSiteConfigOnce.Do(func() {
		if rawSiteConfig == "" {
			// Init a stub object to avoid all the top-level nit- and probing-checks
			parsedSiteConfig = &SchemaSiteConfig{}
			return
		}

		var err error
		parsedSiteConfig, err = parseSiteConfig(rawSiteConfig)
		if err != nil {
			panic("failed to parse Sourcegraph Cloud site config: " + err.Error())
		}
	})
	return parsedSiteConfig
}

// SchemaSiteConfig contains the Sourcegraph Cloud site config.
type SchemaSiteConfig struct {
	AuthProviders *SchemaAuthProviders `json:"authProviders"`
	// SiteConfigAllowlist controls what site config attributes
	// Cloud customers are allowed to change
	SiteConfigAllowlist SiteConfigAllowlistSpec `json:"siteConfigAllowlist,omitempty"`
}

type SiteConfigAllowlistSpec struct {
	// NotAllowedErrorMessage is a go template string to show error message.
	// Available variables: {{.Paths}}
	NotAllowedErrorMessage string `json:"notAllowedErrorMessage,omitempty"`
	// Paths is a list of keys in the site config that are allowed to be changed
	// Notes:
	// 	- only top-level keys are supported
	Paths []string `json:"paths"`

	errorMessageTmpl *template.Template `json:"-"`
}

// SchemaAuthProviders contains the authentication providers for Sourcegraph
// Cloud.
type SchemaAuthProviders struct {
	SourcegraphOperator *SchemaAuthProviderSourcegraphOperator `json:"sourcegraphOperator"`
}

// SchemaAuthProviderSourcegraphOperator contains configuration for the
// Sourcegraph Operator authentication provider.
type SchemaAuthProviderSourcegraphOperator struct {
	Issuer       string `json:"issuer"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`

	// LifecycleDuration indicates duration in minutes before accounts created
	// through SOAP are expired and removed.
	LifecycleDuration int `json:"lifecycleDuration"`
}

// SourcegraphOperatorAuthProviderEnabled returns true if the Sourcegraph
// Operator authentication provider has been enabled.
func (s *SchemaSiteConfig) SourcegraphOperatorAuthProviderEnabled() bool {
	return s.AuthProviders != nil && s.AuthProviders.SourcegraphOperator != nil
}

func (s *SchemaSiteConfig) SiteConfigAllowlistEnabled() bool {
	return s.SourcegraphOperatorAuthProviderEnabled() && len(s.SiteConfigAllowlist.Paths) > 0
}

func (s *SchemaSiteConfig) SiteConfigAllowlistOnError(paths []string) error {
	if !s.SiteConfigAllowlistEnabled() {
		return nil
	}
	var b bytes.Buffer
	if err := s.SiteConfigAllowlist.errorMessageTmpl.Execute(&b, struct {
		Paths string
	}{
		Paths: fmt.Sprintf("[%s]", strings.Join(paths, ", ")),
	}); err != nil {
		return errors.Wrapf(err, "Execute error message template: Editing %q in site configuration is not allowed. Please contact support", paths)
	}
	return errors.New(b.String())
}
