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

type webhookService struct {
	db      database.DB
	keyRing keyring.Ring
}

func NewWebhookService(db database.DB, keyRing keyring.Ring) webhookService {
	return webhookService{
		db:      db,
		keyRing: keyRing,
	}
}

func (ws *webhookService) CreateWebhook(ctx context.Context, codeHostKind, codeHostURN string, secretStr *string) (*types.Webhook, error) {
	err := validateCodeHostKindAndSecret(codeHostKind, secretStr)
	if err != nil {
		return nil, err
	}
	var secret *types.EncryptableSecret
	if secretStr != nil {
		secret = types.NewUnencryptedSecret(*secretStr)
	}
	return ws.db.Webhooks(ws.keyRing.WebhookKey).Create(ctx, codeHostKind, codeHostURN, actor.FromContext(ctx).UID, secret)
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
