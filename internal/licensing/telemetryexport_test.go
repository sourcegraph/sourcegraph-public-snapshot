package licensing

import (
	"testing"
	"time"

	"github.com/hexops/autogold/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/license/licensetest"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestTelemetryEventsExportEnablementCutOffDate(t *testing.T) {
	autogold.Expect("2023-10-04 00:00:00 +0000 UTC").Equal(t, telemetryEventsExportEnablementCutOffDate.String())
}

func TestGetTelemetryEventsExportMode(t *testing.T) {
	t.Run("global state defaults to nil", func(t *testing.T) {
		assert.Nil(t, telemetryEventsExportMode.Load())
	})

	t.Cleanup(func() { telemetryEventsExportMode.Store(nil) })

	c := conf.MockClient()
	c.Mock(&conf.Unified{
		SiteConfiguration: schema.SiteConfiguration{
			LicenseKey: "", // no license key
		},
	})

	t.Run("sets default mode enabled", func(t *testing.T) {
		mode := GetTelemetryEventsExportMode(c)
		assert.Equal(t, TelemetryEventsExportAll, mode)
		// cache is set
		cached := telemetryEventsExportMode.Load()
		assert.NotNil(t, cached)
		assert.Equal(t, mode, cached.mode)
	})
}

func TestNewTelemetryEventsExportMode(t *testing.T) {
	var mustKey = func(info license.Info) string {
		key, _, err := license.GenerateSignedKey(info, licensetest.PrivateKey)
		require.NoError(t, err)
		return key
	}

	for _, tc := range []struct {
		name       string
		licenseKey string
		wantMode   TelemetryEventsExportMode
	}{
		{
			name:       "no key",
			licenseKey: "",
			wantMode:   TelemetryEventsExportAll,
		},
		{
			name:       "invalid key",
			licenseKey: "robert",
			wantMode:   TelemetryEventsExportAll,
		},
		{
			name: "export disabled (cody-only export)",
			licenseKey: mustKey(license.Info{
				Tags: []string{TelemetryEventsExportDisabledTag},
			}),
			wantMode: TelemetryEventsExportCodyOnly, // cody export is allowed due to terms of use
		},
		{
			name: "export disabled via airgapped plan (cody-only export)",
			licenseKey: mustKey(license.Info{
				Tags: []string{PlanAirGappedEnterprise.tag()},
			}),
			wantMode: TelemetryEventsExportCodyOnly, // cody export is allowed due to terms of use
		},
		{
			name:       "no tags, unknown creation",
			licenseKey: mustKey(license.Info{}),
			wantMode:   TelemetryEventsExportCodyOnly,
		},
		{
			name: "before cutoff",
			licenseKey: mustKey(license.Info{
				CreatedAt: telemetryEventsExportEnablementCutOffDate.
					Add(-1 * time.Hour),
			}),
			wantMode: TelemetryEventsExportCodyOnly,
		},
		{
			name: "after cutoff",
			licenseKey: mustKey(license.Info{
				CreatedAt: telemetryEventsExportEnablementCutOffDate.
					Add(1 * time.Hour),
			}),
			wantMode: TelemetryEventsExportAll,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mode := newTelemetryEventsExportMode(tc.licenseKey, licensetest.PublicKey)
			assert.Equal(t, tc.wantMode, mode)
		})
	}

	t.Run("forceExportNone", func(t *testing.T) {
		forceExportNone = true
		t.Cleanup(func() { forceExportNone = false })

		t.Run("disabled with airgapped tag", func(t *testing.T) {
			mode := newTelemetryEventsExportMode(mustKey(license.Info{
				Tags: []string{PlanAirGappedEnterprise.tag()},
			}), licensetest.PublicKey)
			assert.Equal(t, TelemetryEventsExportDisabled, mode)
		})

		t.Run("cody-only with disabled", func(t *testing.T) {
			mode := newTelemetryEventsExportMode(mustKey(license.Info{
				Tags: []string{TelemetryEventsExportDisabledTag},
			}), licensetest.PublicKey)
			assert.Equal(t, TelemetryEventsExportCodyOnly, mode)
		})

		t.Run("no-op without tags", func(t *testing.T) {
			mode := newTelemetryEventsExportMode(mustKey(license.Info{
				CreatedAt: telemetryEventsExportEnablementCutOffDate.
					Add(1 * time.Hour),
			}), licensetest.PublicKey)
			assert.Equal(t, TelemetryEventsExportAll, mode)
		})
	})
}
