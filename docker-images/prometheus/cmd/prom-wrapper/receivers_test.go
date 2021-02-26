package main

import (
	"fmt"
	"strings"
	"testing"

	amconfig "github.com/prometheus/alertmanager/config"

	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAlertSolutionsURL(t *testing.T) {
	defaultURL := fmt.Sprintf("%s/%s", docsURL, alertSolutionsPagePath)
	tests := []struct {
		name         string
		mockVersion  string
		wantIncludes string
	}{
		{
			name:         "no version set",
			mockVersion:  "",
			wantIncludes: defaultURL,
		}, {
			name:         "dev version set",
			mockVersion:  "0.0.0+dev",
			wantIncludes: defaultURL,
		}, {
			name:         "not a semver",
			mockVersion:  "85633_2021-01-28_f6a6fef",
			wantIncludes: defaultURL,
		}, {
			name:         "semver",
			mockVersion:  "3.24.1",
			wantIncludes: "@v3.24.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version.Mock(tt.mockVersion)
			if got := alertSolutionsURL(); !strings.Contains(got, tt.wantIncludes) {
				t.Errorf("alertSolutionsURL() = %q, should include %q", got, tt.wantIncludes)
			}
		})
	}
}

func TestNewRoutesAndReceivers(t *testing.T) {
	type args struct {
		newAlerts []*schema.ObservabilityAlerts
	}
	tests := []struct {
		name           string
		args           args
		wantProblems   []string // partial message matches
		wantReceivers  int      // = 3 without additional receivers
		wantRoutes     int      // = 2 without additional routes
		wantRenderFail bool     // if rendered config is accepted by Alertmanager
	}{
		{
			name: "invalid notifier",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{{
					Level:    "warning",
					Notifier: schema.Notifier{},
				}},
			},
			wantProblems:  []string{"no configuration found"},
			wantReceivers: 3,
			wantRoutes:    2,
		},
		{
			name: "invalid generated configuration",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{{
					Level: "warning",
					Notifier: schema.Notifier{
						// Alertmanager requires a URL here, so this will fail
						Slack: &schema.NotifierSlack{
							Type: "email",
							Url:  "",
						},
					},
				}},
			},
			wantReceivers:  3,
			wantRoutes:     2,
			wantRenderFail: true,
		},
		{
			name: "one warning one critical",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{{
					Level: "warning",
					Notifier: schema.Notifier{
						Slack: &schema.NotifierSlack{
							Type: "slack",
							Url:  "https://sourcegraph.com",
						},
					},
				}, {
					Level: "critical",
					Notifier: schema.Notifier{
						Slack: &schema.NotifierSlack{
							Type: "slack",
							Url:  "https://sourcegraph.com",
						},
					},
				}},
			},
			wantReceivers: 3,
			wantRoutes:    2,
		}, {
			name: "one custom route",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{{
					Level: "warning",
					Notifier: schema.Notifier{
						Slack: &schema.NotifierSlack{
							Type: "slack",
							Url:  "https://sourcegraph.com",
						},
					},
					Owners: []string{"distribution"},
				}},
			},
			wantReceivers: 4,
			wantRoutes:    3,
		}, {
			name: "multiple alerts on same owner-level combination",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{{
					Level: "warning",
					Notifier: schema.Notifier{
						Slack: &schema.NotifierSlack{
							Type: "slack",
							Url:  "https://sourcegraph.com",
						},
					},
					Owners: []string{"distribution"},
				}, {
					Level: "warning",
					Notifier: schema.Notifier{
						Opsgenie: &schema.NotifierOpsGenie{
							Type:   "opsgenie",
							ApiUrl: "https://ubclaunchpad.com",
							ApiKey: "hi-im-bob",
						},
					},
					Owners: []string{"distribution"},
				}},
			},
			wantReceivers: 4,
			wantRoutes:    3,
		},
		{
			name: "missing env var for opsgenie",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{{
					Level: "warning",
					Notifier: schema.Notifier{
						Opsgenie: &schema.NotifierOpsGenie{
							Type: "opsgenie",
						},
					},
					Owners: []string{"distribution"},
				}},
			},

			wantReceivers:  4,
			wantRoutes:     3,
			wantRenderFail: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			problems := []string{}
			receivers, routes := newRoutesAndReceivers(tt.args.newAlerts, "https://sourcegraph.com", func(err error) {
				problems = append(problems, err.Error())
			})
			if len(tt.wantProblems) != len(problems) {
				t.Errorf("expected problems %+v, got %+v", tt.wantProblems, problems)
				return
			}
			for i, p := range problems {
				if !strings.Contains(p, tt.wantProblems[i]) {
					t.Errorf("expected problem %v to contain %q, got %q", i, tt.wantProblems[i], p)
					return
				}
			}
			if len(receivers) != tt.wantReceivers {
				t.Errorf("expected %d receivers, got %d", tt.wantReceivers, len(receivers))
				return
			}
			if len(routes) != tt.wantRoutes {
				t.Errorf("expected %d routes, got %d", tt.wantRoutes, len(routes))
				return
			}

			// check each route has valid receiver
			receiverNames := map[string]struct{}{}
			for _, rc := range receivers {
				receiverNames[rc.Name] = struct{}{}
			}
			for i, rt := range routes {
				if _, receiverExists := receiverNames[rt.Receiver]; !receiverExists {
					t.Errorf("route %d uses receiver %q, but receiver does not exist", i, rt.Receiver)
				}
			}

			// ensure configuration is valid
			data, err := renderConfiguration(&amconfig.Config{
				Receivers: receivers,
				Route:     newRootRoute(routes),
			})
			t.Log(string(data))
			if err != nil && !tt.wantRenderFail {
				t.Errorf("generated config is invalid: %s", err)
			} else if err == nil && tt.wantRenderFail {
				t.Error("expected load to fail, but succeeded")
			}
		})
	}
}
