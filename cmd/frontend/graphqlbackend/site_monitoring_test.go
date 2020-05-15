package graphqlbackend

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func Test_siteMonitoringStatisticsResolver_Alerts(t *testing.T) {
	mock := NewMockPrometheusQuerier()
	sampleT := model.Time(time.Now().Unix())
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
		{"discards alerts with no occurrences", fields{
			queryValue: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"name": "hello", "service_name": "world"},
					Values: []model.SamplePair{{Timestamp: sampleT, Value: model.SampleValue(0)}}},
			},
		}, []*MonitoringAlert{}, nil},
		{"includes alerts with occurrences", fields{
			queryValue: model.Matrix{
				&model.SampleStream{
					Metric: model.Metric{"name": "hello", "service_name": "world"},
					Values: []model.SamplePair{{Timestamp: sampleT, Value: model.SampleValue(1)}}},
			},
		}, []*MonitoringAlert{{
			TimestampValue:   DateTime{sampleT.Time()},
			NameValue:        "hello",
			ServiceNameValue: "world",
		}}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.QueryRangeFunc.SetDefaultHook(func(ctx context.Context, query string, r prometheus.Range) (model.Value, prometheus.Warnings, error) {
				return tt.fields.queryValue, tt.fields.queryWarnings, tt.fields.queryErr
			})
			r := &siteMonitoringStatisticsResolver{
				ctx:      context.Background(),
				prom:     mock,
				timespan: 24 * time.Hour,
			}
			alerts, err := r.Alerts()
			if err != nil {
				if tt.wantErr != nil {
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
