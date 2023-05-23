// The functions in this file are used to redact secrets from ExternalServices in
// transit, eg when written back and forth between the client and API, as we
// don't want to leak an access token once it's been configured. Any config
// written back from the client with a redacted token should then be updated with
// the real token from the database, the validation in the ExternalService DB
// methods will check for this field and throw an error if it's not been
// replaced, to prevent us accidentally blanking tokens in the DB.

package types

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"sort"

	"github.com/sourcegraph/jsonx"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// RedactedSecret is used as a placeholder for secret fields when reading external service config
const RedactedSecret = "REDACTED"

type errCodeHostIdentityChanged struct {
	identityProperty string
	redactedProperty string
}

func (e errCodeHostIdentityChanged) Error() string {
	return fmt.Sprintf("Property %q has been changed, please re-enter %q", e.identityProperty, e.redactedProperty)
}

// RedactedConfig returns the external service config with all secret fields replaces by RedactedSecret.
func (e *ExternalService) RedactedConfig(ctx context.Context) (string, error) {
	rawConfig, err := e.Config.Decrypt(ctx)
	if err != nil {
		return "", err
	}
	if rawConfig == "" {
		return "", nil
	}

	cfg, err := e.Configuration(ctx)
	if err != nil {
		return "", err
	}

	var es edits
	switch c := cfg.(type) {
	case *schema.GitHubConnection:
		es.redactString(c.Token, "token")
	case *schema.GitLabConnection:
		es.redactString(c.Token, "token")
		es.redactString(c.TokenOauthRefresh, "token.oauth.refresh")
	case *schema.AzureDevOpsConnection:
		es.redactString(c.Token, "token")
	case *schema.GerritConnection:
		es.redactString(c.Password, "password")
	case *schema.BitbucketServerConnection:
		es.redactString(c.Password, "password")
		es.redactString(c.Token, "token")
	case *schema.BitbucketCloudConnection:
		es.redactString(c.AppPassword, "appPassword")
	case *schema.AWSCodeCommitConnection:
		es.redactString(c.SecretAccessKey, "secretAccessKey")
		es.redactString(c.GitCredentials.Password, "gitCredentials", "password")
	case *schema.PhabricatorConnection:
		es.redactString(c.Token, "token")
	case *schema.PerforceConnection:
		es.redactString(c.P4Passwd, "p4.passwd")
	case *schema.GitoliteConnection:
		// Nothing to redact
	case *schema.GoModulesConnection:
		for i := range c.Urls {
			err = es.redactURL(c.Urls[i], "urls", i)
			if err != nil {
				return "", err
			}
		}
	case *schema.PythonPackagesConnection:
		for i := range c.Urls {
			err = es.redactURL(c.Urls[i], "urls", i)
			if err != nil {
				return "", err
			}
		}
	case *schema.RustPackagesConnection:
		// Nothing to redact
	case *schema.RubyPackagesConnection:
		es.redactString(c.Repository, "repository")
	case *schema.JVMPackagesConnection:
		es.redactString(c.Maven.Credentials, "maven", "credentials")
	case *schema.PagureConnection:
		es.redactString(c.Token, "token")
	case *schema.NpmPackagesConnection:
		es.redactString(c.Credentials, "credentials")
	case *schema.OtherExternalServiceConnection:
		err = es.redactURL(c.Url, "url")
		if err != nil {
			return "", err
		}
	case *schema.LocalGitExternalService:
		// Nothing to redact
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return "", errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(cfg))
	}

	return es.apply(rawConfig)
}

