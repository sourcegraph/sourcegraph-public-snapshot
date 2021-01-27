/*
	The functions in this file are used to redact secrets from ExternalServices in transit, eg when written back and
	forth between the client and API, as we don't want to leak an access token once it's been configured. Any config
	written back from the client with a redacted token should then be updated with the real token from the database,
	the validation in the ExternalService DB methods will check for this field and throw an error if it's not been
	replaced, to prevent us accidentally blanking tokens in the DB.
	This is risky, hacky, and ugly, and we fully intend to replace it ASAP, once our Vault tooling is ready we will
	migrate external services into their own tables for each kind, and encrypt secrets using Vault's KMS.
	if you wanna speak to someone about this talk to @arussellsaw or the Cloud team.
*/

package types

import (
	"fmt"

	"github.com/fatih/structs"
	jsoniter "github.com/json-iterator/go"

	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
)

const RedactedSecret = "SRCGRAPH_REDACTED"

// RedactConfig replaces any secret fields in the Config field with RedactedSecret, be sure to call
// UnRedactExternalServiceConfig before writing back to the database, otherwise validation will throw errors.
func (e *ExternalService) RedactConfig() error {
	var (
		newCfg string
		err    error
	)
	cfg, err := e.Configuration()
	if err != nil {
		return err
	}
	switch cfg := cfg.(type) {
	case *schema.GitHubConnection:
		newCfg, err = redactField(e.Config, &cfg, &cfg.Token)
	case *schema.GitLabConnection:
		newCfg, err = redactField(e.Config, &cfg, &cfg.Token)
	case *schema.BitbucketServerConnection:
		newCfg, err = redactField(e.Config, &cfg, &cfg.Token, &cfg.Password)
	case *schema.BitbucketCloudConnection:
		newCfg, err = redactField(e.Config, &cfg, &cfg.AppPassword)
	case *schema.AWSCodeCommitConnection:
		newCfg, err = redactField(e.Config, &cfg, &cfg.SecretAccessKey)
	case *schema.PhabricatorConnection:
		newCfg, err = redactField(e.Config, &cfg, &cfg.Token)
	case *schema.PerforceConnection:
		newCfg, err = redactField(e.Config, &cfg, &cfg.P4Passwd)
	case *schema.GitoliteConnection:
		// no secret fields?
		err = nil
	case *schema.OtherExternalServiceConnection:
		// no secret fields?
		err = nil
	default:
		// return an error here, it's safer to fail than to incorrectly return unsafe data.
		err = fmt.Errorf("RedactExternalServiceConfig: kind %q not implemented", e.Kind)
	}
	if err != nil {
		return err
	}
	e.Config = newCfg
	return nil
}

// redactField will unmarshal the passed JSON string into the passed value, and then replace the pointer fields you pass
// with RedactedSecret, see RedactExternalServiceConfig for usage examples.
// who needs generics anyway?
func redactField(buf string, v interface{}, fields ...*string) (string, error) {
	err := unmarshalConfig(buf, v)
	if err != nil {
		return "", err
	}
	for _, field := range fields {
		if *field != "" {
			*field = RedactedSecret
		}
	}
	out, err := jsoniter.Marshal(v)
	return string(out), err
}

// UnredactConfig will replace redacted fields with their undredacted form from the 'old' ExternalService.
// You should call this when accepting updated config from a user that may have been
// previously redacted, and pass in the unredacted form directly from the DB as the 'old' parameter
func (e *ExternalService) UnredactConfig(old *ExternalService) error {
	if old.Kind != e.Kind {
		return fmt.Errorf(
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
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.Token)
	case *schema.GitLabConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.Token)
	case *schema.BitbucketServerConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.Token, &cfg.Password)
	case *schema.BitbucketCloudConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.AppPassword)
	case *schema.AWSCodeCommitConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.SecretAccessKey)
	case *schema.PhabricatorConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.Token)
	case *schema.PerforceConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.P4Passwd)
	case *schema.GitoliteConnection:
		// no secret fields?
		err = nil
	case *schema.OtherExternalServiceConnection:
		unredacted, err = unredactField(old.Config, e.Config, &cfg, &cfg.Url)
	default:
		// return an error here, it's safer to fail than to incorrectly return unsafe data.
		err = fmt.Errorf("UnRedactExternalServiceConfig: kind %q not implemented", e.Kind)
	}
	if err != nil {
		return err
	}
	e.Config = unredacted
	return nil
}

func unredactField(old, new string, cfg interface{}, fields ...*string) (string, error) {
	// first we zero the fields on cfg, as they will contain data we don't need from the e.Configuration() call
	// we just want an empty struct of the correct type for marshaling into
	if err := zeroFields(cfg); err != nil {
		return "", err
	}
	if err := unmarshalConfig(old, cfg); err != nil {
		return "", err
	}
	// now take copies of the unredacted fields from the old JSON
	oldSecrets := []string{}
	for _, field := range fields {
		oldSecrets = append(oldSecrets, *field)
	}
	// zero the fields of our config in case we are deleting any fields, the unmarshaler might preserve
	// fields from the old JSON otherwise.
	if err := zeroFields(cfg); err != nil {
		return "", err
	}
	// now we unmarshal the new JSON that contains redacted fields
	if err := unmarshalConfig(new, cfg); err != nil {
		return "", err
	}
	for i, field := range fields {
		// only replace fields that are RedactedSecret, so the user can still update their config
		if *field == RedactedSecret {
			*field = oldSecrets[i]
		}
	}

	// marshal the output, now containing the new json with redacted fields replaced with fields from the old JSON
	out, err := jsoniter.Marshal(cfg)
	return string(out), err
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
