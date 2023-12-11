package main

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	logger "github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	pb "github.com/sourcegraph/sourcegraph/internal/grpc/example/weather/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:embed sun.png
var sunDrawingImage []byte

func main() {
	logger.Init(logger.Resource{
		Name: "weather-client",
	})

	l := logger.Scoped("weather-client")

	conn, err := defaults.Dial("localhost:50051", l)
	if err != nil {
		l.Fatal("Failed to connect", logger.Error(err))
	}
	defer conn.Close()

	client := pb.NewWeatherServiceClient(conn)

	// Unary RPC: Normal case - get weather for a specific location

	weather, err := client.GetCurrentWeather(context.Background(), &pb.GetCurrentWeatherRequest{Location: "New York"})
	if err != nil {
		l.Fatal("Could not get weather", logger.Error(err))
	}

	// We use the generated getter method to safety access the location
	// since there are no required fields in Protobuf messages:
	// The getters return the zero value for the type if the field is not set.
	//
	// See https://protobuf.dev/programming-guides/field_presence/ and https://stackoverflow.com/a/42634681 for more information.
	w, t := weather.GetDescription(), weather.GetTemperature()
	l.Debug("Weather in NYC", logger.String("description", w), logger.Float64("temperature", t.GetValue()))

	// Unary RPC: Error case - get weather for a specific location (that doesn't exist for didactic purposes)
	weather, err = client.GetCurrentWeather(context.Background(), &pb.GetCurrentWeatherRequest{Location: "Ravenholm"})

	l.Debug("This is what a gRPC status error looks like", logger.Error(err))

	if status.Code(err) != codes.InvalidArgument { // You can extract the error code from the error object using the status.Code function, and then assert on it.
		l.Fatal("Expected InvalidArgument error for going to Ravenholm", logger.Error(err), logger.String("code", status.Code(err).String()))
	}

	weather, err = client.GetCurrentWeather(context.Background(), &pb.GetCurrentWeatherRequest{Location: "Black Mesa"})
	s := status.Convert(err)
	if s.Code() != codes.Internal { // You can extract the error code from the error object using the status.Code function, and then assert on it.
		l.Fatal("Expected Internal error for going to Black Mesa", logger.Error(err), logger.String("code", s.Code().String()))
	}

	// Demonstrate that all string fields must be utf-8 encoded.
	//
	// The protobuf spec says that all strings must be utf-8 encoded (https://protobuf.dev/programming-guides/proto3/#scalar).
	// ("A string must always contain UTF-8 encoded or 7-bit ASCII text ...").
	//
	// When sending protobuf strings that come from user input,
	// you need to ensure that they are utf-8 encoded or otherwise sanitized. Otherwise, the gRPC library will return an error.
	//
	// If you can't guarantee that the string is utf-8 encoded, you can try to switch the type of the underlying field to bytes instead.
	// See https://github.com/sourcegraph/zoekt/pull/641 for an example of this.
	//
	// See https://github.com/sourcegraph/sourcegraph/issues/52181 for more context.
	locationFromUnsanitizedUserInput := "\x80\x80\x80\x80\x80\x80\x80\x80\x80\x80\x80\x80\x80\x80\x80"
	weather, err = client.GetCurrentWeather(context.Background(), &pb.GetCurrentWeatherRequest{Location: locationFromUnsanitizedUserInput})
	if err == nil {
		l.Fatal("Expected error for invalid utf-8 input, got nil", logger.String("location", locationFromUnsanitizedUserInput), logger.String("weather", weather.GetDescription()))
	}
	l.Debug("got expected utf-8 string error when sending in unsanitized user input", logger.Error(err))

	for _, d := range s.Details() {
		switch info := d.(type) {
		// You can also extract the error details from the error object using the status.Details function, and then assert on it.
		// This allows you to pass arbitrary structs over the wire with as much context as you need to do something in the application.
		case *pb.SensorOfflineError:
			l.Debug("Sensor is offline", logger.String("sensor_id", info.GetSensorId()), logger.String("message", info.GetMessage()))
		default:
			// If you don't recognize the error detail, you can log it / ignore it and move on.
			// This is helpful for forwards compatibility
			// (newer server versions may send new error details that older clients don't know about).
			l.Debug("Unexpected error detail", logger.String("info", fmt.Sprintf("%v", info)))
		}
	}

	// Server Streaming RPC: get weather alerts for a specific region
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second)) // Set a deadline for the RPC
	defer cancel()

	alertStream, err := client.SubscribeWeatherAlerts(ctx, &pb.SubscribeWeatherAlertsRequest{Region: "Midwest"})
	if err != nil {
		l.Fatal("Error on subscribe weather alerts", logger.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			goto clientstreaming

		default:
			alert, err := alertStream.Recv()
			if errors.Is(err, io.EOF) {
				goto clientstreaming // The server closed the stream
			}

			if ctx.Err() != nil {
				goto clientstreaming // We canceled the RPC ourselves
			}

			if err != nil {
				l.Fatal("Error while receiving alert", logger.Error(err))
			}
			l.Debug("Alert", logger.String("alert", alert.GetAlert()))
		}
	}

