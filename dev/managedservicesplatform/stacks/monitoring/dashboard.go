package monitoring

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringdashboard"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

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
}

type alertChart struct {
	Name string `json:"name"`
}

type collapsibleGroup struct {
	Collapsed bool `json:"collapsed"`
}

const dashboardColumns = 48 // GCP default
const widgetWidth = 24
const widgetHeight = 16

func createMonitoringDashboard(stack cdktf.TerraformStack,
	id resourceid.ID,
	vars Variables,
	alertGroups map[string][]monitoringalertpolicy.MonitoringAlertPolicy,
) error {
	dashboard := dashboard{
		DisplayName: fmt.Sprintf("MSP Alerts - %s-%s", vars.Service.ID, vars.EnvironmentID),
		MosaicLayout: mosaicLayout{
			Columns: dashboardColumns,
		},
	}

	tiles := []tile{}

	// absolute distance from top of dashboard
	// multiples of 16
	height := 0
	for section, alerts := range alertGroups {
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
		sectionHeight := (len(alerts) + 1) * widgetHeight / 2

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

	dashboard.MosaicLayout.Tiles = tiles

	dashboardJSON, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}

	_ = monitoringdashboard.NewMonitoringDashboard(stack, id.TerraformID("dashboard"), &monitoringdashboard.MonitoringDashboardConfig{
		DashboardJson: pointers.Ptr(string(dashboardJSON)),
	})

	return nil
}
