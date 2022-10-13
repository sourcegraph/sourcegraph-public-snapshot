package backend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func CreateWebhook(ctx context.Context, db database.DB, codeHostKind, codeHostURN string, secretStr *string) (*types.Webhook, error) {
	err := validateCodeHostKindAndSecret(codeHostKind, secretStr)
	if err != nil {
		return nil, err
	}
	var secret *types.EncryptableSecret
	if secret != nil {
		secret = types.NewUnencryptedSecret(*secretStr)
	}
	return db.Webhooks(keyring.Default().WebhookKey).Create(ctx, codeHostKind, codeHostURN, actor.FromContext(ctx).UID, secret)
}

func validateCodeHostKindAndSecret(codeHostKind string, secret *string) error {
	switch codeHostKind {
	case extsvc.KindGitHub, extsvc.KindGitLab:
		return nil
	case extsvc.KindBitbucketCloud, extsvc.KindBitbucketServer:
		if secret != nil {
			return errors.Newf("webhooks do not support secrets for code host kind %s", codeHostKind)
		}
		return nil
	default:
		return errors.Newf("webhooks are not supported for code host kind %s", codeHostKind)
	}
}
