package main

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	proto "github.com/sourcegraph/sourcegraph/internal/grpc/example/weather/v1"
)

// flakyWeatherService is a wrapper around a WeatherServiceServer that simulates
// a flaky service by returning an error on the first call to GetCurrentWeather.
type flakyWeatherService struct {
	getCurrentWeatherCallCounter atomic.Int64

	base proto.WeatherServiceServer
	proto.UnimplementedWeatherServiceServer
}

func (f *flakyWeatherService) GetCurrentWeather(ctx context.Context, request *proto.GetCurrentWeatherRequest) (*proto.GetCurrentWeatherResponse, error) {
	count := f.getCurrentWeatherCallCounter.Add(1)
	if count == 1 {
		return nil, status.Error(codes.Unavailable, "simulated service outage")
	}

	return f.base.GetCurrentWeather(ctx, request)
}

func (f *flakyWeatherService) SubscribeWeatherAlerts(request *proto.SubscribeWeatherAlertsRequest, server proto.WeatherService_SubscribeWeatherAlertsServer) error {
	return f.base.SubscribeWeatherAlerts(request, server)
}

func (f *flakyWeatherService) UploadWeatherData(server proto.WeatherService_UploadWeatherDataServer) error {
	return f.base.UploadWeatherData(server)
}

func (f *flakyWeatherService) RealTimeWeather(server proto.WeatherService_RealTimeWeatherServer) error {
	return f.base.RealTimeWeather(server)
}

func (f *flakyWeatherService) UploadWeatherPhoto(server proto.WeatherService_UploadWeatherPhotoServer) error {
	return f.base.UploadWeatherPhoto(server)
}

func (f *flakyWeatherService) GetCurrentWeatherOld(ctx context.Context, request *proto.GetCurrentWeatherOldRequest) (*proto.GetCurrentWeatherOldResponse, error) {
	return f.base.GetCurrentWeatherOld(ctx, request)
}

var _ proto.WeatherServiceServer = &flakyWeatherService{}
