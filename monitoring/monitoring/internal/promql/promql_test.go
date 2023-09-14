package promql

import (
	"testing"

	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/assert"
)

func TestValidate(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		vars       VariableApplier

		wantErr bool
	}{
		{
			name:       "valid expression",
			expression: "foobar",
			wantErr:    false,
		},
		{
			name:       "valid variable expression",
			expression: `foobar{foo="$var"}`, // "$variable" is valid promql
			wantErr:    false,
		},
		{
			name:       "invalid variable expression",
			expression: `foobar[$time]`, // not valid promql
			wantErr:    true,
		},
		{
			name:       "invalid expression fixed by vars",
			expression: `foobar[$time]`, // not valid promql
			vars:       VariableApplier{"time": "1m"},
			wantErr:    false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := Validate(tc.expression, tc.vars)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected result '%+v'", err)
			}
		})
	}
}

func TestInjectMatchers(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		matchers   []*labels.Matcher
		vars       VariableApplier

		want    string
		wantErr bool
	}{
		{
			name:       "valid expression, nothing to inject",
			expression: "foobar",
			matchers:   []*labels.Matcher{},

			want:    "foobar",
			wantErr: false,
		},
		{
			name:       "valid expression",
			expression: "foobar",
			matchers:   []*labels.Matcher{labels.MustNewMatcher(labels.MatchEqual, "key", "value")},

			want:    `foobar{key="value"}`,
			wantErr: false,
		},
		{
			name:       "valid expression with labels",
			expression: `foobar{foo="var"}`,
			matchers:   []*labels.Matcher{labels.MustNewMatcher(labels.MatchEqual, "key", "value")},

			want:    `foobar{foo="var",key="value"}`,
			wantErr: false,
		},
		{
			name:       "invalid expression",
			expression: `foobar[$time]`, // not valid promql
			matchers:   []*labels.Matcher{labels.MustNewMatcher(labels.MatchEqual, "key", "value")},

			want:    "foobar[$time]",
			wantErr: true,
		},
		{
			name:       "invalid expression fixed by vars",
			expression: `avg_over_time(foobar[$time])`, // not valid promql
			matchers:   []*labels.Matcher{labels.MustNewMatcher(labels.MatchEqual, "key", "value")},
			vars:       VariableApplier{"time": "59m"}, // use default sentinel value from getSentinelValue

			want:    `avg_over_time(foobar{key="value"}[$time])`,
			wantErr: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := InjectMatchers(tc.expression, tc.matchers, tc.vars)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected result '%+v'", err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInjectAsAlert(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		matchers   []*labels.Matcher
		vars       VariableApplier

		want    string
		wantErr bool
	}{
		{
			name:       "valid expression, nothing to inject or drop",
			expression: "foobar",
			matchers:   []*labels.Matcher{},

			want:    "foobar",
			wantErr: false,
		},
		{
			name:       "valid expression, nothing to drop",
			expression: "foobar",
			matchers:   []*labels.Matcher{labels.MustNewMatcher(labels.MatchEqual, "key", "value")},

			want:    `foobar{key="value"}`,
			wantErr: false,
		},
		{
			name:       "valid expression, drop variable label",
			expression: `foobar{foo="${var:foo}"}`,
			matchers:   []*labels.Matcher{labels.MustNewMatcher(labels.MatchEqual, "key", "value")},
			vars:       VariableApplier{"var": "asdf"},

			want:    `foobar{key="value"}`,
			wantErr: false,
		},
		{
			name:       "undroppable label",
			expression: `foobar[$time]`, // not valid promql
			want:       "foobar[$time]",
			wantErr:    true,
		},
		{
			name:       "variable used as regexp match",
			expression: `src_executor_processor_handlers{queue=~"${queue:regex}",sg_job=~"^sourcegraph-executors.*"}`,
			vars:       VariableApplier{"queue": "foobar"},
			want:       "src_executor_processor_handlers{sg_job=~\"^sourcegraph-executors.*\"}",
			wantErr:    false,
		},
		{
			name:       "variable used as regexp match without '${...:regex}'",
			expression: `max((max(src_codeintel_commit_graph_queued_duration_seconds_total{job=~"^$source.*"})) >= 3600)`,
			vars:       VariableApplier{"source": "frontend"},
			want:       `max((max(src_codeintel_commit_graph_queued_duration_seconds_total{job=~"^$source.*"})) >= 3600)`,
			wantErr:    true,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := InjectAsAlert(tc.expression, tc.matchers, tc.vars)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected result '%+v'", err)
			} else if err != nil {
				t.Logf("got expected error '%s'", err.Error())
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestInjectGroupings(t *testing.T) {
	for _, tc := range []struct {
		name       string
		expression string
		groupings  []string
		vars       VariableApplier

		want    string
		wantErr bool
	}{
		{
			name:       "repeated and without existing by()",
			expression: `max((max(src_codeintel_commit_graph_queued_duration_seconds_total)) >= 3600)`,
			groupings:  []string{"project_id"},
			want:       `max by (project_id) ((max by (project_id) (src_codeintel_commit_graph_queued_duration_seconds_total)) >= 3600)`,
			wantErr:    false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := InjectGroupings(tc.expression, tc.groupings, tc.vars)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected result '%+v'", err)
			} else if err != nil {
				t.Logf("got expected error '%s'", err.Error())
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestVarKeyRegexp(t *testing.T) {
	re, err := newVarKeyRegexp("queue")
	assert.NoError(t, err)
	assert.True(t, re.MatchString(`src_executor_processor_handlers{queue=~"${queue:regex}",sg_job=~"^sourcegraph-executors.*"}`))
}
