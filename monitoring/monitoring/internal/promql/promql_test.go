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

func TestInject(t *testing.T) {
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
			got, err := Inject(tc.expression, tc.matchers, tc.vars)
			if (err != nil) != tc.wantErr {
				t.Errorf("unexpected result '%+v'", err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