// UnredactConfig will replace redacted fields with their unredacted form from the 'old' ExternalService.
// You should call this when accepting updated config from a user that may have been
// previously redacted, and pass in the unredacted form directly from the DB as the 'old' parameter
func (e *ExternalService) UnredactConfig(ctx context.Context, old *ExternalService) error {
	if e == nil || old == nil {
		return nil
	}

	eConfig, err := e.Config.Decrypt(ctx)
	if err != nil || eConfig == "" {
		return err
	}
	oldConfig, err := old.Config.Decrypt(ctx)
	if err != nil || oldConfig == "" {
		return err
	}

	if old.Kind != e.Kind {
		return errors.Errorf(
			"UnredactExternalServiceConfig: unmatched external service kinds, old: %q, e: %q",
			old.Kind,
			e.Kind,
		)
	}

	newCfg, err := e.Configuration(ctx)
	if err != nil {
		return err
	}

	oldCfg, err := old.Configuration(ctx)
	if err != nil {
		return err
	}

	var es edits

	switch c := newCfg.(type) {
	case *schema.GitHubConnection:
		o := oldCfg.(*schema.GitHubConnection)
		if c.Token == RedactedSecret && c.Url != o.Url {
			return errCodeHostIdentityChanged{"url", "token"}
		}
		es.unredactString(c.Token, o.Token, "token")
	case *schema.GitLabConnection:
		o := oldCfg.(*schema.GitLabConnection)
		if c.Token == RedactedSecret && c.Url != o.Url {
			return errCodeHostIdentityChanged{"url", "token"}
		}
		es.unredactString(c.Token, o.Token, "token")
		es.unredactString(c.TokenOauthRefresh, o.TokenOauthRefresh, "token.oauth.refresh")
	case *schema.BitbucketServerConnection:
		o := oldCfg.(*schema.BitbucketServerConnection)
		var redactedProperty string
		if c.Url != o.Url {
			if c.Password == RedactedSecret {
				redactedProperty = "password"
			}
			if c.Token == RedactedSecret {
				redactedProperty = "token"
			}

			if redactedProperty != "" {
				return errCodeHostIdentityChanged{"url", redactedProperty}
			}
		}
		es.unredactString(c.Password, o.Password, "password")
		es.unredactString(c.Token, o.Token, "token")
	case *schema.BitbucketCloudConnection:
		o := oldCfg.(*schema.BitbucketCloudConnection)
		es.unredactString(c.AppPassword, o.AppPassword, "appPassword")
		if c.Url != o.Url {
			return errCodeHostIdentityChanged{"apiUrl", "appPassword"}
		}
	case *schema.AWSCodeCommitConnection:
		o := oldCfg.(*schema.AWSCodeCommitConnection)
		es.unredactString(c.SecretAccessKey, o.SecretAccessKey, "secretAccessKey")
		es.unredactString(c.GitCredentials.Password, o.GitCredentials.Password, "gitCredentials", "password")
	case *schema.PhabricatorConnection:
		o := oldCfg.(*schema.PhabricatorConnection)
		if c.Token == RedactedSecret && c.Url != o.Url {
			return errCodeHostIdentityChanged{"url", "token"}
		}
		es.unredactString(c.Token, o.Token, "token")
	case *schema.PerforceConnection:
		o := oldCfg.(*schema.PerforceConnection)
		if c.P4Passwd == RedactedSecret && c.P4Port != o.P4Port {
			return errCodeHostIdentityChanged{"p4.port", "p4.passwd"}
		}
		es.unredactString(c.P4Passwd, o.P4Passwd, "p4.passwd")
	case *schema.GerritConnection:
		o := oldCfg.(*schema.GerritConnection)
		es.unredactString(c.Password, o.Password, "password")
		if c.Url != o.Url {
			return errCodeHostIdentityChanged{"url", "password"}
		}
	case *schema.GitoliteConnection:
		// Nothing to redact
	case *schema.GoModulesConnection:
		err = es.unredactURLs(c.Urls, oldCfg.(*schema.GoModulesConnection).Urls)
		if err != nil {
			return err
		}
	case *schema.PythonPackagesConnection:
		err = es.unredactURLs(c.Urls, oldCfg.(*schema.PythonPackagesConnection).Urls)
		if err != nil {
			return err
		}
	case *schema.RustPackagesConnection:
		// Nothing to unredact
	case *schema.RubyPackagesConnection:
		o := oldCfg.(*schema.RubyPackagesConnection)
		es.unredactString(c.Repository, o.Repository, "repository")
	case *schema.JVMPackagesConnection:
		o := oldCfg.(*schema.JVMPackagesConnection)
		// credentials didn't change check if repositories did
		if c.Maven.Credentials == RedactedSecret {
			oldRepos := o.Maven.Repositories
			sort.Strings(oldRepos)

			newRepos := c.Maven.Repositories
			sort.Strings(newRepos)

			// if we only remove a known repo, it's fine
			if len(newRepos) < len(oldRepos) {
				for _, r := range newRepos {
					// we have a new repo in the list, return error
					if !slices.Contains(oldRepos, r) {
						return errCodeHostIdentityChanged{"repositories", "credentials"}
					}
				}
			} else if !slices.Equal(oldRepos, newRepos) {
				return errCodeHostIdentityChanged{"repositories", "credentials"}
			}
		}
		es.unredactString(c.Maven.Credentials, o.Maven.Credentials, "maven", "credentials")
	case *schema.PagureConnection:
		o := oldCfg.(*schema.PagureConnection)
		if c.Token == RedactedSecret && c.Url != o.Url {
			return errCodeHostIdentityChanged{"url", "token"}
		}
		es.unredactString(c.Token, o.Token, "token")
	case *schema.AzureDevOpsConnection:
		o := oldCfg.(*schema.AzureDevOpsConnection)
		if c.Token == RedactedSecret && c.Url != o.Url {
			return errCodeHostIdentityChanged{"url", "token"}
		}
		es.unredactString(c.Token, o.Token, "token")
	case *schema.NpmPackagesConnection:
		o := oldCfg.(*schema.NpmPackagesConnection)
		if c.Credentials == RedactedSecret && c.Registry != o.Registry {
			return errCodeHostIdentityChanged{"registry", "credentials"}
		}
		es.unredactString(c.Credentials, o.Credentials, "credentials")
	case *schema.OtherExternalServiceConnection:
		o := oldCfg.(*schema.OtherExternalServiceConnection)
		err := es.unredactURL(c.Url, o.Url, "url")
		if err != nil {
			return err
		}
		oldParsed, err := url.Parse(o.Url)
		if err != nil {
			return err
		}
		newParsed, err := url.Parse(c.Url)
		if err != nil {
			return err
		}
		// compare URLs and see if password changed
		pwd, ok := newParsed.User.Password()

		// remove UserInfo so we can compare URLs
		oldParsed.User = nil
		newParsed.User = nil

		if newParsed.String() != oldParsed.String() {
			if ok && pwd == RedactedSecret {
				return errCodeHostIdentityChanged{"url", "password"}
			}
		}

	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(newCfg))
	}

	unredacted, err := es.apply(eConfig)
	if err != nil {
		return err
	}

	e.Config.Set(unredacted)
	return nil
}

