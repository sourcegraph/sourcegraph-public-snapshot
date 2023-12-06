package main

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/grpc/example/server/service"
	pb "github.com/sourcegraph/sourcegraph/internal/grpc/example/weather/v1"
)

func WeatherResponseToProto(r *service.WeatherResponse) *pb.WeatherResponse {
	if r == nil {
		return nil
	}

	return &pb.WeatherResponse{
		Description: r.Description,
		Temperature: TemperatureToProto(r.Temperature),
	}
}

func WeatherResponseFromProto(r *pb.WeatherResponse) *service.WeatherResponse {
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

func WeatherAlertToProto(a *service.WeatherAlert) *pb.AlertResponse {
	if a == nil {
		return nil
	}

	return &pb.AlertResponse{
		Alert: a.Alert,
	}
}

func WeatherAlertFromProto(a *pb.AlertResponse) *service.WeatherAlert {
	if a == nil {
		return nil
	}

	return &service.WeatherAlert{
		Alert: a.GetAlert(),
	}
}

func SensorDataToProto(a *service.SensorData) *pb.SensorData {
	if a == nil {
		return nil
	}

	return &pb.SensorData{
		SensorId:    a.SensorId,
		Humidity:    a.Humidity,
		Temperature: TemperatureToProto(a.Temperature),
	}
}

func SensorDataFromProto(a *pb.SensorData) *service.SensorData {
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
