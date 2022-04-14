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

	config := e.Config

	switch c := cfg.(type) {
	case *schema.GitHubConnection:
		if c.Token != "" {
			config, err = patch(config, redactedString(c.Token), "token")
		}
	case *schema.GitLabConnection:
		if c.Token != "" {
			config, err = patch(config, redactedString(c.Token), "token")
		}
	case *schema.BitbucketServerConnection:
		if c.Password != "" {
			config, err = patch(config, redactedString(c.Password), "password")
			if err != nil {
				return "", err
			}
		}

		if c.Token != "" {
			config, err = patch(config, redactedString(c.Token), "token")
			if err != nil {
				return "", err
			}
		}
	case *schema.BitbucketCloudConnection:
		if c.AppPassword != "" {
			config, err = patch(config, redactedString(c.AppPassword), "appPassword")
		}
	case *schema.AWSCodeCommitConnection:
		if c.SecretAccessKey != "" {
			config, err = patch(config, redactedString(c.SecretAccessKey), "secretAccessKey")
			if err != nil {
				return "", err
			}
		}
		if c.GitCredentials.Password != "" {
			config, err = patch(config, redactedString(c.GitCredentials.Password), "gitCredentials", "password")
			if err != nil {
				return "", err
			}
		}
	case *schema.PhabricatorConnection:
		if c.Token != "" {
			config, err = patch(config, redactedString(c.Token), "token")
		}
	case *schema.PerforceConnection:
		if c.P4Passwd != "" {
			config, err = patch(config, redactedString(c.P4Passwd), "p4.passwd")
		}
	case *schema.GitoliteConnection:
		// Nothing to redact
	case *schema.GoModulesConnection:
		for i := range c.Urls {
			if c.Urls[i] == "" {
				continue
			}

			redacted, err := redactedURL(c.Urls[i])
			if err != nil {
				return "", err
			}

			config, err = patch(config, redacted, "urls", i)
		}
	case *schema.JVMPackagesConnection:
		if c.Maven != nil {
			config, err = patch(config, redactedString(c.Maven.Credentials), "maven", "credentials")
		}
	case *schema.PagureConnection:
		if c.Token != "" {
			config, err = patch(config, redactedString(c.Token), "token")
		}
	case *schema.NpmPackagesConnection:
		if c.Credentials != "" {
			config, err = patch(config, redactedString(c.Credentials), "credentials")
		}
	case *schema.OtherExternalServiceConnection:
		if c.Url != "" {
			redacted, err := redactedURL(c.Url)
			if err != nil {
				return "", err
			}
			config, err = patch(config, redacted, "url")
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return "", errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(cfg))
	}

	if err != nil {
		return "", err
	}

	return config, nil
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

	config := e.Config

	switch c := newCfg.(type) {
	case *schema.GitHubConnection:
		o := oldCfg.(*schema.GitHubConnection)
		if c.Token != "" {
			config, err = patch(config, unredactedString(c.Token, o.Token), "token")
		}
	case *schema.GitLabConnection:
		o := oldCfg.(*schema.GitLabConnection)
		if c.Token != "" {
			config, err = patch(config, unredactedString(c.Token, o.Token), "token")
		}
	case *schema.BitbucketServerConnection:
		o := oldCfg.(*schema.BitbucketServerConnection)
		if c.Password != "" {
			config, err = patch(config, unredactedString(c.Password, o.Password), "password")
			if err != nil {
				return err
			}
		}
		if c.Token != "" {
			config, err = patch(config, unredactedString(c.Token, o.Token), "token")
			if err != nil {
				return err
			}
		}
	case *schema.BitbucketCloudConnection:
		o := oldCfg.(*schema.BitbucketCloudConnection)
		if c.AppPassword != "" {
			config, err = patch(config, unredactedString(c.AppPassword, o.AppPassword), "appPassword")
		}
	case *schema.AWSCodeCommitConnection:
		o := oldCfg.(*schema.AWSCodeCommitConnection)
		if c.SecretAccessKey != "" {
			config, err = patch(config, unredactedString(c.SecretAccessKey, o.SecretAccessKey), "secretAccessKey")
			if err != nil {
				return err
			}
		}
		if c.GitCredentials.Password != "" {
			config, err = patch(config, unredactedString(c.GitCredentials.Password, o.GitCredentials.Password), "gitCredentials", "password")
			if err != nil {
				return err
			}
		}
	case *schema.PhabricatorConnection:
		o := oldCfg.(*schema.PhabricatorConnection)
		if c.Token != "" {
			config, err = patch(config, unredactedString(c.Token, o.Token), "token")
		}
	case *schema.PerforceConnection:
		o := oldCfg.(*schema.PerforceConnection)
		if c.P4Passwd != "" {
			config, err = patch(config, unredactedString(c.P4Passwd, o.P4Passwd), "p4.passwd")
		}
	case *schema.GitoliteConnection:
		// Nothing to redact
	case *schema.GoModulesConnection:
		o := oldCfg.(*schema.GoModulesConnection)
		m := make(map[string]string, len(o.Urls))

		for _, oldURL := range o.Urls {
			if oldURL == "" {
				continue
			}

			redactedOldURL, err := redactedURL(oldURL)
			if err != nil {
				return err
			}

			m[redactedOldURL] = oldURL
		}

		for i := range c.Urls {
			oldURL, ok := m[c.Urls[i]]
			if !ok {
				continue
			}

			unredacted, err := unredactedURL(c.Urls[i], oldURL)
			if err != nil {
				return err
			}

			config, err = patch(config, unredacted, "urls", i)
			if err != nil {
				return err
			}
		}
	case *schema.JVMPackagesConnection:
		o := oldCfg.(*schema.JVMPackagesConnection)
		if c.Maven != nil && c.Maven.Credentials != "" && o.Maven != nil {
			config, err = patch(config, unredactedString(c.Maven.Credentials, o.Maven.Credentials), "maven", "credentials")
		}
	case *schema.PagureConnection:
		o := oldCfg.(*schema.PagureConnection)
		if c.Token != "" {
			config, err = patch(config, unredactedString(c.Token, o.Token), "token")
		}
	case *schema.NpmPackagesConnection:
		o := oldCfg.(*schema.NpmPackagesConnection)
		if c.Credentials != "" {
			config, err = patch(config, unredactedString(c.Credentials, o.Credentials), "credentials")
		}
	case *schema.OtherExternalServiceConnection:
		o := oldCfg.(*schema.OtherExternalServiceConnection)
		if c.Url != "" {
			unredacted, err := unredactedURL(c.Url, o.Url)
			if err != nil {
				return err
			}
			config, err = patch(config, unredacted, "url")
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(newCfg))
	}

	if err != nil {
		return err
	}

	e.Config = config
	return nil
}

func patch(input string, value interface{}, path ...interface{}) (string, error) {
	edits, _, err := jsonx.ComputePropertyEdit(input, jsonx.MakePath(path...), value, nil, jsonc.DefaultFormatOptions)
	if err != nil {
		return input, err
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
