package v1

import (
	"context"
	"encoding/json"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/lib/background"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/atomic"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

type subscriber struct {
	logger       log.Logger
	handlers     SubscriberHandlers
	subscription *pubsub.Subscription

	// state indicates the state of workers.
	state state
	// cancelContext is the function to cancel the context of the receiver that
	// effectively stops the receiver.
	cancelContext context.CancelFunc
}

type SubscriberOptions struct {
	// ProjectID is the GCP project ID that the Pub/Sub subscription belongs to. It
	// is almost always the same GCP project that the Cloud Run service is deployed
	// to.
	ProjectID string
	// SubscriptionID is the GCP Pub/Sub subscription ID to receive SAMS
	// notifications from.
	SubscriptionID string
	// ReceiveSettings is the settings for receiving messages of the subscription. A
	// zero value means to use the default settings.
	ReceiveSettings ReceiveSettings
	// Handlers is the collection of subscription handlers for each type of SAMS
	// notifications.
	Handlers SubscriberHandlers
	// Credentials is the account credentials to be used for the GCP Pub/Sub client.
	// Default credentials will be used when not set.
	Credentials *google.Credentials
}

func (opts SubscriberOptions) Validate() error {
	if opts.ProjectID == "" {
		return errors.New("ProjectID is required")
	}
	if opts.SubscriptionID == "" {
		return errors.New("SubscriptionID is required")
	}
	return nil
}

// NewSubscriber creates a new background routine for receiving SAMS
// notifications from given GCP project ID and Pub/Sub subscription ID.
//
// Users should prefer to use the top-level 'sams.NewNotificationsV1Subscriber'
// constructor instead.
func NewSubscriber(logger log.Logger, opts SubscriberOptions) (background.Routine, error) {
	if err := opts.Validate(); err != nil {
		return nil, err
	}

	logger = logger.Scoped("notification.subscriber")
	client, err := pubsub.NewClient(context.Background(), opts.ProjectID, option.WithCredentials(opts.Credentials))
	if err != nil {
		return nil, errors.Wrap(err, "create GCP Pub/Sub client")
	}
	subscription := client.Subscription(opts.SubscriptionID)
	subscription.ReceiveSettings = opts.ReceiveSettings
	return &subscriber{
		logger:       logger,
		handlers:     opts.Handlers,
		subscription: subscription,
		state:        newState(),
		cancelContext: func() {
			logger.Error("cancelContext is not set")
		},
	}, nil
}

func (s *subscriber) Name() string {
	return "SAMS Notifications Subscriber"
}

func (s *subscriber) Start() {
	if err := s.state.toStarted(); err != nil {
		panic(fmt.Sprintf("failed to start subscriber: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancelContext = cancel
	err := s.subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		var span trace.Span
		ctx, span = tracer.Start(ctx, "subscriber.Receive",
			trace.WithAttributes(attribute.String("msg.id", msg.ID)))
		if msg.DeliveryAttempt != nil {
			span.SetAttributes(attribute.Int("msg.deliveryAttempt", *msg.DeliveryAttempt))
		}

		// Create a logger with important context. Only log messages related to
		// this message handling with this logger.
		logger := s.logger.
			With(
				log.String("msg.id", msg.ID),
				log.Intp("msg.deliveryAttempt", msg.DeliveryAttempt),
			).
			WithTrace(log.TraceContext{
				TraceID: span.SpanContext().TraceID().String(),
				SpanID:  span.SpanContext().SpanID().String(),
			})

		var msgData struct {
			Name     string          `json:"name"`
			Metadata json.RawMessage `json:"metadata"`
		}
		err := json.Unmarshal(msg.Data, &msgData)
		if err != nil {
			logger.Error("failed to unmarshal notification message", log.Error(err))
			msg.Nack()
			return
		}

		status, err := s.handleReceive(ctx, msgData.Name, msgData.Metadata)

		logger = logger.With(
			log.String("msg.name", msgData.Name),
			log.String("handleReceive.status", status))
		span.SetAttributes(
			attribute.String("msg.name", msgData.Name),
			attribute.String("handleReceive.status", status),
		)

		if err == nil {
			if status == handleReceiveStatusUnknownMessage {
				logger.Warn("acknowledging unknown notification name")
			} else {
				logger.Debug("message processed")
			}
			msg.Ack()
		} else {
			logger.Error("failed to process notification message", log.Error(err))
			msg.Nack()
		}
	})
	if err != nil {
		s.logger.Error("failed to receive notifications", log.Error(err))
		return
	}
}

func (s *subscriber) Stop(context.Context) error {
	if err := s.state.toStopped(); err != nil {
		panic(fmt.Sprintf("failed to stop subscriber: %v", err))
	}

	s.cancelContext()
	s.logger.Info("subscriber stopped")
	return nil
}

// SubscriberHandlers is a collection of subscription handlers for each type of
// SAMS notifications. If the handler of a notification is nil, the notification
// will be acknowledged automatically without any processing.
//
// If a handler returns an error, the notification will be unacknowledged and
// retried later.
type SubscriberHandlers struct {
	// OnUserDeleted is called when a "UserDeleted" notification is received.
	//
	// It indicates that a user has been permanently deleted from SAMS and the
	// handler MUST delete any user-related PII from the system and/or integrated
	// vendor systems to stay in compliance. In the event of an error, the handler
	// MUST make sure the error is surfaced (by either returning or logging the
	// error) to be retried or to a human operator.
	OnUserDeleted func(ctx context.Context, data *UserDeletedData) error
}

type ReceiveSettings = pubsub.ReceiveSettings

var DefaultReceiveSettings = pubsub.DefaultReceiveSettings

const handleReceiveStatusUnknownMessage = "unknown_message"

func (s *subscriber) handleReceive(ctx context.Context, name string, metadata json.RawMessage) (status string, _ error) {
	switch name {
	case nameUserDeleted:
		if s.handlers.OnUserDeleted == nil {
			return "skipped", nil
		}

		var data UserDeletedData
		if err := json.Unmarshal(metadata, &data); err != nil {
			return "malformed_message", errors.Wrap(err, "unmarshal metadata")
		}

		return "handled", s.handlers.OnUserDeleted(ctx, &data)
	}

	// Unknown message type
	return handleReceiveStatusUnknownMessage, nil
}

// state is a concurrent-safe state machine that transitions between "idle",
// "started", and "stopped" states.
type state struct {
	value *atomic.String
}

func newState() state {
	return state{value: atomic.NewString(stateIdle)}
}

func (s state) toStarted() error {
	if !s.value.CompareAndSwap(stateIdle, stateStarted) {
		return errors.Newf("not in %q state", stateIdle)
	}
	return nil
}

func (s state) toStopped() error {
	if !s.value.CompareAndSwap(stateStarted, stateStopped) {
		return errors.Newf("not in %q state", stateStarted)
	}
	return nil
}

const (
	stateIdle    = "idle"
	stateStarted = "started"
	stateStopped = "stopped"
)
