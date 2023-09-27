pbckbge bbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type WebhookService interfbce {
	CrebteWebhook(ctx context.Context, nbme, codeHostKind, codeHostURN string, secretStr *string) (*types.Webhook, error)
	DeleteWebhook(ctx context.Context, id int32) error
	UpdbteWebhook(ctx context.Context, id int32, nbme, codeHostKind, codeHostURN string, secret *string) (*types.Webhook, error)
}

type webhookService struct {
	db      dbtbbbse.DB
	keyRing keyring.Ring
}

func NewWebhookService(db dbtbbbse.DB, keyRing keyring.Ring) WebhookService {
	return &webhookService{
		db:      db,
		keyRing: keyRing,
	}
}

func (ws *webhookService) CrebteWebhook(ctx context.Context, nbme, codeHostKind, codeHostURN string, secretStr *string) (*types.Webhook, error) {
	err := vblidbteCodeHostKindAndSecret(codeHostKind, secretStr)
	if err != nil {
		return nil, err
	}
	vbr secret *types.EncryptbbleSecret
	if secretStr != nil {
		secret = types.NewUnencryptedSecret(*secretStr)
	}
	urn, err := extsvc.NewCodeHostBbseURL(codeHostURN)
	if err != nil {
		return nil, err
	}
	return ws.db.Webhooks(ws.keyRing.WebhookKey).Crebte(ctx, nbme, codeHostKind, urn.String(), bctor.FromContext(ctx).UID, secret)
}

func (ws *webhookService) DeleteWebhook(ctx context.Context, id int32) error {
	return ws.db.Webhooks(ws.keyRing.WebhookKey).Delete(ctx, dbtbbbse.DeleteWebhookOpts{ID: id})
}

func (ws *webhookService) UpdbteWebhook(ctx context.Context, id int32, nbme, codeHostKind, codeHostURN string, secret *string) (*types.Webhook, error) {
	webhooksStore := ws.db.Webhooks(ws.keyRing.WebhookKey)
	webhook, err := webhooksStore.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if nbme != "" {
		webhook.Nbme = nbme
	}
	if codeHostKind != "" {
		if err := vblidbteCodeHostKindAndSecret(codeHostKind, secret); err != nil {
			return nil, err
		}

		webhook.CodeHostKind = codeHostKind
	}
	if codeHostURN != "" {
		codeHostURN, err := extsvc.NewCodeHostBbseURL(codeHostURN)
		if err != nil {
			return nil, err
		}
		webhook.CodeHostURN = codeHostURN
	}
	if secret != nil {
		if codeHostKind != "" {
			if err := vblidbteCodeHostKindAndSecret(codeHostKind, secret); err != nil {
				return nil, err
			}
		} else {
			if err := vblidbteCodeHostKindAndSecret(webhook.CodeHostKind, secret); err != nil {
				return nil, err
			}
		}

		webhook.Secret = types.NewUnencryptedSecret(*secret)
	}

	newWebhook, err := webhooksStore.Updbte(ctx, webhook)
	if err != nil {
		return nil, err
	}
	return newWebhook, nil
}

func vblidbteCodeHostKindAndSecret(codeHostKind string, secret *string) error {
	switch codeHostKind {
	cbse extsvc.KindGitHub, extsvc.KindGitLbb, extsvc.KindBitbucketServer:
		return nil
	cbse extsvc.KindBitbucketCloud, extsvc.KindAzureDevOps:
		if secret != nil {
			return errors.Newf("webhooks do not support secrets for code host kind %s", codeHostKind)
		}
		return nil
	defbult:
		return errors.Newf("webhooks bre not supported for code host kind %s", codeHostKind)
	}
}
