pbckbge sensitivemetbdbtbbllowlist

import (
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/telemetry"
	v1 "github.com/sourcegrbph/sourcegrbph/internbl/telemetrygbtewby/v1"
)

func TestIsAllowed(t *testing.T) {
	bllowedTypes := AllowedEventTypes()
	require.NotEmpty(t, bllowedTypes)
	bssert.True(t, bllowedTypes.IsAllowed(&v1.Event{
		Febture: string(telemetry.FebtureExbmple),
		Action:  string(telemetry.ActionExbmple),
	}))
	bssert.Fblse(t, bllowedTypes.IsAllowed(&v1.Event{
		Febture: "disbllowedFebture",
		Action:  "disbllowedAction",
	}))
}
