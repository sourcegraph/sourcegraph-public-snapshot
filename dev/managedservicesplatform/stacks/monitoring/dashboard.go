package monitoring

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringdashboard"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// Schema reference: https://cloud.google.com/monitoring/api/ref_v3/rest/v1/projects.dashboards#resource:-dashboard
type dashboard struct {
	DisplayName  string       `json:"displayName"`
	MosaicLayout mosaicLayout `json:"mosaicLayout"`
}

type mosaicLayout struct {
	Columns int    `json:"columns"`
	Tiles   []tile `json:"tiles"`
}

type tile struct {
	YPos   int    `json:"yPos,omitempty"`
	XPos   int    `json:"xPos,omitempty"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Widget widget `json:"widget"`
}

type widget struct {
	Title            string            `json:"title,omitempty"`
	AlertChart       *alertChart       `json:"alertChart,omitempty"`
	CollapsibleGroup *collapsibleGroup `json:"collapsibleGroup,omitempty"`
	Text             *text             `json:"text,omitempty"`
}

type alertChart struct {
	Name string `json:"name"`
}

type collapsibleGroup struct {
	Collapsed bool `json:"collapsed"`
}

type text struct {
	Content string    `json:"content"`
	Format  string    `json:"format"`
	Style   textStyle `json:"style"`
}

type textStyle struct {
	BackgroundColor     string `json:"backgroundColor"`
	FontSize            string `json:"fontSize"`
	HorizontalAlignment string `json:"horizontalAlignment"`
	Padding             string `json:"padding"`
	PointerLocation     string `json:"pointerLocation"`
	TextColor           string `json:"textColor"`
	VerticalAlignment   string `json:"verticalAlignment"`
}

func createMonitoringDashboard(stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	alertGroups map[string][]monitoringalertpolicy.MonitoringAlertPolicy,
) error {
	dashboard := generateDashboard(vars.Service, vars.EnvironmentID, alertGroups)

	dashboardJSON, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}

	_ = monitoringdashboard.NewMonitoringDashboard(stack, id.TerraformID("dashboard"), &monitoringdashboard.MonitoringDashboardConfig{
		DashboardJson: pointers.Ptr(string(dashboardJSON)),
	})

	return nil
}

// GCP defaults
const dashboardColumns = 48
const widgetWidth = 24
const widgetHeight = 16

// Groups with special treatment
const (
	customAlertsGroupName            = "Custom Alerts"
	responseCodeRatioAlertsGroupName = "Response Code Ratio Alerts"
)

func generateDashboard(svc spec.ServiceSpec, envID string, alertGroups map[string][]monitoringalertpolicy.MonitoringAlertPolicy) dashboard {
	// Add a banner informing operators not to edit the dashboard
	infoBanner := tile{
		Width:  dashboardColumns,
		Height: 8,
		Widget: widget{
			Text: &text{
				Content: fmt.Sprintf("This dashboard is auto-generated - please do not edit!\n\nFor more details, see: %s",
					svc.GetHandbookPageURL()),
				Format: "MARKDOWN",
				Style: textStyle{
					BackgroundColor:     "#FFFFFF",
					FontSize:            "FS_EXTRA_LARGE",
					HorizontalAlignment: "H_CENTER",
					Padding:             "P_EXTRA_SMALL",
					PointerLocation:     "POINTER_LOCATION_UNSPECIFIED",
					TextColor:           "#000000",
					VerticalAlignment:   "V_CENTER",
				},
			},
		},
	}

	tiles := []tile{infoBanner}

	// absolute distance from top of dashboard
	height := infoBanner.Height

	// for consistency we sort the map keys first, but place the custom alerts
	// at the start
	firstKeys := []string{customAlertsGroupName, responseCodeRatioAlertsGroupName}
	keys := remove(maps.Keys(alertGroups), firstKeys)
	slices.Sort(keys)
	keys = append(firstKeys, keys...)
	for _, section := range keys {
		alerts := alertGroups[section]
		// Ensure we don't create empty sections
		if len(alerts) == 0 {
			continue
		}

		sectionOffset := 0
		for i, alert := range alerts {
			// add the alertPolicy widget
			tiles = append(tiles, tile{
				XPos:   (i % 2) * widgetWidth,
				YPos:   height + sectionOffset,
				Width:  widgetWidth,
				Height: widgetHeight,
				Widget: widget{
					AlertChart: &alertChart{
						Name: *alert.Name(),
					},
				},
			})
			// two alerts per row
			if i%2 == 1 {
				sectionOffset += widgetHeight
			}
		}

		// Use integer division to calculate the height of the section
		sectionHeight := (len(alerts) + 1) / 2 * widgetHeight

		// Add the section
		tiles = append(tiles, tile{
			Width:  dashboardColumns,
			YPos:   height,
			Height: sectionHeight,
			Widget: widget{
				Title:            section,
				CollapsibleGroup: &collapsibleGroup{},
			},
		})

		height += sectionHeight
	}

	return dashboard{
		DisplayName: fmt.Sprintf("MSP Alerts - %s (%s)", svc.GetName(), envID),
		MosaicLayout: mosaicLayout{
			Columns: dashboardColumns,
			Tiles:   tiles,
		},
	}
}

func remove(v []string, filter []string) []string {
	for _, f := range filter {
		for i, s := range v {
			if s == f {
				v = append(v[:i], v[i+1:]...)
			}
		}
	}
	return v
}
