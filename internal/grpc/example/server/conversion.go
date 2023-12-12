// A good practice is to maintain a good separation between the gRPC-specific service implementation
// (which only handles any gRCP specifics like error handling) and the "internal" service which handles the actual
// business logic.
//
// This makes it easier to identify dependencies and test the business logic without having to worry about transport-specific
// details.
//
// One strategy to achieve this is to define a "conversion" layer that converts between the gRPC types and the internal types (this file).
// You can use fuzz tests to ensure that the conversion layer is correct even if the underlying types change (see conversion_test.go).
package main

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/grpc/example/server/service"
	pb "github.com/sourcegraph/sourcegraph/internal/grpc/example/weather/v1"
)

func WeatherResponseToGetCurrentWeatherProto(r *service.WeatherResponse) *pb.GetCurrentWeatherResponse {
	if r == nil {
		return nil
	}

	return &pb.GetCurrentWeatherResponse{
		Description: r.Description,
		Temperature: TemperatureToProto(r.Temperature),
	}
}

func WeatherResponseFromGetCurrentWeatherProto(r *pb.GetCurrentWeatherResponse) *service.WeatherResponse {
	if r == nil {
		return nil
	}

	return &service.WeatherResponse{
		Description: r.GetDescription(),
		Temperature: TemperatureFromProto(r.GetTemperature()),
	}
}

func WeatherResponseToRealTimeWeatherProto(r *service.WeatherResponse) *pb.RealTimeWeatherResponse {
	if r == nil {
		return nil
	}

	return &pb.RealTimeWeatherResponse{
		Description: r.Description,
		Temperature: TemperatureToProto(r.Temperature),
	}
}

func WeatherResponseFromRealTimeWeatherProto(r *pb.RealTimeWeatherResponse) *service.WeatherResponse {
	if r == nil {
		return nil
	}

	return &service.WeatherResponse{
		Description: r.GetDescription(),
		Temperature: TemperatureFromProto(r.GetTemperature()),
	}
}

func WeatherResponseFromProto(r *pb.GetCurrentWeatherResponse) *service.WeatherResponse {
	if r == nil {
		return nil
	}

	return &service.WeatherResponse{
		Description: r.GetDescription(),
		Temperature: TemperatureFromProto(r.GetTemperature()),
	}
}

func SensorOfflineErrorToProto(e *service.SensorOfflineError) *pb.SensorOfflineError {
	if e == nil {
		return nil
	}

	return &pb.SensorOfflineError{
		SensorId: e.SensorId,
		Message:  e.Message,
	}
}

func SensorOfflineErrorFromProto(e *pb.SensorOfflineError) *service.SensorOfflineError {
	if e == nil {
		return nil
	}

	return &service.SensorOfflineError{
		SensorId: e.GetSensorId(),
		Message:  e.GetMessage(),
	}
}

func WeatherAlertToProto(a *service.WeatherAlert) *pb.SubscribeWeatherAlertsResponse {
	if a == nil {
		return nil
	}

	return &pb.SubscribeWeatherAlertsResponse{
		Alert: a.Alert,
	}
}

func WeatherAlertFromProto(a *pb.SubscribeWeatherAlertsResponse) *service.WeatherAlert {
	if a == nil {
		return nil
	}

	return &service.WeatherAlert{
		Alert: a.GetAlert(),
	}
}

func UploadWeatherDataRequestToProto(a *service.SensorData) *pb.UploadWeatherDataRequest {
	if a == nil {
		return nil
	}

	return &pb.UploadWeatherDataRequest{
		SensorId:    a.SensorId,
		Humidity:    a.Humidity,
		Temperature: TemperatureToProto(a.Temperature),
	}
}

func SensorDataFromProto(a *pb.UploadWeatherDataRequest) *service.SensorData {
	if a == nil {
		return nil
	}

	return &service.SensorData{
		SensorId:    a.GetSensorId(),
		Humidity:    a.GetHumidity(),
		Temperature: TemperatureFromProto(a.GetTemperature()),
	}
}

func TemperatureToProto(a *service.Temperature) *pb.Temperature {
	if a == nil {
		return nil
	}

	var unit pb.Temperature_Unit
	switch a.Unit {
	case service.UNSPECIFIED:
		unit = pb.Temperature_UNIT_UNSPECIFIED
	case service.CELSIUS:
		unit = pb.Temperature_UNIT_CELSIUS
	case service.FAHRENHEIT:
		unit = pb.Temperature_UNIT_FAHRENHEIT
	case service.KELVIN:
		unit = pb.Temperature_UNIT_KELVIN
	default:
		panic(fmt.Sprintf("invalid temperature unit: %v", a.Unit))
	}

	return &pb.Temperature{
		Value: a.Value,
		Unit:  unit,
	}
}

func TemperatureFromProto(a *pb.Temperature) *service.Temperature {
	if a == nil {
		return nil
	}

	var unit service.TemperatureUnit
	switch a.GetUnit() {
	case pb.Temperature_UNIT_UNSPECIFIED:
		unit = service.UNSPECIFIED
	case pb.Temperature_UNIT_CELSIUS:
		unit = service.CELSIUS
	case pb.Temperature_UNIT_FAHRENHEIT:
		unit = service.FAHRENHEIT
	case pb.Temperature_UNIT_KELVIN:
		unit = service.KELVIN
	default:
		panic(fmt.Sprintf("invalid temperature unit: %v", a))
	}

	return &service.Temperature{
		Value: a.GetValue(),
		Unit:  unit,
	}
}
