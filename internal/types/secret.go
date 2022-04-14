// The functions in this file are used to redact secrets from ExternalServices in
// transit, eg when written back and forth between the client and API, as we
// don't want to leak an access token once it's been configured. Any config
// written back from the client with a redacted token should then be updated with
// the real token from the database, the validation in the ExternalService DB
// methods will check for this field and throw an error if it's not been
// replaced, to prevent us accidentally blanking tokens in the DB.

package types

import (
	"encoding/json"
	"net/url"
	"reflect"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// RedactedSecret is used as a placeholder for secret fields when reading external service config
const RedactedSecret = "REDACTED"

// RedactConfigSecrets replaces any secret fields in the Config field with RedactedSecret, be sure to call
// UnRedactExternalServiceConfig before writing back to the database, otherwise validation will throw errors.
func (e *ExternalService) RedactConfigSecrets() (string, error) {
	if e.Config == "" {
		return "", nil
	}

	cfg, err := e.Configuration()
	if err != nil {
		return "", err
	}

	switch c := cfg.(type) {
	case *schema.GitHubConnection:
		redactString(&c.Token)
	case *schema.GitLabConnection:
		redactString(&c.Token)
	case *schema.BitbucketServerConnection:
		redactString(&c.Password)
		redactString(&c.Token)
	case *schema.BitbucketCloudConnection:
		redactString(&c.AppPassword)
	case *schema.AWSCodeCommitConnection:
		redactString(&c.SecretAccessKey)
		redactString(&c.GitCredentials.Password)
	case *schema.PhabricatorConnection:
		redactString(&c.Token)
	case *schema.PerforceConnection:
		redactString(&c.P4Passwd)
	case *schema.GitoliteConnection:
		// Nothing to redact
	case *schema.GoModulesConnection:
		for i := range c.Urls {
			if err := redactURL(&c.Urls[i]); err != nil {
				return e.Config, err
			}
		}
	case *schema.JVMPackagesConnection:
		if c.Maven != nil {
			redactString(&c.Maven.Credentials)
		}
	case *schema.PagureConnection:
		redactString(&c.Token)
	case *schema.NpmPackagesConnection:
		redactString(&c.Credentials)
	case *schema.OtherExternalServiceConnection:
		if err := redactURL(&c.Url); err != nil {
			return e.Config, err
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return e.Config, errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(cfg))
	}

	redacted, err := json.Marshal(cfg)
	if err != nil {
		return e.Config, err
	}

	return string(redacted), nil
}

// UnredactConfig will replace redacted fields with their undredacted form from the 'old' ExternalService.
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

	switch c := newCfg.(type) {
	case *schema.GitHubConnection:
		unredactString(&c.Token, oldCfg.(*schema.GitHubConnection).Token)
	case *schema.GitLabConnection:
		unredactString(&c.Token, oldCfg.(*schema.GitLabConnection).Token)
	case *schema.BitbucketServerConnection:
		unredactString(&c.Password, oldCfg.(*schema.BitbucketServerConnection).Password)
		unredactString(&c.Token, oldCfg.(*schema.BitbucketServerConnection).Token)
	case *schema.BitbucketCloudConnection:
		unredactString(&c.AppPassword, oldCfg.(*schema.BitbucketCloudConnection).AppPassword)
	case *schema.AWSCodeCommitConnection:
		unredactString(&c.SecretAccessKey, oldCfg.(*schema.AWSCodeCommitConnection).SecretAccessKey)
		unredactString(&c.GitCredentials.Password, oldCfg.(*schema.AWSCodeCommitConnection).GitCredentials.Password)
	case *schema.PhabricatorConnection:
		unredactString(&c.Token, oldCfg.(*schema.PhabricatorConnection).Token)
	case *schema.PerforceConnection:
		unredactString(&c.P4Passwd, oldCfg.(*schema.PerforceConnection).P4Passwd)
	case *schema.GitoliteConnection:
		// Nothing to redact
	case *schema.GoModulesConnection:
		oldURLs := oldCfg.(*schema.GoModulesConnection).Urls
		for i := range c.Urls {
			if err := unredactURL(&c.Urls[i], oldURLs[i]); err != nil {
				return err
			}
		}
	case *schema.JVMPackagesConnection:
		o := oldCfg.(*schema.JVMPackagesConnection)
		if c.Maven != nil && o.Maven != nil {
			unredactString(&c.Maven.Credentials, o.Maven.Credentials)
		}
	case *schema.PagureConnection:
		unredactString(&c.Token, oldCfg.(*schema.PagureConnection).Token)
	case *schema.NpmPackagesConnection:
		unredactString(&c.Credentials, oldCfg.(*schema.NpmPackagesConnection).Credentials)
	case *schema.OtherExternalServiceConnection:
		o := oldCfg.(*schema.OtherExternalServiceConnection)
		if err := unredactURL(&c.Url, o.Url); err != nil {
			return err
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(newCfg))
	}

	unredacted, err := json.Marshal(newCfg)
	if err != nil {
		return err
	}

	e.Config = string(unredacted)
	return nil
}

func redactString(s *string) {
	if *s != "" {
		*s = RedactedSecret
	}
}

func unredactString(new *string, old string) {
	if *new == RedactedSecret {
		*new = old
	}
}

func redactURL(rawURL *string) (err error) {
	if *rawURL != "" {
		*rawURL, err = redactedURL(*rawURL)
	}
	return
}

func redactedURL(rawURL string) (string, error) {
	redacted, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	if redacted.User == nil {
		return rawURL, nil
	}

	if _, ok := redacted.User.Password(); !ok {
		return rawURL, nil
	}

	redacted.User = url.UserPassword(redacted.User.Username(), RedactedSecret)
	return redacted.String(), nil
}

func unredactURL(new *string, old string) error {
	newURL, err := url.Parse(*new)
	if err != nil {
		return err
	}

	oldURL, err := url.Parse(old)
	if err != nil {
		return err
	}

	passwd, ok := newURL.User.Password()
	if !ok || passwd != RedactedSecret {
		return nil
	}

	oldPasswd, _ := oldURL.User.Password()
	if oldPasswd != "" {
		newURL.User = url.UserPassword(newURL.User.Username(), oldPasswd)
	} else {
		newURL.User = url.User(newURL.User.Username())
	}

	*new = newURL.String()
	return nil
}
