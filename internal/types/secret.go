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

	var edits []edit
	switch c := cfg.(type) {
	case *schema.GitHubConnection:
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), redactedString(c.Token)})
		}
	case *schema.GitLabConnection:
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), redactedString(c.Token)})
		}
	case *schema.BitbucketServerConnection:
		if c.Password != "" {
			edits = append(edits, edit{jsonx.MakePath("password"), redactedString(c.Password)})
		}

		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), redactedString(c.Token)})
		}
	case *schema.BitbucketCloudConnection:
		if c.AppPassword != "" {
			edits = append(edits, edit{jsonx.MakePath("appPassword"), redactedString(c.AppPassword)})
		}
	case *schema.AWSCodeCommitConnection:
		if c.SecretAccessKey != "" {
			edits = append(edits, edit{jsonx.MakePath("secretAccessKey"), redactedString(c.SecretAccessKey)})
		}
		if c.GitCredentials.Password != "" {
			edits = append(edits, edit{jsonx.MakePath("gitCredentials", "password"), redactedString(c.GitCredentials.Password)})
		}
	case *schema.PhabricatorConnection:
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), redactedString(c.Token)})
		}
	case *schema.PerforceConnection:
		if c.P4Passwd != "" {
			edits = append(edits, edit{jsonx.MakePath("p4.passwd"), redactedString(c.P4Passwd)})
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

			edits = append(edits, edit{jsonx.MakePath("urls", i), redacted})
		}
	case *schema.JVMPackagesConnection:
		if c.Maven != nil && c.Maven.Credentials != "" {
			edits = append(edits, edit{jsonx.MakePath("maven", "credentials"), redactedString(c.Maven.Credentials)})
		}
	case *schema.PagureConnection:
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), redactedString(c.Token)})
		}
	case *schema.NpmPackagesConnection:
		if c.Credentials != "" {
			edits = append(edits, edit{jsonx.MakePath("credentials"), redactedString(c.Credentials)})
		}
	case *schema.OtherExternalServiceConnection:
		if c.Url != "" {
			redacted, err := redactedURL(c.Url)
			if err != nil {
				return "", err
			}
			edits = append(edits, edit{jsonx.MakePath("url"), redacted})
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return "", errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(cfg))
	}

	redacted := e.Config
	for _, e := range edits {
		if redacted, err = e.apply(redacted); err != nil {
			return "", err
		}
	}

	return redacted, nil
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

	var edits []edit
	switch c := newCfg.(type) {
	case *schema.GitHubConnection:
		o := oldCfg.(*schema.GitHubConnection)
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), unredactedString(c.Token, o.Token)})
		}
	case *schema.GitLabConnection:
		o := oldCfg.(*schema.GitLabConnection)
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), unredactedString(c.Token, o.Token)})
		}
	case *schema.BitbucketServerConnection:
		o := oldCfg.(*schema.BitbucketServerConnection)
		if c.Password != "" {
			edits = append(edits, edit{jsonx.MakePath("password"), unredactedString(c.Password, o.Password)})
		}
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), unredactedString(c.Token, o.Token)})
		}
	case *schema.BitbucketCloudConnection:
		o := oldCfg.(*schema.BitbucketCloudConnection)
		if c.AppPassword != "" {
			edits = append(edits, edit{jsonx.MakePath("appPassword"), unredactedString(c.AppPassword, o.AppPassword)})
		}
	case *schema.AWSCodeCommitConnection:
		o := oldCfg.(*schema.AWSCodeCommitConnection)
		if c.SecretAccessKey != "" {
			edits = append(edits, edit{jsonx.MakePath("secretAccessKey"), unredactedString(c.SecretAccessKey, o.SecretAccessKey)})
		}
		if c.GitCredentials.Password != "" {
			edits = append(edits, edit{jsonx.MakePath("gitCredentials", "password"), unredactedString(c.GitCredentials.Password, o.GitCredentials.Password)})
		}
	case *schema.PhabricatorConnection:
		o := oldCfg.(*schema.PhabricatorConnection)
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), unredactedString(c.Token, o.Token)})
		}
	case *schema.PerforceConnection:
		o := oldCfg.(*schema.PerforceConnection)
		if c.P4Passwd != "" {
			edits = append(edits, edit{jsonx.MakePath("p4.passwd"), unredactedString(c.P4Passwd, o.P4Passwd)})
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

			edits = append(edits, edit{jsonx.MakePath("urls", i), unredacted})
		}
	case *schema.JVMPackagesConnection:
		o := oldCfg.(*schema.JVMPackagesConnection)
		if c.Maven != nil && c.Maven.Credentials != "" && o.Maven != nil {
			edits = append(edits, edit{jsonx.MakePath("maven", "credentials"), unredactedString(c.Maven.Credentials, o.Maven.Credentials)})
		}
	case *schema.PagureConnection:
		o := oldCfg.(*schema.PagureConnection)
		if c.Token != "" {
			edits = append(edits, edit{jsonx.MakePath("token"), unredactedString(c.Token, o.Token)})
		}
	case *schema.NpmPackagesConnection:
		o := oldCfg.(*schema.NpmPackagesConnection)
		if c.Credentials != "" {
			edits = append(edits, edit{jsonx.MakePath("credentials"), unredactedString(c.Credentials, o.Credentials)})
		}
	case *schema.OtherExternalServiceConnection:
		o := oldCfg.(*schema.OtherExternalServiceConnection)
		if c.Url != "" {
			unredacted, err := unredactedURL(c.Url, o.Url)
			if err != nil {
				return err
			}
			edits = append(edits, edit{jsonx.MakePath("url"), unredacted})
		}
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(newCfg))
	}

	unredacted := e.Config
	for _, e := range edits {
		if unredacted, err = e.apply(unredacted); err != nil {
			return err
		}
	}

	e.Config = unredacted
	return nil
}

type edit struct {
	path  jsonx.Path
	value interface{}
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
