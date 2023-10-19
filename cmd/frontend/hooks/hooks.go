// Package hooks allow hooking into the frontend.
package hooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

// PostAuthMiddleware is an HTTP handler middleware that, if set, runs just before auth-related
// middleware. The client is authenticated when PostAuthMiddleware is called.
var PostAuthMiddleware func(http.Handler) http.Handler

// FeatureBatchChanges describes if and how the Batch Changes feature is available on
// the given license plan. It mirrors the type licensing.FeatureBatchChanges.
type FeatureBatchChanges struct {
	// If true, there is no limit to the number of changesets that can be created.
	Unrestricted bool `json:"unrestricted"`
	// Maximum number of changesets that can be created per batch change.
	// If Unrestricted is true, this is ignored.
	MaxNumChangesets int `json:"maxNumChangesets"`
}

// LicenseInfo contains non-sensitive information about the legitimate usage of the
// current license on the instance. It is technically accessible to all users, so only
// include information that is safe to be seen by others.
type LicenseInfo struct {
	CurrentPlan string `json:"currentPlan"`

	CodeScaleLimit         string               `json:"codeScaleLimit"`
	CodeScaleCloseToLimit  bool                 `json:"codeScaleCloseToLimit"`
	CodeScaleExceededLimit bool                 `json:"codeScaleExceededLimit"`
	KnownLicenseTags       []string             `json:"knownLicenseTags"`
	BatchChanges           *FeatureBatchChanges `json:"batchChanges"`
}

// Surface basic, non-sensitive information about the license type. This information
// can be used to soft-gate features from the UI, and to provide info to admins from
// site admin settings pages in the UI.
func GetLicenseInfo(ctx context.Context, logger log.Logger, db database.DB) *LicenseInfo {
	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		logger.Error("Failed to get license info", log.Error(err))
		return nil
	}

	licenseInfo := &LicenseInfo{
		CurrentPlan: string(info.Plan()),
	}
	if info.Plan() == licensing.PlanBusiness0 {
		const codeScaleLimit = 100 * 1024 * 1024 * 1024
		licenseInfo.CodeScaleLimit = "100GiB"

		stats, err := usagestats.GetRepositories(ctx, db)
		if err != nil {
			logger.Error("Failed to get repository stats", log.Error(err))
			return nil
		}

		if stats.GitDirBytes >= codeScaleLimit {
			licenseInfo.CodeScaleExceededLimit = true
		} else if stats.GitDirBytes >= codeScaleLimit*0.9 {
			licenseInfo.CodeScaleCloseToLimit = true
		}
	}

	// returning this only makes sense on dotcom
	if envvar.SourcegraphDotComMode() {
		for _, plan := range licensing.AllPlans {
			licenseInfo.KnownLicenseTags = append(licenseInfo.KnownLicenseTags, fmt.Sprintf("plan:%s", plan))
		}
		for _, feature := range licensing.AllFeatures {
			licenseInfo.KnownLicenseTags = append(licenseInfo.KnownLicenseTags, feature.FeatureName())
		}
		licenseInfo.KnownLicenseTags = append(licenseInfo.KnownLicenseTags, licensing.MiscTags...)
	} else { // returning BC info only makes sense on non-dotcom
		bcFeature := &licensing.FeatureBatchChanges{}
		if err := licensing.Check(bcFeature); err == nil {
			if bcFeature.Unrestricted {
				licenseInfo.BatchChanges = &FeatureBatchChanges{
					Unrestricted: true,
					// Superceded by being unrestricted
					MaxNumChangesets: -1,
				}
			} else {
				max := int(bcFeature.MaxNumChangesets)
				licenseInfo.BatchChanges = &FeatureBatchChanges{
					MaxNumChangesets: max,
				}
			}
		}
	}

	return licenseInfo
}
