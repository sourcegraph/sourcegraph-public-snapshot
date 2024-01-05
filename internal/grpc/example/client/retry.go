package main

import (
	"context"

	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	proto "github.com/sourcegraph/sourcegraph/internal/grpc/example/weather/v1"
)

// automaticRetryClient is a convenience wrapper around a base proto.WeatherServiceClient that automatically retries
// idempotent ("safe") methods in accordance to the policy defined in internal/grpc/defaults.RetryPolicy.
//
// Read the implementation of this type for more details on what kinds of RPCs are automatically retried (and why).
//
// Callers are free to override the default retry behavior by proving their own grpc.CallOptions when invoking the RPC.
// (example: providing retry.WithMax(0) will disable retries even when invoking GetCurrentWeather - which is idempotent).
type automaticRetryClient struct {
	base proto.WeatherServiceClient
}

// Unsupported methods.

func (a *automaticRetryClient) UploadWeatherData(ctx context.Context, opts ...grpc.CallOption) (proto.WeatherService_UploadWeatherDataClient, error) {
	// UploadWeatherData is a client streaming method, which isn't supported by our automatic retry logic.
	// Trying to use our automatic retry logic with this method immediately returns the following error (fail fast):
	//
	// code = Unimplemented desc = grpc_retry: cannot retry on ClientStreams, set grpc_retry.Disable()
	return a.base.UploadWeatherData(ctx, opts...)
}

func (a *automaticRetryClient) RealTimeWeather(ctx context.Context, opts ...grpc.CallOption) (proto.WeatherService_RealTimeWeatherClient, error) {
	// RealTimeWeather is a bidirectional streaming method, which isn't supported by our automatic retry logic.
	// Trying to use our automatic retry logic with this method immediately returns the following error (fail fast):
	//
	// code = Unimplemented desc = grpc_retry: cannot retry on ClientStreams, set grpc_retry.Disable()
	return a.base.RealTimeWeather(ctx, opts...)
}

// Non idempotent methods.

func (a *automaticRetryClient) UploadWeatherPhoto(ctx context.Context, opts ...grpc.CallOption) (proto.WeatherService_UploadWeatherPhotoClient, error) {
	// UploadWeatherPhoto's RPC documentation states that it is not idempotent because it doesn't do any deduplication.
	return a.base.UploadWeatherPhoto(ctx, opts...)
}

// Idempotent methods.

func (a *automaticRetryClient) GetCurrentWeather(ctx context.Context, in *proto.GetCurrentWeatherRequest, opts ...grpc.CallOption) (*proto.GetCurrentWeatherResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.GetCurrentWeather(ctx, in, opts...)
}

func (a *automaticRetryClient) SubscribeWeatherAlerts(ctx context.Context, in *proto.SubscribeWeatherAlertsRequest, opts ...grpc.CallOption) (proto.WeatherService_SubscribeWeatherAlertsClient, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.SubscribeWeatherAlerts(ctx, in, opts...)
}

func (a *automaticRetryClient) GetCurrentWeatherOld(ctx context.Context, in *proto.GetCurrentWeatherOldRequest, opts ...grpc.CallOption) (*proto.GetCurrentWeatherOldResponse, error) {
	opts = append(defaults.RetryPolicy, opts...)
	return a.base.GetCurrentWeatherOld(ctx, in, opts...)
}

var _ proto.WeatherServiceClient = &automaticRetryClient{}
