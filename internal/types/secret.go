// The functions in this file are used to redact secrets from ExternalServices in
// transit, eg when written back and forth between the client and API, as we
// don't want to leak an access token once it's been configured. Any config
// written back from the client with a redacted token should then be updated with
// the real token from the database, the validation in the ExternalService DB
// methods will check for this field and throw an error if it's not been
// replaced, to prevent us accidentally blanking tokens in the DB.

package types

import (
	"net/url"
	"reflect"

	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// RedactedSecret is used as a placeholder for secret fields when reading external service config
const RedactedSecret = "REDACTED"

// RedactedConfig returns the external service config with all secret fields replaces by RedactedSecret.
func (e *ExternalService) RedactedConfig() (string, error) {
	if e.Config == "" {
		return "", nil
	}

	cfg, err := e.Configuration()
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
	case *schema.JVMPackagesConnection:
		if c.Maven != nil {
			es.redactString(c.Maven.Credentials, "maven", "credentials")
		}
	case *schema.PagureConnection:
		es.redactString(c.Token, "token")
	case *schema.NpmPackagesConnection:
		es.redactString(c.Credentials, "credentials")
	case *schema.OtherExternalServiceConnection:
		err = es.redactURL(c.Url, "url")
		if err != nil {
			return "", err
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return "", errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(cfg))
	}

	return es.apply(e.Config)
}

// UnredactConfig will replace redacted fields with their unredacted form from the 'old' ExternalService.
// You should call this when accepting updated config from a user that may have been
// previously redacted, and pass in the unredacted form directly from the DB as the 'old' parameter
func (e *ExternalService) UnredactConfig(old *ExternalService) error {
	if e == nil || old == nil || e.Config == "" || old.Config == "" {
		return nil
	}

	if old.Kind != e.Kind {
		return errors.Errorf(
			"UnRedactExternalServiceConfig: unmatched external service kinds, old: %q, e: %q",
			old.Kind,
			e.Kind,
		)
	}

	newCfg, err := e.Configuration()
	if err != nil {
		return err
	}

	oldCfg, err := old.Configuration()
	if err != nil {
		return err
	}

	var es edits

	switch c := newCfg.(type) {
	case *schema.GitHubConnection:
		o := oldCfg.(*schema.GitHubConnection)
		es.unredactString(c.Token, o.Token, "token")
	case *schema.GitLabConnection:
		o := oldCfg.(*schema.GitLabConnection)
		es.unredactString(c.Token, o.Token, "token")
		es.unredactString(c.TokenOauthRefresh, o.TokenOauthRefresh, "token.oauth.refresh")
	case *schema.BitbucketServerConnection:
		o := oldCfg.(*schema.BitbucketServerConnection)
		es.unredactString(c.Password, o.Password, "password")
		es.unredactString(c.Token, o.Token, "token")
	case *schema.BitbucketCloudConnection:
		o := oldCfg.(*schema.BitbucketCloudConnection)
		es.unredactString(c.AppPassword, o.AppPassword, "appPassword")
	case *schema.AWSCodeCommitConnection:
		o := oldCfg.(*schema.AWSCodeCommitConnection)
		es.unredactString(c.SecretAccessKey, o.SecretAccessKey, "secretAccessKey")
		es.unredactString(c.GitCredentials.Password, o.GitCredentials.Password, "gitCredentials", "password")
	case *schema.PhabricatorConnection:
		o := oldCfg.(*schema.PhabricatorConnection)
		es.unredactString(c.Token, o.Token, "token")
	case *schema.PerforceConnection:
		o := oldCfg.(*schema.PerforceConnection)
		es.unredactString(c.P4Passwd, o.P4Passwd, "p4.passwd")
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
	case *schema.JVMPackagesConnection:
		o := oldCfg.(*schema.JVMPackagesConnection)
		if c.Maven != nil && o.Maven != nil {
			es.unredactString(c.Maven.Credentials, o.Maven.Credentials, "maven", "credentials")
		}
	case *schema.PagureConnection:
		o := oldCfg.(*schema.PagureConnection)
		es.unredactString(c.Token, o.Token, "token")
	case *schema.NpmPackagesConnection:
		o := oldCfg.(*schema.NpmPackagesConnection)
		es.unredactString(c.Credentials, o.Credentials, "credentials")
	case *schema.OtherExternalServiceConnection:
		o := oldCfg.(*schema.OtherExternalServiceConnection)
		err = es.unredactURL(c.Url, o.Url, "url")
		if err != nil {
			return err
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(newCfg))
	}

	unredacted, err := es.apply(e.Config)
	if err != nil {
		return err
	}

	e.Config = unredacted
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

		redactedOldURL, err := redactedURL(oldURL)
		if err != nil {
			return err
		}

		m[redactedOldURL] = oldURL
	}

	for i := range new {
		oldURL, ok := m[new[i]]
		if !ok {
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
	redacted, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if _, ok := redacted.User.Password(); !ok {
		return rawURL, nil
	}

	redacted.User = url.UserPassword(redacted.User.Username(), RedactedSecret)
	return redacted.String(), nil
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
