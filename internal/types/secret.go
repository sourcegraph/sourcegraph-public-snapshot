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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/structs"

	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/schema"
)

const RedactedSecret = "REDACTED"

// RedactExternalServiceConfig replaces any secret fields in the Config field with RedactedSecret, be sure to call
// UnRedactExternalServiceConfig before writing back to the database, otherwise validation will throw errors.
func RedactExternalServiceConfig(svc *ExternalService) error {
	var (
		newCfg string
		err    error
	)
	switch strings.ToUpper(svc.Kind) {
	case extsvc.KindGitHub:
		cfg := &schema.GitHubConnection{}
		newCfg, err = redactField(svc.Config, &cfg, &cfg.Token)
	case extsvc.KindGitLab:
		cfg := schema.GitLabConnection{}
		newCfg, err = redactField(svc.Config, &cfg, &cfg.Token)
	case extsvc.KindBitbucketServer:
		cfg := schema.BitbucketServerConnection{}
		newCfg, err = redactField(svc.Config, &cfg, &cfg.Token, &cfg.Password)
	case extsvc.KindBitbucketCloud:
		cfg := schema.BitbucketCloudConnection{}
		newCfg, err = redactField(svc.Config, &cfg, &cfg.AppPassword)
	case extsvc.KindAWSCodeCommit:
		cfg := schema.AWSCodeCommitConnection{}
		newCfg, err = redactField(svc.Config, &cfg, &cfg.SecretAccessKey)
	case extsvc.KindPhabricator:
		cfg := schema.PhabricatorConnection{}
		newCfg, err = redactField(svc.Config, &cfg, &cfg.Token)
	case extsvc.KindPerforce:
		cfg := schema.PerforceConnection{}
		newCfg, err = redactField(svc.Config, &cfg, &cfg.P4Passwd)
	case extsvc.KindGitolite:
		// no secret fields?
		err = nil
	case extsvc.KindOther:
		// no secret fields?
		err = nil
	default:
		// return an error here, it's safer to fail than to incorrectly return unsafe data.
		err = fmt.Errorf("RedactExternalServiceConfig: kind %q not implemented", svc.Kind)
	}
	if err != nil {
		return err
	}
	svc.Config = newCfg
	return nil
}

// redactField will unmarshal the passed JSON string into the passed value, and then replace the pointer fields you pass
// with RedactedSecret, see RedactExternalServiceConfig for usage examples.
// who needs generics anyway?
func redactField(buf string, v interface{}, fields ...*string) (string, error) {
	err := json.Unmarshal([]byte(buf), v)
	if err != nil {
		return "", err
	}
	for _, field := range fields {
		if *field != "" {
			*field = RedactedSecret
		}
	}
	out, err := json.Marshal(v)
	return string(out), err
}

// UnRedactExternalServiceConfig will replace redacted fields in the 'new' ExternalService with their undredacted form
// from the 'old' ExternalService. You should call this when accepting updated config from a user that may have been
// previously redacted, and pass in the unredacted form directly from the DB as the 'old' parameter
func UnRedactExternalServiceConfig(old, new *ExternalService) error {
	if old.Kind != new.Kind {
		return fmt.Errorf(
			"UnRedactExternalServiceConfig: unmatched external service kinds, old: %q, new: %q",
			old.Kind,
			new.Kind,
		)
	}
	switch strings.ToUpper(new.Kind) {
	case extsvc.KindGitHub:
		cfg := schema.GitHubConnection{}
		unRedacted, err := unRedactField(old.Config, new.Config, &cfg, &cfg.Token)
		if err != nil {
			return err
		}
		new.Config = unRedacted
	case extsvc.KindGitLab:
		cfg := schema.GitLabConnection{}
		unRedacted, err := unRedactField(old.Config, new.Config, &cfg, &cfg.Token)
		if err != nil {
			return err
		}
		new.Config = unRedacted
	case extsvc.KindBitbucketServer:
		cfg := schema.BitbucketServerConnection{}
		unRedacted, err := unRedactField(old.Config, new.Config, &cfg, &cfg.Token, &cfg.Password)
		if err != nil {
			return err
		}
		new.Config = unRedacted
	case extsvc.KindBitbucketCloud:
		cfg := schema.BitbucketCloudConnection{}
		unRedacted, err := unRedactField(old.Config, new.Config, &cfg, &cfg.AppPassword)
		if err != nil {
			return err
		}
		new.Config = unRedacted
	case extsvc.KindAWSCodeCommit:
		cfg := schema.AWSCodeCommitConnection{}
		unRedacted, err := unRedactField(old.Config, new.Config, &cfg, &cfg.SecretAccessKey)
		if err != nil {
			return err
		}
		new.Config = unRedacted
	case extsvc.KindPhabricator:
		cfg := schema.PhabricatorConnection{}
		unRedacted, err := unRedactField(old.Config, new.Config, &cfg, &cfg.Token)
		if err != nil {
			return err
		}
		new.Config = unRedacted
	case extsvc.KindPerforce:
		cfg := schema.PerforceConnection{}
		unRedacted, err := unRedactField(old.Config, new.Config, &cfg, &cfg.P4Passwd)
		if err != nil {
			return err
		}
		new.Config = unRedacted
	case extsvc.KindGitolite:
		// no secret fields?
		return nil
	case extsvc.KindOther:
		// no secret fields?
		return nil
	default:
		// return an error here, it's safer to fail than to incorrectly return unsafe data.
		return fmt.Errorf("UnRedactExternalServiceConfig: kind %q not implemented", new.Kind)
	}
	return nil
}

func unRedactField(old, new string, cfg interface{}, fields ...*string) (string, error) {
	err := json.Unmarshal([]byte(old), cfg)
	if err != nil {
		return "", err
	}
	// first take copies of the unredacted fields from the old JSON
	oldSecrets := []string{}
	for _, field := range fields {
		oldSecrets = append(oldSecrets, *field)
	}
	// zero the fields of our config in case we are deleting any fields, the unmarshaler might preserve
	// fields from the old JSON otherwise.
	err = zeroFields(cfg)
	if err != nil {
		return "", err
	}
	// now we unmarshal the new JSON that contains redacted fields
	err = json.Unmarshal([]byte(new), cfg)
	for i, field := range fields {
		// only replace fields that are RedactedSecret, so the user can still update their config
		if *field == RedactedSecret {
			*field = oldSecrets[i]
		}
	}

	// marshal the output, now containing the new json with redacted fields replaced with fields from the old JSON
	out, err := json.Marshal(cfg)
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
