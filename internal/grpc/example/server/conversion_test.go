package main

import (
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/grpc/example/server/service"
)

func TestRoundTripWeatherResponse(t *testing.T) {
	t.Run("current weather", func(t *testing.T) {
		diff := ""

		err := quick.Check(func(original *service.WeatherResponse) bool {
			if original != nil && original.Temperature != nil {
				original.Temperature.Unit = original.Temperature.Unit % 4
				if original.Temperature.Unit < 0 {
					original.Temperature.Unit = -original.Temperature.Unit
				}
			}

			converted := WeatherResponseFromGetCurrentWeatherProto(WeatherResponseToGetCurrentWeatherProto(original))
			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}, nil)
		if err != nil {
			t.Fatalf("unexpected diff in weather response (-want +got):\n%s", diff)
		}
	})

	t.Run("real time weather", func(t *testing.T) {
		diff := ""

		err := quick.Check(func(original *service.WeatherResponse) bool {
			if original != nil && original.Temperature != nil {
				original.Temperature.Unit = original.Temperature.Unit % 4
				if original.Temperature.Unit < 0 {
					original.Temperature.Unit = -original.Temperature.Unit
				}
			}

			converted := WeatherResponseFromRealTimeWeatherProto(WeatherResponseToRealTimeWeatherProto(original))
			if diff = cmp.Diff(original, converted); diff != "" {
				return false
			}

			return true
		}, nil)
		if err != nil {
			t.Fatalf("unexpected diff in weather response (-want +got):\n%s", diff)
		}
	})

}

func TestRoundTripSensorOfflineError(t *testing.T) {
	diff := ""

	err := quick.Check(func(original *service.SensorOfflineError) bool {
		converted := SensorOfflineErrorFromProto(SensorOfflineErrorToProto(original))
		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}, nil)
	if err != nil {
		t.Fatalf("unexpected diff in sensor offline error (-want +got):\n%s", diff)
	}
}

func TestRoundtripWeatherAlert(t *testing.T) {
	diff := ""

	err := quick.Check(func(original *service.WeatherAlert) bool {
		converted := WeatherAlertFromProto(WeatherAlertToProto(original))
		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}, nil)
	if err != nil {
		t.Fatalf("unexpected diff in weather alert (-want +got):\n%s", diff)
	}
}

func TestRoundtripSensorData(t *testing.T) {
	diff := ""

	err := quick.Check(func(original *service.SensorData) bool {
		if original != nil && original.Temperature != nil {
			original.Temperature.Unit = original.Temperature.Unit % 3
			if original.Temperature.Unit < 0 {
				original.Temperature.Unit = -original.Temperature.Unit
			}
		}

		converted := SensorDataFromProto(UploadWeatherDataRequestToProto(original))
		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}, nil)
	if err != nil {
		t.Fatalf("unexpected diff in sensor data (-want +got):\n%s", diff)
	}
}
