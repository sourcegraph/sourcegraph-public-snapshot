package graphqlbackend

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/sourcegraph/sourcegraph/internal/prometheusutil"
)

func Test_siteMonitoringStatisticsResolver_Alerts(t *testing.T) {
	mock := prometheusutil.NewMockPrometheusQuerier()
	sampleT := model.Time(time.Now().UTC().Unix())
	type fields struct {
		queryValue    model.Value
		queryWarnings prometheus.Warnings
		queryErr      error
	}
	tests := []struct {
		name    string
		fields  fields
		want    []*MonitoringAlert
		wantErr error
	}{
		{
			name: "includes alerts with no occurrences",
			fields: fields{
				queryValue: model.Matrix{
					&model.SampleStream{
						Metric: model.Metric{"name": "hello", "service_name": "world", "level": "warn"},
						Values: []model.SamplePair{{Timestamp: sampleT, Value: model.SampleValue(0)}}},
				},
			},
			want: []*MonitoringAlert{{
				TimestampValue:   DateTime{sampleT.Time().Truncate(time.Hour)},
				NameValue:        "warn: hello",
				ServiceNameValue: "world",
				AverageValue:     0,
			}},
			wantErr: nil,
		}, {
			name: "includes alerts with occurrences",
			fields: fields{
				queryValue: model.Matrix{
					&model.SampleStream{
						Metric: model.Metric{"name": "hello", "service_name": "world", "level": "warn"},
						Values: []model.SamplePair{{Timestamp: sampleT, Value: model.SampleValue(1)}}},
				},
			},
			want: []*MonitoringAlert{{
				TimestampValue:   DateTime{sampleT.Time().Truncate(time.Hour)},
				NameValue:        "warn: hello",
				ServiceNameValue: "world",
				AverageValue:     1,
			}},
			wantErr: nil,
		}, {
			name: "discards repeated values",
			fields: fields{
				queryValue: model.Matrix{
					&model.SampleStream{
						Metric: model.Metric{"name": "hello", "service_name": "world", "level": "warn"},
						Values: []model.SamplePair{
							{Timestamp: sampleT, Value: model.SampleValue(1)},
							{Timestamp: sampleT.Add(time.Hour), Value: model.SampleValue(1)},
							{Timestamp: sampleT.Add(2 * time.Hour), Value: model.SampleValue(1)},
						}},
				},
			},
			want: []*MonitoringAlert{{
				TimestampValue:   DateTime{sampleT.Time().Truncate(time.Hour)},
				NameValue:        "warn: hello",
				ServiceNameValue: "world",
				AverageValue:     1,
			}},
			wantErr: nil,
		}, {
			name: "elements are sorted",
			fields: fields{
				queryValue: model.Matrix{
					&model.SampleStream{
						Metric: model.Metric{"name": "b", "service_name": "b", "level": "warn"},
						Values: []model.SamplePair{
							{Timestamp: sampleT, Value: model.SampleValue(1)},
						},
					},
					&model.SampleStream{
						Metric: model.Metric{"name": "a", "service_name": "b", "level": "warn"},
						Values: []model.SamplePair{
							{Timestamp: sampleT, Value: model.SampleValue(1)},
							{Timestamp: sampleT.Add(time.Hour), Value: model.SampleValue(2)},
						},
					},
					&model.SampleStream{
						Metric: model.Metric{"name": "a", "service_name": "a", "level": "warn"},
						Values: []model.SamplePair{
							{Timestamp: sampleT, Value: model.SampleValue(1)},
						},
					},
				},
			},
			want: []*MonitoringAlert{{
				TimestampValue:   DateTime{sampleT.Time().Truncate(time.Hour)},
				NameValue:        "warn: a",
				ServiceNameValue: "a",
				AverageValue:     1,
			}, {
				TimestampValue:   DateTime{sampleT.Time().Truncate(time.Hour)},
				NameValue:        "warn: a",
				ServiceNameValue: "b",
				AverageValue:     1,
			}, {
				TimestampValue:   DateTime{sampleT.Time().Truncate(time.Hour)},
				NameValue:        "warn: b",
				ServiceNameValue: "b",
				AverageValue:     1,
			}, {
				TimestampValue:   DateTime{sampleT.Time().Add(time.Hour).Truncate(time.Hour)},
				NameValue:        "warn: a",
				ServiceNameValue: "b",
				AverageValue:     2,
			}},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.QueryRangeFunc.SetDefaultHook(func(ctx context.Context, query string, r prometheus.Range) (model.Value, prometheus.Warnings, error) {
				return tt.fields.queryValue, tt.fields.queryWarnings, tt.fields.queryErr
			})
			r := &siteMonitoringStatisticsResolver{
				prom:     mock,
				timespan: 24 * time.Hour,
			}
			alerts, err := r.Alerts(context.Background())
			if err != nil {
				if tt.wantErr == nil {
					t.Errorf("expected no error, got %v", err)
				} else if !errors.Is(err, tt.wantErr) {
					t.Errorf("expected error %v, got %v", tt.wantErr, err)
				}
			}
			if diff := cmp.Diff(tt.want, alerts); diff != "" {
				t.Errorf("alerts: %s", diff)
			}
		})
	}
}
