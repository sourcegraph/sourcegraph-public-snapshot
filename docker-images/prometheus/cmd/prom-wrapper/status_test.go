package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	prometheus "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"

	"github.com/sourcegraph/sourcegraph/docker-images/prometheus/cmd/prom-wrapper/mocks"
	srcprometheus "github.com/sourcegraph/sourcegraph/internal/src-prometheus"
)

func TestAlertsStatusReporterHistory(t *testing.T) {
	mock := mocks.NewMockAPI()
	sampleT := model.Time(time.Now().UTC().Unix())
	type fields struct {
		queryValue    model.Value
		queryWarnings prometheus.Warnings
		queryErr      error
	}
	tests := []struct {
		name   string
		fields fields
		want   srcprometheus.MonitoringAlerts
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
			want: []*srcprometheus.MonitoringAlert{{
				TimestampValue:   sampleT.Time().Truncate(time.Hour),
				NameValue:        "warn: hello",
				ServiceNameValue: "world",
				AverageValue:     0,
			}},
		}, {
			name: "includes alerts with occurrences",
			fields: fields{
				queryValue: model.Matrix{
					&model.SampleStream{
						Metric: model.Metric{"name": "hello", "service_name": "world", "level": "warn"},
						Values: []model.SamplePair{{Timestamp: sampleT, Value: model.SampleValue(1)}}},
				},
			},
			want: []*srcprometheus.MonitoringAlert{{
				TimestampValue:   sampleT.Time().Truncate(time.Hour),
				NameValue:        "warn: hello",
				ServiceNameValue: "world",
				AverageValue:     1,
			}},
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
			want: []*srcprometheus.MonitoringAlert{{
				TimestampValue:   sampleT.Time().Truncate(time.Hour),
				NameValue:        "warn: hello",
				ServiceNameValue: "world",
				AverageValue:     1,
			}},
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
			want: []*srcprometheus.MonitoringAlert{{
				TimestampValue:   sampleT.Time().Truncate(time.Hour),
				NameValue:        "warn: a",
				ServiceNameValue: "a",
				AverageValue:     1,
			}, {
				TimestampValue:   sampleT.Time().Truncate(time.Hour),
				NameValue:        "warn: a",
				ServiceNameValue: "b",
				AverageValue:     1,
			}, {
				TimestampValue:   sampleT.Time().Truncate(time.Hour),
				NameValue:        "warn: b",
				ServiceNameValue: "b",
				AverageValue:     1,
			}, {
				TimestampValue:   sampleT.Time().Add(time.Hour).Truncate(time.Hour),
				NameValue:        "warn: a",
				ServiceNameValue: "b",
				AverageValue:     2,
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock.QueryRangeFunc.SetDefaultHook(func(ctx context.Context, query string, r prometheus.Range) (model.Value, prometheus.Warnings, error) {
				return tt.fields.queryValue, tt.fields.queryWarnings, tt.fields.queryErr
			})
			r := &AlertsStatusReporter{prometheus: mock}
			recorder := httptest.NewRecorder()
			r.Handler().ServeHTTP(recorder, &http.Request{
				Method: http.MethodGet,
				URL:    &url.URL{Path: srcprometheus.EndpointAlertsStatusHistory},
			})
			var history srcprometheus.AlertsHistory
			if err := json.NewDecoder(recorder.Result().Body).Decode(&history); err != nil {
				t.Errorf("expected no error, got %v", err)
				return
			}
			if diff := cmp.Diff(tt.want, history.Alerts); diff != "" {
				t.Errorf("alerts: %s", diff)
			}
		})
	}
}
