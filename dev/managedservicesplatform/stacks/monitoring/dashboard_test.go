package monitoring

import (
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy"
)

type mockAlertPolicy struct {
	monitoringalertpolicy.MonitoringAlertPolicy
	name string
}

func (m *mockAlertPolicy) Name() *string {
	return &m.name
}

func TestDashboardCreation(t *testing.T) {
	tests := []struct {
		name        string
		serviceID   string
		envID       string
		alertGroups map[string][]monitoringalertpolicy.MonitoringAlertPolicy
		want        autogold.Value
	}{
		{
			name:      "single alert",
			serviceID: "msp-testbed",
			envID:     "dev",
			alertGroups: map[string][]monitoringalertpolicy.MonitoringAlertPolicy{
				"Container Alerts": {
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00000"},
				},
			},
			want: autogold.Expect(dashboard{
				DisplayName: "MSP Alerts - msp-testbed-dev",
				MosaicLayout: mosaicLayout{
					Columns: 48,
					Tiles: []tile{
						{
							Width:  48,
							Height: 8,
							Widget: widget{Text: &text{
								Content: `Auto-generated - Please do not edit

For more details see: [go/msp-ops/msp-testbed](https://handbook.sourcegraph.com/departments/engineering/managed-services/msp-testbed/)`,
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
							}},
						},
						{
							YPos:   8,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00000"}},
						},
						{
							YPos:   8,
							Width:  48,
							Height: 16,
							Widget: widget{
								Title:            "Container Alerts",
								CollapsibleGroup: &collapsibleGroup{},
							},
						},
					},
				},
			}),
		},
		{
			name:      "two alerts",
			serviceID: "msp-testbed",
			envID:     "dev",
			alertGroups: map[string][]monitoringalertpolicy.MonitoringAlertPolicy{
				"Container Alerts": {
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00000"},
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00001"},
				},
			},
			want: autogold.Expect(dashboard{
				DisplayName: "MSP Alerts - msp-testbed-dev",
				MosaicLayout: mosaicLayout{
					Columns: 48,
					Tiles: []tile{
						{
							Width:  48,
							Height: 8,
							Widget: widget{Text: &text{
								Content: `Auto-generated - Please do not edit

For more details see: [go/msp-ops/msp-testbed](https://handbook.sourcegraph.com/departments/engineering/managed-services/msp-testbed/)`,
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
							}},
						},
						{
							YPos:   8,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00000"}},
						},
						{
							YPos:   8,
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00001"}},
						},
						{
							YPos:   8,
							Width:  48,
							Height: 16,
							Widget: widget{
								Title:            "Container Alerts",
								CollapsibleGroup: &collapsibleGroup{},
							},
						},
					},
				},
			}),
		},
		{
			name:      "three alerts",
			serviceID: "msp-testbed",
			envID:     "dev",
			alertGroups: map[string][]monitoringalertpolicy.MonitoringAlertPolicy{
				"Container Alerts": {
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00000"},
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00001"},
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00002"},
				},
			},
			want: autogold.Expect(dashboard{
				DisplayName: "MSP Alerts - msp-testbed-dev",
				MosaicLayout: mosaicLayout{
					Columns: 48,
					Tiles: []tile{
						{
							Width:  48,
							Height: 8,
							Widget: widget{Text: &text{
								Content: `Auto-generated - Please do not edit

For more details see: [go/msp-ops/msp-testbed](https://handbook.sourcegraph.com/departments/engineering/managed-services/msp-testbed/)`,
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
							}},
						},
						{
							YPos:   8,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00000"}},
						},
						{
							YPos:   8,
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00001"}},
						},
						{
							YPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00002"}},
						},
						{
							YPos:   8,
							Width:  48,
							Height: 32,
							Widget: widget{
								Title:            "Container Alerts",
								CollapsibleGroup: &collapsibleGroup{},
							},
						},
					},
				},
			}),
		},
		{
			name:      "multiple sections",
			serviceID: "msp-testbed",
			envID:     "dev",
			alertGroups: map[string][]monitoringalertpolicy.MonitoringAlertPolicy{
				"Container Alerts": {
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00000"},
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00001"},
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00002"},
				},
				"Cloud SQL Alerts": {
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00010"},
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00020"},
					&mockAlertPolicy{name: "/projects/msp-testbed/alertPolicies/00030"},
				},
			},
			want: autogold.Expect(dashboard{
				DisplayName: "MSP Alerts - msp-testbed-dev",
				MosaicLayout: mosaicLayout{
					Columns: 48,
					Tiles: []tile{
						{
							Width:  48,
							Height: 8,
							Widget: widget{Text: &text{
								Content: `Auto-generated - Please do not edit

For more details see: [go/msp-ops/msp-testbed](https://handbook.sourcegraph.com/departments/engineering/managed-services/msp-testbed/)`,
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
							}},
						},
						{
							YPos:   8,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00010"}},
						},
						{
							YPos:   8,
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00020"}},
						},
						{
							YPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00030"}},
						},
						{
							YPos:   8,
							Width:  48,
							Height: 32,
							Widget: widget{
								Title:            "Cloud SQL Alerts",
								CollapsibleGroup: &collapsibleGroup{},
							},
						},
						{
							YPos:   40,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00000"}},
						},
						{
							YPos:   40,
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00001"}},
						},
						{
							YPos:   56,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00002"}},
						},
						{
							YPos:   40,
							Width:  48,
							Height: 32,
							Widget: widget{
								Title:            "Container Alerts",
								CollapsibleGroup: &collapsibleGroup{},
							},
						},
					},
				},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dashboard := generateDashboard(tt.serviceID, tt.envID, tt.alertGroups)
			tt.want.Equal(t, dashboard)
		})
	}
}
