package lifecycle

import "context"

type event interface {
	// This doesn't use json.Marshaler because we want to be able to pass in a
	// context.
	MarshalPayload(ctx context.Context) ([]byte, error)

	// TODO: add methods to access the things we would want to filter webhooks
	// by: organisation, maybe repo, maybe user.
}
