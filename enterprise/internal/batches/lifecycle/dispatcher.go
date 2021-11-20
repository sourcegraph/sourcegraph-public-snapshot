package lifecycle

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"net/http"

	"github.com/inconshreveable/log15"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/batches/store"
	btypes "github.com/sourcegraph/sourcegraph/enterprise/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/actor"
)

type Dispatcher struct {
	store *store.Store

	c chan event
}

func NewDispatcher(store *store.Store) *Dispatcher {
	c := make(chan event, 0)

	go func() {
		// Since this is an internal process, this uses an internal actor.
		ctx := actor.WithInternalActor(context.Background())

		for e := range c {
			if err := dispatchEvent(ctx, store, e); err != nil {
				log15.Error("error dispatching lifecycle hook event", "err", err, "event", e)
			}
		}
	}()

	return &Dispatcher{store: store, c: c}
}

func (d *Dispatcher) ChangesetPublished(ctx context.Context, cs *btypes.Changeset) {
	if d != nil {
		d.c <- &changesetEvent{
			store:     d.store,
			verb:      "published",
			changeset: cs,
		}
	}
}

func dispatchEvent(ctx context.Context, s *store.Store, e event) error {
	payload, err := e.MarshalPayload(ctx)
	if err != nil {
		return errors.Wrap(err, "marshalling event payload")
	}

	// TODO: filter webhooks based on the methods to be added to the event.
	//
	// TODO: memoise this return value so we don't have to go back to the
	// database every time.
	hooks, _, err := s.ListLifecycleHooks(ctx, store.ListLifecycleHookOpts{})
	if err != nil {
		return errors.Wrap(err, "getting lifecycle hooks")
	}

	for _, hook := range hooks {
		go func(hook *btypes.LifecycleHook) {
			// Calculate signature.
			hasher := hmac.New(sha256.New, []byte(hook.Secret))
			if _, err := hasher.Write(payload); err != nil {
				log15.Error("calcuating signature for payload", "err", err, "hook", hook)
				return
			}
			sig := hasher.Sum(nil)

			// Construct and send request.
			buf := &bytes.Buffer{}
			if _, err := buf.Write(payload); err != nil {
				log15.Error("constructing request buffer", "err", err, "hook", hook)
				return
			}

			req, err := http.NewRequestWithContext(ctx, "POST", hook.URL, buf)
			if err != nil {
				log15.Error("constructing request", "err", err, "hook", hook)
			}

			req.Header.Add("Content-Type", "application/json; charset=UTF-8")
			req.Header.Add("X-Sourcegraph-Signature-256", base64.StdEncoding.EncodeToString(sig))

			if resp, err := http.DefaultClient.Do(req); err != nil {
				log15.Error("sending lifecycle hook", "err", err, "hook", hook)
			} else if resp.StatusCode >= 400 {
				log15.Warn("received error from lifecycle hook endpoint", "response", resp, "hook", hook)
			} else if resp.StatusCode >= 300 {
				// The default redirect policy should make this unlikely.
				log15.Info("received redirect from lifecycle hook endpoint", "response", resp, "hook", hook)
			} else {
				log15.Debug("successful lifecycle hook", "response", resp, "hook", hook)
			}
		}(hook)
	}

	return nil
}
