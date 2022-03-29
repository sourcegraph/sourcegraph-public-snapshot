// The functions in this file are used to redact secrets from ExternalServices in
// transit, eg when written back and forth between the client and API, as we
// don't want to leak an access token once it's been configured. Any config
// written back from the client with a redacted token should then be updated with
// the real token from the database, the validation in the ExternalService DB
// methods will check for this field and throw an error if it's not been
// replaced, to prevent us accidentally blanking tokens in the DB.

package types

import (
	"bytes"
	"net/url"
	"reflect"
	"strings"

	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/jsonx"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
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

	fields, err := redactionInfo(cfg)
	if err != nil {
		return "", err
	}

	newCfg := e.Config
	for _, field := range fields {
		if newCfg, err = field.Redact(newCfg); err != nil {
			return "", err
		}
	}
	return newCfg, nil
}

// redactionInfo returns redactableFields for the given config.
func redactionInfo(cfg interface{}) ([]redactableField, error) {
	switch cfg := cfg.(type) {
	case *schema.GitHubConnection:
		return []redactableField{jsonStringField{[]string{"token"}, &cfg.Token}}, nil
	case *schema.GitLabConnection:
		return []redactableField{jsonStringField{[]string{"token"}, &cfg.Token}}, nil
	case *schema.BitbucketServerConnection:
		// BitbucketServer can have a token OR password
		var fields []redactableField
		if cfg.Password != "" {
			fields = append(fields, jsonStringField{[]string{"password"}, &cfg.Password})
		}
		if cfg.Token != "" {
			fields = append(fields, jsonStringField{[]string{"token"}, &cfg.Token})
		}
		return fields, nil
	case *schema.BitbucketCloudConnection:
		return []redactableField{jsonStringField{[]string{"appPassword"}, &cfg.AppPassword}}, nil
	case *schema.AWSCodeCommitConnection:
		return []redactableField{
			jsonStringField{[]string{"secretAccessKey"}, &cfg.SecretAccessKey},
			jsonStringField{[]string{"gitCredentials", "password"}, &cfg.GitCredentials.Password},
		}, nil
	case *schema.PhabricatorConnection:
		return []redactableField{jsonStringField{[]string{"token"}, &cfg.Token}}, nil
	case *schema.PerforceConnection:
		return []redactableField{jsonStringField{[]string{"p4.passwd"}, &cfg.P4Passwd}}, nil
	case *schema.GitoliteConnection:
		return []redactableField{}, nil
	case *schema.GoModulesConnection:
		return []redactableField{stringArrayField{
			path:  jsonx.MakePath("urls"),
			field: &cfg.Urls,
			redactor: func(path jsonx.Path, field *string) redactableField {
				return urlField{path, field}
			},
		}}, nil
	case *schema.JVMPackagesConnection:
		return []redactableField{jsonStringField{[]string{"maven", "credentials"}, &cfg.Maven.Credentials}}, nil
	case *schema.PagureConnection:
		if cfg.Token != "" {
			return []redactableField{jsonStringField{[]string{"token"}, &cfg.Token}}, nil
		}
		return []redactableField{}, nil
	case *schema.NpmPackagesConnection:
		return []redactableField{jsonStringField{[]string{"credentials"}, &cfg.Credentials}}, nil
	case *schema.OtherExternalServiceConnection:
		return []redactableField{jsonStringField{[]string{"url"}, &cfg.Url}}, nil
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

type redactableField interface {
	Redact(input string) (output string, err error)
	Unredact(input string) (output string, err error)
	String() string
}

type stringArrayField struct {
	path     jsonx.Path
	field    *[]string
	redactor func(path jsonx.Path, field *string) redactableField
}

func (f stringArrayField) Redact(input string) (string, error) {
	return f.apply(input, func(input string, field redactableField) (string, error) {
		return field.Redact(input)
	})
}

func (f stringArrayField) Unredact(input string) (string, error) {
	return f.apply(input, func(input string, field redactableField) (string, error) {
		return field.Unredact(input)
	})
}

func (f stringArrayField) String() string {
	var b bytes.Buffer
	b.WriteString("[")
	for _, f := range *f.field {
		b.WriteString(`"` + f + `"`)
		b.WriteString(",")
	}

	if b.Len() > 1 {
		b.Truncate(b.Len() - 1)
	}

	b.WriteString("]")
	return b.String()
}

func (f stringArrayField) apply(input string, op func(string, redactableField) (string, error)) (string, error) {
	node, err := jsonc.ReadPath(input, f.path)
	if err != nil {
		// This field was deleted, so we skip any edits to it.
		return input, nil
	}

	if node.Type != jsonx.Array {
		return input, errors.Errorf("invalid type %T for field %v", node.Type, f.path)
	}

	if n := len(node.Children) - len(*f.field); n > 0 {
		*f.field = append(*f.field, make([]string, n)...)
	}

	for i := range node.Children {
		var path jsonx.Path
		path = append(path, f.path...)
		path = append(path, jsonx.Segment{Index: i})

		if input, err = op(input, f.redactor(path, &(*f.field)[i])); err != nil {
			return input, err
		}
	}

	return input, nil
}

type urlField struct {
	path  jsonx.Path
	field *string
}

func (f urlField) Redact(input string) (string, error) {
	redacted, err := url.Parse(*f.field)
	if err != nil {
		return input, errors.Wrapf(err, "failed parsing url field")
	}

	if redacted.User == nil {
		return input, nil
	}

	if _, ok := redacted.User.Password(); !ok {
		return input, nil
	}

	redacted.User = url.UserPassword(redacted.User.Username(), RedactedSecret)
	return jsonc.EditPath(input, redacted.String(), f.path)
}

func (f urlField) Unredact(input string) (string, error) {
	node, err := jsonc.ReadPath(input, f.path)
	if err != nil {
		// This field was deleted, so we skip any edits to it.
		return input, nil
	}

	stringValue, ok := node.Value.(string)
	if !ok {
		return input, errors.Errorf("invalid type %T for field %s", node.Value, f.path)
	}

	newURL, err := url.Parse(stringValue)
	if err != nil {
		return input, errors.Wrap(err, "failed to parse new url")
	}

	oldURL, err := url.Parse(*f.field)
	if err != nil {
		return input, errors.Wrap(err, "failed to parse old url")
	}

	if passwd, ok := newURL.User.Password(); ok && passwd == RedactedSecret {
		oldPasswd, _ := oldURL.User.Password()
		if oldPasswd != "" {
			newURL.User = url.UserPassword(newURL.User.Username(), oldPasswd)
		} else {
			newURL.User = url.User(newURL.User.Username())
		}
	}

	return jsonc.EditPath(input, newURL.String(), f.path)
}

func (f urlField) String() string {
	if f.field == nil {
		return ""
	}
	return *f.field
}

type jsonStringField struct {
	path  []string
	field *string
}

// Redact will unmarshal the passed JSON string into the passed value, and then replace the pointer fields you pass
// with RedactedSecret, see RedactExternalServiceConfig for usage examples.
// who needs generics anyway?
func (f jsonStringField) Redact(input string) (string, error) {
	return jsonc.Edit(input, RedactedSecret, f.path...)
}

func (f jsonStringField) Unredact(input string) (string, error) {
	v, err := jsonc.ReadProperty(input, f.path...)
	if err != nil {
		// This field was deleted, so we skip any edits to it.
		return input, nil
	}

	stringValue, ok := v.(string)
	if !ok {
		return input, errors.Errorf("invalid type %T for field %s", v, f.path)
	}

	// if the field has been edited we should skip unredaction to allow edits
	if stringValue != RedactedSecret {
		// using unicode zero width space might mean the user includes it when editing still, we strip that out here
		return jsonc.Edit(input, strings.ReplaceAll(stringValue, RedactedSecret, ""), f.path...)
	}

	return jsonc.Edit(input, *f.field, f.path...)
}

func (f jsonStringField) String() string {
	if f.field == nil {
		return ""
	}
	return *f.field
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
		if new, err = field.Unredact(new); err != nil {
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