clientstreaming:

	// Client Streaming RPC: upload fake weather data
	dataStream, err := client.UploadWeatherData(context.Background())
	if err != nil {
		l.Fatal("Error on upload weather data", logger.Error(err))
	}
	for i := 0; i < 5; i++ {
		err := dataStream.Send(&pb.UploadWeatherDataRequest{
			SensorId: "sensor-123",
			Temperature: &pb.Temperature{
				Value: 26.5,
				Unit:  pb.Temperature_UNIT_CELSIUS,
			},
			Humidity: 80.0,
		})
		if err != nil {
			l.Fatal("Error while sending data", logger.Error(err))
		}
		time.Sleep(time.Second)
	}
	uploadStatus, err := dataStream.CloseAndRecv() // CloseAndRecv closes our end of the stream (indicating that we're doing sending) and returns the response from the server.
	if err != nil {
		l.Fatal("Error while receiving upload status", logger.Error(err))
	}
	l.Debug("Upload status", logger.String("message", uploadStatus.GetMessage()))

	// Upload a weather screenshot using our helpers to send a byte stream.
	stream, err := client.UploadWeatherScreenshot(context.Background())
	if err != nil {
		l.Fatal("Error on upload weather screenshot", logger.Error(err))
	}

	err = stream.Send(&pb.UploadWeatherScreenshotRequest{
		Content: &pb.UploadWeatherScreenshotRequest_Metadata_{
			Metadata: &pb.UploadWeatherScreenshotRequest_Metadata{
				Location: "Philadelphia",
				SensorId: "sensor-123",
				FileName: "sun.png",
			},
		},
	})

	if err != nil {
		l.Fatal("Error while sending image metadata", logger.Error(err))
	}

	// Send the image data in chunks.
	// We have a helper package named "streamio" that chunks a byte stream into smaller gRPC messages that can
	// be incrementally handled by the server.
	writer := streamio.NewWriter(func(p []byte) error {
		return stream.Send(&pb.UploadWeatherScreenshotRequest{
			Content: &pb.UploadWeatherScreenshotRequest_Payload_{
				Payload: &pb.UploadWeatherScreenshotRequest_Payload{
					Data: p,
				},
			},
		})
	})

	_, err = writer.Write(sunDrawingImage)
	if err != nil {
		l.Fatal("Error while sending image data", logger.Error(err))
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		l.Fatal("Error while receiving upload status", logger.Error(err))
	}

	l.Debug("Image upload status", logger.String("message", response.GetMessage()))

	// Bidirectional Streaming RPC
	biStream, err := client.RealTimeWeather(context.Background())
	if err != nil {
		l.Fatal("Error on real-time weather", logger.Error(err))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Receive messages from the server in a separate goroutine
		for {
			weather, err := biStream.Recv()
			if errors.Is(err, io.EOF) {
				return
			}
			if err != nil {
				l.Fatal("Error while receiving weather", logger.Error(err))
				return
			}
			l.Debug("Real-time weather", logger.String("weather", weather.GetDescription()))
		}
	}()

	for i := 0; i < 5; i++ { // send location information to the server
		err := biStream.Send(&pb.RealTimeWeatherRequest{
			Location: "Location " + strconv.Itoa(i),
		})
		if err != nil {
			l.Fatal("Error while sending location update", logger.Error(err))
		}
		time.Sleep(2 * time.Second)
	}

	err = biStream.CloseSend()
	if err != nil {
		l.Fatal("Error while closing client end of bidirectional stream", logger.Error(err))
	}

	wg.Wait()
}
