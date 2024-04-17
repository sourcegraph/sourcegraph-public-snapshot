package monitoring

import (
	"encoding/json"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringdashboard"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"golang.org/x/exp/maps"
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
	dashboard := generateDashboard(vars.Service.ID, vars.EnvironmentID, alertGroups)

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

func generateDashboard(serviceID, envID string, alertGroups map[string][]monitoringalertpolicy.MonitoringAlertPolicy) dashboard {
	// Add a banner informing operators not to edit the dashboard
	infoBanner := tile{
		Width:  dashboardColumns,
		Height: 8,
		Widget: widget{
			Text: &text{
				Content: fmt.Sprintf("Auto-generated - Please do not edit\n\nFor more details see: [go/msp-ops/%[1]s](https://handbook.sourcegraph.com/departments/engineering/managed-services/%[1]s/)", serviceID),
				Format:  "MARKDOWN",
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

	// for consistency we sort the map keys first
	keys := maps.Keys(alertGroups)
	slices.Sort(keys)
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
		DisplayName: fmt.Sprintf("MSP Alerts - %s-%s", serviceID, envID),
		MosaicLayout: mosaicLayout{
			Columns: dashboardColumns,
			Tiles:   tiles,
		},
	}
}
