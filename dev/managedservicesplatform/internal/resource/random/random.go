package random

import (
	"github.com/aws/constructs-go/constructs/v10"
	randomid "github.com/sourcegraph/managed-services-platform-cdktf/gen/random/id"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	ByteLength int `validate:"required"`
	// Prefix is added to the start of the random output followed by a '-', for
	// example:
	//
	//   ${prefix}-${randomSuffix}
	Prefix string
	// Keepers, if changed, rotates the random ID.
	Keepers map[string]*string
}

type Output struct {
	HexValue string
}

// New creates a randomized value.
//
// Requires stack to be created with randomprovider.With().
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	var prefix *string
	if config.Prefix != "" {
		prefix = pointers.Ptr(config.Prefix + "-")
	}
	rid := randomid.NewId(
		scope,
		id.TerraformID("random"),
		&randomid.IdConfig{
			ByteLength: pointers.Float64(config.ByteLength),
			Prefix:     prefix,

			Keepers: func() *map[string]*string {
				if config.Keepers != nil {
					return &config.Keepers
				}
				return nil
			}(),
		},
	)
	return &Output{HexValue: *rid.Hex()}
}
