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

func (ws *webhookService) CreateWebhook(ctx context.Context, name, codeHostKind, codeHostURN string, secretStr *string) (*types.Webhook, error) {
	err := validateCodeHostKindAndSecret(codeHostKind, secretStr)
	if err != nil {
		return nil, err
	}
	var secret *types.EncryptableSecret
	if secretStr != nil {
		secret = types.NewUnencryptedSecret(*secretStr)
	}
	return ws.db.Webhooks(ws.keyRing.WebhookKey).Create(ctx, name, codeHostKind, codeHostURN, actor.FromContext(ctx).UID, secret)
}

func (ws *webhookService) DeleteWebhook(ctx context.Context, id int32) error {
	return ws.db.Webhooks(ws.keyRing.WebhookKey).Delete(ctx, database.DeleteWebhookOpts{ID: id})
}

func (ws *webhookService) UpdateWebhook(ctx context.Context, id int32, name, codeHostKind, codeHostURN, secret string) (*types.Webhook, error) {
	webhooksStore := ws.db.Webhooks(ws.keyRing.WebhookKey)
	webhook, err := webhooksStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != "" {
		webhook.Name = name
	}
	if codeHostKind != "" {
		if err := validateCodeHostKindAndSecret(codeHostKind, &secret); err != nil {
			return nil, err
		}

		webhook.CodeHostKind = codeHostKind
	}
	if codeHostURN != "" {
		codeHostURN, err := extsvc.NewCodeHostBaseURL(codeHostURN)
		if err != nil {
			return nil, err
		}
		webhook.CodeHostURN = codeHostURN
	}
	if secret != "" {
		if codeHostKind != "" {
			if err := validateCodeHostKindAndSecret(codeHostKind, &secret); err != nil {
				return nil, err
			}
		} else {
			if err := validateCodeHostKindAndSecret(webhook.CodeHostKind, &secret); err != nil {
				return nil, err
			}
		}

		webhook.Secret = types.NewUnencryptedSecret(secret)
	}

	newWebhook, err := webhooksStore.Update(ctx, webhook)
	if err != nil {
		return nil, err
	}
	return newWebhook, nil
}

func validateCodeHostKindAndSecret(codeHostKind string, secret *string) error {
	switch codeHostKind {
	case extsvc.KindGitHub, extsvc.KindGitLab, extsvc.KindBitbucketServer:
		return nil
	case extsvc.KindBitbucketCloud:
		if secret != nil {
			return errors.Newf("webhooks do not support secrets for code host kind %s", codeHostKind)
		}
		return nil
	default:
		return errors.Newf("webhooks are not supported for code host kind %s", codeHostKind)
	}
}