type edits []edit

func (es edits) apply(input string) (output string, err error) {
	output = input
	for _, e := range es {
		if output, err = e.apply(output); err != nil {
			return "", err
		}
	}
	return
}

func (es *edits) edit(v any, path ...any) {
	*es = append(*es, edit{jsonx.MakePath(path...), v})
}

func (es *edits) redactString(s string, path ...any) {
	if s != "" {
		es.edit(redactedString(s), path...)
	}
}

func (es *edits) unredactString(new, old string, path ...any) {
	if new != "" && old != "" {
		es.edit(unredactedString(new, old), path...)
	}
}

func (es *edits) redactURL(s string, path ...any) error {
	if s == "" {
		return nil
	}

	redacted, err := redactedURL(s)
	if err != nil {
		return err
	}

	es.edit(redacted, path...)
	return nil
}

func (es *edits) unredactURLs(new, old []string) (err error) {
	m := make(map[string]string, len(old))

	for _, oldURL := range old {
		if oldURL == "" {
			continue
		}

		parsed, err := url.Parse(oldURL)
		if err != nil {
			return err
		}

		parsed.User = nil
		m[parsed.String()] = oldURL
	}

	for i := range new {
		parsed, err := url.Parse(new[i])
		if err != nil {
			return err
		}

		pwd, set := parsed.User.Password()
		parsed.User = nil

		oldURL, ok := m[parsed.String()]

		if !ok {
			if set && pwd == RedactedSecret {
				return errCodeHostIdentityChanged{"url", "password"}
			}
			continue
		}

		err = es.unredactURL(new[i], oldURL, "urls", i)
		if err != nil {
			return err
		}
	}

	return nil
}

func (es *edits) unredactURL(new, old string, path ...any) error {
	if new == "" || old == "" {
		return nil
	}

	unredacted, err := unredactedURL(new, old)
	if err != nil {
		return err
	}

	es.edit(unredacted, path...)
	return nil
}

type edit struct {
	path  jsonx.Path
	value any
}

func (p edit) apply(input string) (string, error) {
	edits, _, err := jsonx.ComputePropertyEdit(input, p.path, p.value, nil, jsonc.DefaultFormatOptions)
	if err != nil {
		return "", err
	}
	return jsonx.ApplyEdits(input, edits...)
}

func redactedString(s string) string {
	if s != "" {
		return RedactedSecret
	}
	return ""
}

func unredactedString(new, old string) string {
	if new == RedactedSecret {
		return old
	}
	return new
}

func redactedURL(rawURL string) (string, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if _, ok := parsed.User.Password(); !ok {
		return parsed.String(), nil
	}

	parsed.User = url.UserPassword(parsed.User.Username(), RedactedSecret)
	return parsed.String(), nil
}

func unredactedURL(new, old string) (string, error) {
	newURL, err := url.Parse(new)
	if err != nil {
		return new, err
	}

	oldURL, err := url.Parse(old)
	if err != nil {
		return new, err
	}

	passwd, ok := newURL.User.Password()
	if !ok || passwd != RedactedSecret {
		return new, nil
	}

	oldPasswd, _ := oldURL.User.Password()
	if oldPasswd != "" {
		newURL.User = url.UserPassword(newURL.User.Username(), oldPasswd)
	} else {
		newURL.User = url.User(newURL.User.Username())
	}

	return newURL.String(), nil
}
