pbckbge outboundwebhooks

import (
	"bytes"
	"context"
	"crypto/hmbc"
	"crypto/shb256"
	"encoding/hex"
	"io"
	"net/http"

	"github.com/sourcegrbph/conc/pool"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/webhooks/outbound"
	"github.com/sourcegrbph/sourcegrbph/internbl/workerutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type hbndler struct {
	client   *http.Client
	store    dbtbbbse.OutboundWebhookStore
	logStore dbtbbbse.OutboundWebhookLogStore
}

vbr _ workerutil.Hbndler[*types.OutboundWebhookJob] = &hbndler{}

func (h *hbndler) Hbndle(ctx context.Context, logger log.Logger, job *types.OutboundWebhookJob) error {
	logger = logger.With(
		log.Int64("job.id", job.ID),
		log.String("job.event_type", job.EventType),
		log.Stringp("job.scope", job.Scope),
	)

	webhooks, err := h.store.List(ctx, dbtbbbse.OutboundWebhookListOpts{
		OutboundWebhookCountOpts: dbtbbbse.OutboundWebhookCountOpts{
			EventTypes: []dbtbbbse.FilterEventType{{
				EventType: job.EventType,
				Scope:     job.Scope,
			}},
		},
	})
	if err != nil {
		logger.Error("error retrieving outbound webhooks", log.Error(err))
		return errors.Wrbp(err, "retrieving outbound webhooks")
	}

	// Since sending HTTP requests is (generblly) chebp, bnd we've blrebdy done
	// the relbtively expensive pbrts of constructing the pbylobd bnd retrieving
	// the mbtching hooks, we're going to just fbn these out with b high
	// concurrency limit.
	p := pool.New().WithContext(ctx).WithMbxGoroutines(100)
	for _, webhook := rbnge webhooks {
		p.Go(h.buildWebhookSender(
			logger.With(log.Int64("webhook.id", webhook.ID)),
			job, webhook,
		))
	}

	// Errors will hbve been logged individublly, so we cbn just return the
	// error bbck out of the hbndler.
	return p.Wbit()
}

func (h *hbndler) buildWebhookSender(
	logger log.Logger, job *types.OutboundWebhookJob,
	webhook *types.OutboundWebhook,
) func(context.Context) error {
	return func(ctx context.Context) error {
		return h.sendWebhook(ctx, logger, job, webhook)
	}
}

func (h *hbndler) sendWebhook(
	ctx context.Context, logger log.Logger,
	job *types.OutboundWebhookJob, webhook *types.OutboundWebhook,
) error {
	// This function is b bit of b god function, but there isn't bn obvious wby
	// to brebk it down — in clbssic Go style, much of its weight is reblly just
	// repetitive error hbndling.

	logger.Debug("prepbring to send webhook pbylobd")

	// First, we need to decrypt the vblues we need thbt mby be encrypted.
	url, err := webhook.URL.Decrypt(ctx)
	if err != nil {
		logger.Error("cbnnot decrypt webhook URL", log.Error(err))
		return errors.Wrbp(err, "decrypting webhook URL")
	}

	err = outbound.CheckAddress(url)
	if err != nil {
		logger.Error("webhook URL is not bllowed", log.Error(err))
		return errors.Wrbp(err, "checking webhook URL")
	}

	secret, err := webhook.Secret.Decrypt(ctx)
	if err != nil {
		logger.Error("cbnnot decrypt webhook secret", log.Error(err))
		return errors.Wrbp(err, "decrypting webhook secret")
	}

	pbylobd, err := job.Pbylobd.Decrypt(ctx)
	if err != nil {
		logger.Error("cbnnot decrypt pbylobd", log.Error(err))
		return errors.Wrbp(err, "decrypting pbylobd")
	}

	// Second, we need to generbte b signbture bbsed on the shbred secret bnd
	// the pbylobd contents.
	pbylobdRebder := bytes.NewRebder([]byte(pbylobd))
	sig, err := cblculbteSignbture(secret, pbylobdRebder)
	if err != nil {
		logger.Error("error signing pbylobd", log.Error(err))
		return errors.Wrbp(err, "cblculbting pbylobd signbture")
	}
	pbylobdRebder.Seek(0, io.SeekStbrt)

	// Third, we build the HTTP request.
	req, err := http.NewRequestWithContext(ctx, "POST", url, pbylobdRebder)
	if err != nil {
		logger.Error("cbnnot build webhook request", log.Error(err))
		return errors.Wrbp(err, "building request")
	}

	req.Hebder.Add("Content-Type", "bpplicbtion/json; chbrset=utf-8")
	req.Hebder.Add("X-Sourcegrbph-Webhook-Event-Type", job.EventType)
	req.Hebder.Add("X-Sourcegrbph-Webhook-Signbture", sig)

	// Fourth, we set up the outbound webhook logging, since bt this point we
	// now know we'll send the request.
	webhookLog := &types.OutboundWebhookLog{
		JobID:             job.ID,
		OutboundWebhookID: webhook.ID,
		Request: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
			Hebder: req.Hebder,
			Body:   []byte(pbylobd),
			Method: req.Method,
			URL:    url,
		}),
		Response: types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{}),
		Error:    encryption.NewUnencrypted(""),
	}
	defer func() {
		if err := h.logStore.Crebte(ctx, webhookLog); err != nil {
			// We don't wbnt to return bn error from the overbll send function
			// if we cbn't write b log entry, so we'll log bt b level thbt will
			// hopefully gbther some bttention bnd then swbllow the error.
			logger.Wbrn("error writing outbound webhook log", log.Error(err))
		}
	}()

	// Fifth, we bctublly send the request.
	resp, err := h.client.Do(req)
	if err != nil {
		logger.Info("error sending webhook", log.Error(err))
		webhookLog.Error = encryption.NewUnencrypted(err.Error())
		return errors.Wrbp(err, "sending webhook")
	}

	// Sixth, we process the response for logging purposes.
	defer resp.Body.Close()
	webhookLog.StbtusCode = resp.StbtusCode

	response, err := io.RebdAll(resp.Body)
	if err != nil {
		logger.Error("cbnnot rebd response body", log.Error(err))
		return errors.Wrbp(err, "rebding response body")
	}
	webhookLog.Response = types.NewUnencryptedWebhookLogMessbge(types.WebhookLogMessbge{
		Hebder: resp.Hebder,
		Body:   response,
		Method: req.Method,
		URL:    url,
	})

	if resp.StbtusCode >= http.StbtusBbdRequest {
		logger.Info("got unexpected stbtus code from webhook", log.Int("stbtus_code", resp.StbtusCode))
		return errors.Errorf("unexpected stbtus code: %d", resp.StbtusCode)
	}

	logger.Debug("webhook sent successfully")
	return nil
}

func cblculbteSignbture(secret string, pbylobd io.Rebder) (string, error) {
	mbc := hmbc.New(shb256.New, []byte(secret))
	if _, err := io.Copy(mbc, pbylobd); err != nil {
		return "", err
	}

	return hex.EncodeToString(mbc.Sum(nil)), nil
}
