package requestinteraction

import (
	"context"

	"github.com/sourcegraph/log"
)

type requestInteractionKey struct{}

// Interaction carries information about the interaction associated with a
// request - a sort of manually instrumented trace.
type Interaction struct {
	// ID identifies the interaction
	ID string
}

func FromContext(ctx context.Context) *Interaction {
	ip, ok := ctx.Value(requestInteractionKey{}).(*Interaction)
	if !ok || ip == nil {
		return nil
	}
	return ip
}

// WithClient adds client IP information to context for propagation.
func WithClient(ctx context.Context, client *Interaction) context.Context {
	return context.WithValue(ctx, requestInteractionKey{}, client)
}

func (c *Interaction) LogFields() []log.Field {
	if c == nil {
		return []log.Field{log.String("requestInteraction", "<nil>")}
	}
	return []log.Field{
		log.String("requestInteraction.id", c.ID),
	}
}
