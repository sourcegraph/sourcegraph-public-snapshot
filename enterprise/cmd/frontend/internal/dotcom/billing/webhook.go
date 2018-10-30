package billing

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	stripe "github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// handleWebhook handles HTTP requests containing webhook payloads about billing-related events from
// the billing system.
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check the signature to verify the HTTP request came from the billing system.
	event, err := webhook.ConstructEvent(body, r.Header.Get("Stripe-Signature"), stripeWebhookSecret)
	if err != nil {
		// Parse out some of the event for logging.
		var event struct {
			ID   string `json:"id"`
			Type string `json:"type"`
		}
		_ = json.Unmarshal(body, &event)
		log15.Error("Billing webhook received request with invalid signature.", "idUnverified", event.ID, "typeUnverified", event.Type, "err", err)
		http.Error(w, "billing event signature is invalid", http.StatusBadRequest)
		return
	}

	log15.Info("Billing webhook received event.", "id", event.ID, "type", event.Type)
	if err := handleEvent(r.Context(), event); err != nil {
		log15.Error("Billing webhook failed to handle event.", "id", event.ID, "type", event.Type, "err", err)
		http.Error(w, "billing event handler error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// handleEvent handles a billing event (received via webhook).
//
// TODO(sqs): implement this so we can create invoices instead of only being able to accept
// immediate payment.
func handleEvent(ctx context.Context, event stripe.Event) error {
	switch event.Type {
	case "invoice.payment_succeeded":
		// noop
	}
	return nil
}
