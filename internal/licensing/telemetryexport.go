package licensing

import (
	"os"
	"slices"
	"time"

	"go.uber.org/atomic"
	"golang.org/x/crypto/ssh"

	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/dotcom"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/license"
)

var forceExportAll = env.MustGetBool("SRC_TELEMETRY_EVENTS_EXPORT_ALL", false, "Set to true to forcibly enable all events export.")

// telemetryExportEnablementCutOffDate is Oct 4, 2023 UTC, and all licenses
// created after this date will have telemetry export enabled by default.
//
// The date is based on the Oct 3, 2023 date requested in
// https://docs.google.com/document/d/1Z1Yp7G61WYlQ1B4vO5-mIXVtmvzGmD7PqYHNBQV-2Ik/edit#bookmark=id.vfdpfi9tgzdk
// with a bit of a grace period.
var telemetryEventsExportEnablementCutOffDate = time.Date(2023, time.October, 4,
	0, 0, 0, 0, time.UTC)

// telemetryEventsExportMode caches the evaluated mode globally.
var telemetryEventsExportMode = atomic.NewPointer((*evaluatedTelemetryEventsExportMode)(nil))

// GetTelemetryEventsExportMode returns the degree of telemetry events export
// enabled. See TelemetryEventsExportMode for more details.
func GetTelemetryEventsExportMode(c conftypes.SiteConfigQuerier) TelemetryEventsExportMode {
	if forceExportAll {
		return TelemetryEventsExportAll
	}

	evaluatedMode := telemetryEventsExportMode.Load()

	// Update if changed license key has changed
	if lc := c.SiteConfig().LicenseKey; evaluatedMode == nil || evaluatedMode.licenseKey != lc {
		evaluatedMode = &evaluatedTelemetryEventsExportMode{
			licenseKey: lc,
			mode:       newTelemetryEventsExportMode(lc, publicKey),
		}
		telemetryEventsExportMode.Store(evaluatedMode)
	}

	return evaluatedMode.mode
}

// TelemetryEventsExportMode is only used to gate whether an event is queued for
// export, which effectively decides whether the event will be exported or not.
//
// See newTelemetryEventsExportMode for the conditions to enable each export mode.
type TelemetryEventsExportMode int

const (
	// TelemetryEventsExportDisabled disables queueing for export of any v2 events.
	TelemetryEventsExportDisabled TelemetryEventsExportMode = iota
	// TelemetryEventsExportAll enables export of all recorded v2 events.
	TelemetryEventsExportAll
	// TelemetryEventsExportCodyOnly exports only Cody-related v2 events, with
	// features 'cody' or 'cody.*'.
	TelemetryEventsExportCodyOnly
)

// legacyExportUsageDataEnabled is the legacy 'EXPORT_USAGE_DATA_ENABLED'
// env var is set. Typically used in Cloud, if the legacy export is enabled,
// we can assume exports are enabled as well.
var legacyExportUsageDataEnabled = os.Getenv("EXPORT_USAGE_DATA_ENABLED") == "true"

func newTelemetryEventsExportMode(licenseKey string, pk ssh.PublicKey) TelemetryEventsExportMode {
	if licenseKey == "" {
		return TelemetryEventsExportAll // without licensing
	}

	if dotcom.SourcegraphDotComMode() {
		return TelemetryEventsExportAll // dotcom mode
	}

	// Use parametrized pk for testing
	key, _, err := license.ParseSignedKey(licenseKey, pk)
	if err != nil || key == nil {
		return TelemetryEventsExportAll // without a valid license key
	}

	if (&Info{Info: *key}).HasTag(FeatureAllowAirGapped.FeatureName()) {
		return TelemetryEventsExportDisabled // this is the only way to disable export entirely
	}

	if slices.Contains(key.Tags, TelemetryEventsExportDisabledTag) {
		return TelemetryEventsExportCodyOnly // license-based opt-out mechanism
	}

	if legacyExportUsageDataEnabled {
		return TelemetryEventsExportAll // match legacy configuration
	}

	if key.CreatedAt.Before(telemetryEventsExportEnablementCutOffDate) {
		return TelemetryEventsExportCodyOnly // pre-cutoff policy
	}

	return TelemetryEventsExportAll
}

type evaluatedTelemetryEventsExportMode struct {
	// licenseKey used to evaluate the mode
	licenseKey string
	// mode is the evaluated mode
	mode TelemetryEventsExportMode
}
