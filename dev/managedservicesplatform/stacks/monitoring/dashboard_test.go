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
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{
								Name: "/projects/msp-testbed/alertPolicies/00000",
							}},
						},
						{
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
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{
								Name: "/projects/msp-testbed/alertPolicies/00000",
							}},
						},
						{
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00001"}},
						},
						{
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
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{
								Name: "/projects/msp-testbed/alertPolicies/00000",
							}},
						},
						{
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00001"}},
						},
						{
							YPos:   16,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00002"}},
						},
						{
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
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{
								Name: "/projects/msp-testbed/alertPolicies/00000",
							}},
						},
						{
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00001"}},
						},
						{
							YPos:   16,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00002"}},
						},
						{
							Width:  48,
							Height: 32,
							Widget: widget{
								Title:            "Container Alerts",
								CollapsibleGroup: &collapsibleGroup{},
							},
						},
						{
							YPos:   32,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00010"}},
						},
						{
							YPos:   32,
							XPos:   24,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00020"}},
						},
						{
							YPos:   48,
							Width:  24,
							Height: 16,
							Widget: widget{AlertChart: &alertChart{Name: "/projects/msp-testbed/alertPolicies/00030"}},
						},
						{
							YPos:   32,
							Width:  48,
							Height: 32,
							Widget: widget{
								Title:            "Cloud SQL Alerts",
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
