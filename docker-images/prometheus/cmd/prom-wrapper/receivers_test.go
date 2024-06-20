package main

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2" // same as used in alertmanager

	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAlertSolutionsURL(t *testing.T) {
	defaultURL := fmt.Sprintf("%s/%s", docsURL, alertsDocsPathPath)
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
			if got := alertsReferenceURL(); !strings.Contains(got, tt.wantIncludes) {
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
		name                string
		args                args
		wantProblems        []string // partial message matches
		wantReceiversConfig autogold.Value
		wantReceivers       int  // = 3 without additional receivers
		wantRoutes          int  // = 2 without additional routes
		wantRenderFail      bool // if rendered config is accepted by Alertmanager
	}{
		{
			name: "invalid notifier",
			args: args{
				newAlerts: []*schema.ObservabilityAlerts{{
					Level:    "warning",
					Notifier: schema.Notifier{},
				}},
			},
			wantProblems: []string{"no configuration found"},
			wantReceiversConfig: autogold.Expect(`- name: src-noop-receiver
- name: src-warning-receiver
- name: src-critical-receiver
`),
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
			wantReceiversConfig: autogold.Expect(`- name: src-noop-receiver
- name: src-warning-receiver
  slack_configs:
  - send_resolved: true
    api_url: ""
    username: Sourcegraph Alerts
    color: '{{ if eq .Status "firing" }}#FFFF00{{ else }}#00FF00{{ end }}'
    title: '{{ if eq .Status "firing" }}[{{ .CommonLabels.level | toUpper }}] {{ .CommonLabels.description
      }}{{ else }}[RESOLVED] {{ .CommonLabels.description }}{{ end }}'
    title_link: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
      }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    text: '{{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert ''{{
      .CommonLabels.name }}'' is firing for service ''{{ .CommonLabels.service_name
      }}'' ({{ .CommonLabels.owner }}).{{ else }}{{ .CommonLabels.level | title }}
      alert ''{{ .CommonLabels.name }}'' for service ''{{ .CommonLabels.service_name
      }}'' has resolved.{{ end }}'
    short_fields: false
    link_names: false
    actions:
    - type: button
      text: Next steps
      url: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
        }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    - type: button
      text: Dashboard
      url: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name
        }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id
        }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts
        0).EndsAt.Unix }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{
        else }}&time={{ $start }}000&time.window=3600000{{ end }}
- name: src-critical-receiver
`),
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
						Email: &schema.NotifierEmail{
							Type:    "email",
							Address: "robert@sourcegraph.com",
						},
					},
				}},
			},
			wantReceiversConfig: autogold.Expect(`- name: src-noop-receiver
- name: src-warning-receiver
  slack_configs:
  - send_resolved: true
    api_url: https://sourcegraph.com
    username: Sourcegraph Alerts
    color: '{{ if eq .Status "firing" }}#FFFF00{{ else }}#00FF00{{ end }}'
    title: '{{ if eq .Status "firing" }}[{{ .CommonLabels.level | toUpper }}] {{ .CommonLabels.description
      }}{{ else }}[RESOLVED] {{ .CommonLabels.description }}{{ end }}'
    title_link: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
      }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    text: '{{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert ''{{
      .CommonLabels.name }}'' is firing for service ''{{ .CommonLabels.service_name
      }}'' ({{ .CommonLabels.owner }}).{{ else }}{{ .CommonLabels.level | title }}
      alert ''{{ .CommonLabels.name }}'' for service ''{{ .CommonLabels.service_name
      }}'' has resolved.{{ end }}'
    short_fields: false
    link_names: false
    actions:
    - type: button
      text: Next steps
      url: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
        }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    - type: button
      text: Dashboard
      url: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name
        }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id
        }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts
        0).EndsAt.Unix }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{
        else }}&time={{ $start }}000&time.window=3600000{{ end }}
- name: src-critical-receiver
  email_configs:
  - send_resolved: true
    to: robert@sourcegraph.com
    headers:
      subject: '{{ if eq .Status "firing" }}[{{ .CommonLabels.level | toUpper }}]
        {{ .CommonLabels.description }}{{ else }}[RESOLVED] {{ .CommonLabels.description
        }}{{ end }}'
    html: |-
      <body>{{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert '{{ .CommonLabels.name }}' is firing for service '{{ .CommonLabels.service_name }}' ({{ .CommonLabels.owner }}).

      For next steps, please refer to our documentation: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
      For more details, please refer to the service dashboard: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts 0).EndsAt.Unix }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{ else }}&time={{ $start }}000&time.window=3600000{{ end }}{{ else }}{{ .CommonLabels.level | title }} alert '{{ .CommonLabels.name }}' for service '{{ .CommonLabels.service_name }}' has resolved.{{ end }}</body>
    text: |-
      {{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert '{{ .CommonLabels.name }}' is firing for service '{{ .CommonLabels.service_name }}' ({{ .CommonLabels.owner }}).

      For next steps, please refer to our documentation: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
      For more details, please refer to the service dashboard: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts 0).EndsAt.Unix }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{ else }}&time={{ $start }}000&time.window=3600000{{ end }}{{ else }}{{ .CommonLabels.level | title }} alert '{{ .CommonLabels.name }}' for service '{{ .CommonLabels.service_name }}' has resolved.{{ end }}
`),
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
			wantReceiversConfig: autogold.Expect(`- name: src-noop-receiver
- name: src-warning-on-distribution
  slack_configs:
  - send_resolved: true
    api_url: https://sourcegraph.com
    username: Sourcegraph Alerts
    color: '{{ if eq .Status "firing" }}#FFFF00{{ else }}#00FF00{{ end }}'
    title: '{{ if eq .Status "firing" }}[{{ .CommonLabels.level | toUpper }}] {{ .CommonLabels.description
      }}{{ else }}[RESOLVED] {{ .CommonLabels.description }}{{ end }}'
    title_link: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
      }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    text: '{{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert ''{{
      .CommonLabels.name }}'' is firing for service ''{{ .CommonLabels.service_name
      }}'' ({{ .CommonLabels.owner }}).{{ else }}{{ .CommonLabels.level | title }}
      alert ''{{ .CommonLabels.name }}'' for service ''{{ .CommonLabels.service_name
      }}'' has resolved.{{ end }}'
    short_fields: false
    link_names: false
    actions:
    - type: button
      text: Next steps
      url: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
        }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    - type: button
      text: Dashboard
      url: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name
        }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id
        }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts
        0).EndsAt.Unix }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{
        else }}&time={{ $start }}000&time.window=3600000{{ end }}
- name: src-warning-receiver
- name: src-critical-receiver
`),
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
			wantReceiversConfig: autogold.Expect(`- name: src-noop-receiver
- name: src-warning-on-distribution
  slack_configs:
  - send_resolved: true
    api_url: https://sourcegraph.com
    username: Sourcegraph Alerts
    color: '{{ if eq .Status "firing" }}#FFFF00{{ else }}#00FF00{{ end }}'
    title: '{{ if eq .Status "firing" }}[{{ .CommonLabels.level | toUpper }}] {{ .CommonLabels.description
      }}{{ else }}[RESOLVED] {{ .CommonLabels.description }}{{ end }}'
    title_link: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
      }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    text: '{{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert ''{{
      .CommonLabels.name }}'' is firing for service ''{{ .CommonLabels.service_name
      }}'' ({{ .CommonLabels.owner }}).{{ else }}{{ .CommonLabels.level | title }}
      alert ''{{ .CommonLabels.name }}'' for service ''{{ .CommonLabels.service_name
      }}'' has resolved.{{ end }}'
    short_fields: false
    link_names: false
    actions:
    - type: button
      text: Next steps
      url: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
        }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    - type: button
      text: Dashboard
      url: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name
        }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id
        }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts
        0).EndsAt.Unix }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{
        else }}&time={{ $start }}000&time.window=3600000{{ end }}
  opsgenie_configs:
  - send_resolved: true
    api_key: hi-im-bob
    api_url: https://ubclaunchpad.com
    message: '{{ if eq .Status "firing" }}[{{ .CommonLabels.level | toUpper }}] {{
      .CommonLabels.description }}{{ else }}[RESOLVED] {{ .CommonLabels.description
      }}{{ end }}'
    description: '{{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert
      ''{{ .CommonLabels.name }}'' is firing for service ''{{ .CommonLabels.service_name
      }}'' ({{ .CommonLabels.owner }}).{{ else }}{{ .CommonLabels.level | title }}
      alert ''{{ .CommonLabels.name }}'' for service ''{{ .CommonLabels.service_name
      }}'' has resolved.{{ end }}'
    source: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name
      }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id
      }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts 0).EndsAt.Unix
      }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{ else }}&time={{
      $start }}000&time.window=3600000{{ end }}
    details:
      Next steps: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
        }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    tags: '{{ range $key, $value := .CommonLabels }}{{$key}}={{$value}},{{end}}'
    priority: P2
- name: src-warning-receiver
- name: src-critical-receiver
`),
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
			wantReceiversConfig: autogold.Expect(`- name: src-noop-receiver
- name: src-warning-on-distribution
  opsgenie_configs:
  - send_resolved: true
    message: '{{ if eq .Status "firing" }}[{{ .CommonLabels.level | toUpper }}] {{
      .CommonLabels.description }}{{ else }}[RESOLVED] {{ .CommonLabels.description
      }}{{ end }}'
    description: '{{ if eq .Status "firing" }}{{ .CommonLabels.level | title }} alert
      ''{{ .CommonLabels.name }}'' is firing for service ''{{ .CommonLabels.service_name
      }}'' ({{ .CommonLabels.owner }}).{{ else }}{{ .CommonLabels.level | title }}
      alert ''{{ .CommonLabels.name }}'' for service ''{{ .CommonLabels.service_name
      }}'' has resolved.{{ end }}'
    source: https://sourcegraph.com/-/debug/grafana/d/{{ .CommonLabels.service_name
      }}/{{ .CommonLabels.service_name }}?viewPanel={{ .CommonLabels.grafana_panel_id
      }}{{ $start := (index .Alerts 0).StartsAt.Unix }}{{ $end := (index .Alerts 0).EndsAt.Unix
      }}{{ if gt $end 0 }}&from={{ $start }}000&end={{ $end }}000{{ else }}&time={{
      $start }}000&time.window=3600000{{ end }}
    details:
      Next steps: https://sourcegraph.com/docs/admin/observability/alerts#{{ .CommonLabels.service_name
        }}-{{ .CommonLabels.name | reReplaceAll "_" "-" }}
    tags: '{{ range $key, $value := .CommonLabels }}{{$key}}={{$value}},{{end}}'
    priority: P2
- name: src-warning-receiver
- name: src-critical-receiver
`),
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

			receiversData, err := yaml.Marshal(receivers)
			require.NoError(t, err)
			tt.wantReceiversConfig.Equal(t, string(receiversData))
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
			finalConfig, err := renderConfiguration(&amconfig.Config{
				Global: &amconfig.GlobalConfig{
					// Some global SMTP config is required to test email receivers
					SMTPFrom:  "foo@foo.com",
					SMTPHello: "foo.com",
					SMTPSmarthost: amconfig.HostPort{
						Host: "0.0.0.0",
						Port: "1234",
					},
				},
				Receivers: receivers,
				Route:     newRootRoute(routes),
			})
			t.Log(string(finalConfig))
			if err != nil && !tt.wantRenderFail {
				t.Errorf("generated config is invalid: %s", err)
			} else if err == nil && tt.wantRenderFail {
				t.Error("expected load to fail, but succeeded")
			}
		})
	}
}
