package random

import (
	"github.com/aws/constructs-go/constructs/v10"
	randomid "github.com/sourcegraph/managed-services-platform-cdktf/gen/random/id"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	ByteLength int `validate:"required"`
}

type Output struct {
	HexValue string
}

// New creates a randomized value.
//
// Requires stack to be created with randomprovider.With().
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	rid := randomid.NewId(
		scope,
		id.ResourceID("random"),
		&randomid.IdConfig{
			ByteLength: pointers.Float64(config.ByteLength),
		},
	)
	return &Output{HexValue: *rid.Hex()}
}
