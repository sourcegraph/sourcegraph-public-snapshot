// The functions in this file are used to redact secrets from ExternalServices in
// transit, eg when written back and forth between the client and API, as we
// don't want to leak an access token once it's been configured. Any config
// written back from the client with a redacted token should then be updated with
// the real token from the database, the validation in the ExternalService DB
// methods will check for this field and throw an error if it's not been
// replaced, to prevent us accidentally blanking tokens in the DB.

package types

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
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
	var (
		newCfg string
		err    error
	)
	cfg, err := e.Configuration()
	if err != nil {
		return "", err
	}
	switch cfg := cfg.(type) {
	case *schema.GitHubConnection:
		newCfg, err = redactField(e.Config, []string{"token"})
	case *schema.GitLabConnection:
		newCfg, err = redactField(e.Config, []string{"token"})
	case *schema.BitbucketServerConnection:
		// BitbucketServer can have a token OR password
		var fields [][]string
		if cfg.Password != "" {
			fields = append(fields, []string{"password"})
		}
		if cfg.Token != "" {
			fields = append(fields, []string{"token"})
		}
		newCfg, err = redactField(e.Config, fields...)
	case *schema.BitbucketCloudConnection:
		newCfg, err = redactField(e.Config, []string{"appPassword"})
	case *schema.AWSCodeCommitConnection:
		newCfg, err = redactField(e.Config, []string{"secretAccessKey"}, []string{"gitCredentials", "password"})
	case *schema.PhabricatorConnection:
		newCfg, err = redactField(e.Config, []string{"token"})
	case *schema.PerforceConnection:
		newCfg, err = redactField(e.Config, []string{"p4.passwd"})
	case *schema.GitoliteConnection:
		// Gitolite has no secret fields
		newCfg, err = redactField(e.Config)
	case *schema.OtherExternalServiceConnection:
		newCfg, err = redactField(e.Config, []string{"url"})
	case *schema.JVMPackagesConnection:
		newCfg, err = e.Config, nil
	default:
		// return an error here, it's safer to fail than to incorrectly return unsafe data.
		err = errors.Errorf("RedactExternalServiceConfig: kind %q not implemented", e.Kind)
	}
	if err != nil {
		return "", err
	}
	return newCfg, nil
}

// redactField will unmarshal the passed JSON string into the passed value, and then replace the pointer fields you pass
// with RedactedSecret, see RedactExternalServiceConfig for usage examples.
// who needs generics anyway?
func redactField(buf string, paths ...[]string) (string, error) {
	var err error
	for _, path := range paths {
		buf, err = jsonc.Edit(buf, RedactedSecret, path...)
		if err != nil {
			return buf, err
		}
	}
	return buf, nil
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
	var (
		unredacted string
		err        error
	)
	cfg, err := e.Configuration()
	if err != nil {
		return err
	}
	switch cfg := cfg.(type) {
	case *schema.GitHubConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, jsonStringField{[]string{"token"}, &cfg.Token})
	case *schema.GitLabConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, jsonStringField{[]string{"token"}, &cfg.Token})
	case *schema.BitbucketServerConnection:
		// BitbucketServer can have a token OR password
		var fields []jsonStringField
		if cfg.Password != "" {
			fields = append(fields, jsonStringField{[]string{"password"}, &cfg.Password})
		}
		if cfg.Token != "" {
			fields = append(fields, jsonStringField{[]string{"token"}, &cfg.Token})
		}
		unredacted, err = unredactField(old.Config, e.Config, &cfg, fields...)
	case *schema.BitbucketCloudConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, jsonStringField{[]string{"appPassword"}, &cfg.AppPassword})
	case *schema.AWSCodeCommitConnection:
		unredacted, err = unredactField(old.Config,
			e.Config,
			&cfg,
			jsonStringField{[]string{"secretAccessKey"}, &cfg.SecretAccessKey},
			jsonStringField{[]string{"gitCredentials", "password"}, &cfg.GitCredentials.Password},
		)
	case *schema.PhabricatorConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, jsonStringField{[]string{"token"}, &cfg.Token})
	case *schema.PerforceConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, jsonStringField{[]string{"p4.passwd"}, &cfg.P4Passwd})
	case *schema.GitoliteConnection:
		// no secret fields?
		unredacted, err = unredactField(old.Config, e.Config, &cfg)
	case *schema.OtherExternalServiceConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, jsonStringField{[]string{"url"}, &cfg.Url})
	case *schema.JVMPackagesConnection:
		unredacted, err = e.Config, nil
	default:
		// return an error here, it's safer to fail than to incorrectly return unsafe data.
		err = errors.Errorf("UnRedactExternalServiceConfig: kind %q not implemented", e.Kind)
	}
	if err != nil {
		return err
	}
	e.Config = unredacted
	return nil
}

type jsonStringField struct {
	path []string
	ptr  *string
}

func unredactField(old, new string, cfg interface{}, fields ...jsonStringField) (string, error) {
	// first we zero the fields on cfg, as they will contain data we don't need from the e.Configuration() call
	// we just want an empty struct of the correct type for marshaling into
	if err := zeroFields(cfg); err != nil {
		return "", err
	}
	if err := unmarshalConfig(old, cfg); err != nil {
		return "", err
	}

	// and apply edits to update those fields in the new config
	var err error
	for _, field := range fields {
		v, err := jsonc.ReadProperty(new, field.path...)
		if err != nil {
			return new, err
		}
		stringValue, ok := v.(string)
		if !ok {
			return new, errors.Errorf("invalid type %T for field %s", v, field.path)
		}
		if stringValue != RedactedSecret {
			// using unicode zero width space might mean the user includes it when editing still, we strip that out here
			new, err = jsonc.Edit(new, strings.ReplaceAll(stringValue, RedactedSecret, ""), field.path...)
			if err != nil {
				return new, err
			}
			// if the field has been edited we should skip unredaction to allow edits
			continue
		}
		new, err = jsonc.Edit(new, *field.ptr, field.path...)
		if err != nil {
			return new, err
		}
	}

	return new, err
}

// zeroFields zeroes the fields of a struct
func zeroFields(s interface{}) error {
	for _, f := range structs.Fields(s) {
		if f.IsZero() {
			continue
		}
		err := f.Zero()
		if err != nil {
			return err
		}
	}
	return nil
}

// config may contain comments, normalize with jsonc before unmarshaling with jsoniter
func unmarshalConfig(buf string, v interface{}) error {
	normalized, err := jsonc.Parse(buf)
	if err != nil {
		return err
	}
	return jsoniter.Unmarshal(normalized, v)
}
