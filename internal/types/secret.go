// The functions in this file are used to redact secrets from ExternalServices in
// transit, eg when written back and forth between the client and API, as we
// don't want to leak an access token once it's been configured. Any config
// written back from the client with a redacted token should then be updated with
// the real token from the database, the validation in the ExternalService DB
// methods will check for this field and throw an error if it's not been
// replaced, to prevent us accidentally blanking tokens in the DB.

package types

import (
	"reflect"
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
	jsonStringFields, err := redactionInfo(cfg)
	if err != nil {
		return "", err
	}
	paths := [][]string{}
	for _, field := range jsonStringFields {
		paths = append(paths, field.path)
	}
	newCfg, err = redactField(e.Config, paths...)
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

// Return information for redaction, which includes a JSON path, as well as
// a pointer inside the configuration struct. This information can be used to
// redact or unredact arbitrarily nested fields.
//
// Any pointers returned are not guaranteed to be valid across non-trivial
// modifications of the `cfg` parameter, such as zeroing.
func redactionInfo(cfg interface{}) ([]jsonStringField, error) {
	switch cfg := cfg.(type) {
	case *schema.GitHubConnection:
		return []jsonStringField{{[]string{"token"}, &cfg.Token}}, nil
	case *schema.GitLabConnection:
		return []jsonStringField{{[]string{"token"}, &cfg.Token}}, nil
	case *schema.BitbucketServerConnection:
		// BitbucketServer can have a token OR password
		fields := []jsonStringField{}
		if cfg.Password != "" {
			fields = append(fields, jsonStringField{[]string{"password"}, &cfg.Password})
		}
		if cfg.Token != "" {
			fields = append(fields, jsonStringField{[]string{"token"}, &cfg.Token})
		}
		return fields, nil
	case *schema.BitbucketCloudConnection:
		return []jsonStringField{{[]string{"appPassword"}, &cfg.AppPassword}}, nil
	case *schema.AWSCodeCommitConnection:
		return []jsonStringField{
			{[]string{"secretAccessKey"}, &cfg.SecretAccessKey},
			{[]string{"gitCredentials", "password"}, &cfg.GitCredentials.Password},
		}, nil
	case *schema.PhabricatorConnection:
		return []jsonStringField{{[]string{"token"}, &cfg.Token}}, nil
	case *schema.PerforceConnection:
		return []jsonStringField{{[]string{"p4.passwd"}, &cfg.P4Passwd}}, nil
	case *schema.GitoliteConnection:
		return []jsonStringField{}, nil
	case *schema.JVMPackagesConnection:
		return []jsonStringField{{[]string{"maven", "credentials"}, &cfg.Maven.Credentials}}, nil
	case *schema.PagureConnection:
		if cfg.Token != "" {
			return []jsonStringField{{[]string{"token"}, &cfg.Token}}, nil
		}
		return []jsonStringField{}, nil
	case *schema.NPMPackagesConnection:
		return []jsonStringField{{[]string{"credentials"}, &cfg.Credentials}}, nil
	case *schema.OtherExternalServiceConnection:
		return []jsonStringField{{[]string{"url"}, &cfg.Url}}, nil
	default:
		// return an error; it's safer to fail than to incorrectly return unsafe data.
		return nil, errors.Errorf("Unrecognized ExternalServiceConfig for redaction: kind %+v not implemented", reflect.TypeOf(cfg))
	}
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
	unredacted, err = unredactFields(old.Config, e.Config, cfg)
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

func unredactFields(old, new string, cfg interface{}) (string, error) {
	// first we zero the fields on cfg, as they will contain data we don't need from the e.Configuration() call
	// we just want an empty struct of the correct type for marshaling into
	if err := zeroFields(cfg); err != nil {
		return "", err
	}
	if err := unmarshalConfig(old, cfg); err != nil {
		return "", err
	}
	jsonStringFields, err := redactionInfo(cfg)
	if err != nil {
		return "", err
	}

	// and apply edits to update those fields in the new config
	for _, field := range jsonStringFields {
		v, err := jsonc.ReadProperty(new, field.path...)
		if err != nil {
			// This field was deleted, so we skip any edits to it.
			continue
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
