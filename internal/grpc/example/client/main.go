package main

import (
	"context"
	_ "embed"
	"io"
	"log"
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
	l := logger.Scoped("weather-client")

	conn, err := defaults.Dial("localhost:50051", l)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewWeatherServiceClient(conn)

	// Unary RPC: Normal case - get weather for a specific location

	weather, err := client.GetCurrentWeather(context.Background(), &pb.LocationRequest{Location: "New York"})
	if err != nil {
		log.Fatalf("Could not get weather: %v", err)
	}

	// We use the generated getter method to safety access the location
	// since there are no required fields in Protobuf messages:
	// The getters return the zero value for the type if the field is not set.
	//
	// See https://protobuf.dev/programming-guides/field_presence/ and https://stackoverflow.com/a/42634681 for more information.
	w, t := weather.GetDescription(), weather.GetTemperature()
	log.Printf("Weather in NYC - description: %s, temp: %v", w, t)

	// Unary RPC: Error case - get weather for a specific location (that doesn't exist for didactic purposes)
	weather, err = client.GetCurrentWeather(context.Background(), &pb.LocationRequest{Location: "Ravenholm"})

	log.Printf("This is what a gRPC status error looks like: %v", err)

	if status.Code(err) != codes.InvalidArgument { // You can extract the error code from the error object using the status.Code function, and then assert on it.
		log.Fatalf("Expected InvalidArgument error for going to Ravenholm, got %v, code: %s", err, status.Code(err))
	}

	weather, err = client.GetCurrentWeather(context.Background(), &pb.LocationRequest{Location: "Black Mesa"})
	s := status.Convert(err)
	if s.Code() != codes.Internal { // You can extract the error code from the error object using the status.Code function, and then assert on it.
		log.Fatalf("Expected Internal error for going to Black Mesa, got %v, code: %s", err, s.Code())
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
	weather, err = client.GetCurrentWeather(context.Background(), &pb.LocationRequest{Location: locationFromUnsanitizedUserInput})
	if err == nil {
		log.Fatalf("Expected error for invalid utf-8 input, got %v", weather)
	}
	log.Printf("got expected utf-8 string error when sending in unsanitized user input: %v", err)

	for _, d := range s.Details() {
		switch info := d.(type) {
		// You can also extract the error details from the error object using the status.Details function, and then assert on it.
		// This allows you to pass arbitrary structs over the wire with as much context as you need to do something in the application.
		case *pb.SensorOfflineError:
			log.Printf("Sensor %q is offline: %s", info.GetSensorId(), info.GetMessage())
		default:
			// If you don't recognize the error detail, you can log it / ignore it and move on.
			// This is helpful for forwards compatibility
			// (newer server versions may send new error details that older clients don't know about).
			log.Printf("Unexpected error detail: %v", info)
		}
	}

	// Server Streaming RPC: get weather alerts for a specific region
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second)) // Set a deadline for the RPC
	defer cancel()

	alertStream, err := client.SubscribeWeatherAlerts(ctx, &pb.AlertRequest{Region: "Midwest"})
	if err != nil {
		log.Fatalf("Error on subscribe weather alerts: %v", err)
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
				log.Fatalf("Error while receiving alert: %v", err)
			}
			log.Printf("Alert: %v", alert)
		}
	}

clientstreaming:

	// Client Streaming RPC: upload fake weather data
	dataStream, err := client.UploadWeatherData(context.Background())
	if err != nil {
		log.Fatalf("Error on upload weather data: %v", err)
	}
	for i := 0; i < 5; i++ {
		err := dataStream.Send(&pb.SensorData{
			SensorId: "sensor-123",
			Temperature: &pb.Temperature{
				Value: 26.5,
				Unit:  pb.Temperature_UNIT_CELSIUS,
			},
			Humidity: 80.0,
		})
		if err != nil {
			log.Fatalf("Error while sending data: %v", err)
		}
		time.Sleep(time.Second)
	}
	uploadStatus, err := dataStream.CloseAndRecv() // CloseAndRecv closes our end of the stream (indicating that we're doing sending) and returns the response from the server.
	if err != nil {
		log.Fatalf("Error while receiving upload status: %v", err)
	}
	log.Printf("Upload status: %s", uploadStatus.GetMessage())

	// Upload a weather screenshot using our helpers to send a byte stream.
	stream, err := client.UploadWeatherScreenshot(context.Background())
	if err != nil {
		log.Fatalf("Error on upload weather screenshot: %v", err)
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
		log.Fatalf("Error while sending image metadata: %v", err)
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
		log.Fatalf("Error while sending image data: %v", err)
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("Error while receiving upload status: %v", err)
	}

	log.Printf("Image upload status: %s", response.GetMessage())

	// Bidirectional Streaming RPC
	biStream, err := client.RealTimeWeather(context.Background())
	if err != nil {
		log.Fatalf("Error on real-time weather: %v", err)
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
				log.Fatalf("Error while receiving weather: %v", err)
				return
			}
			log.Printf("Real-time weather: %v", weather)
		}
	}()

	for i := 0; i < 5; i++ { // send location information to the server
		err := biStream.Send(&pb.LocationUpdate{
			Location: "Location " + strconv.Itoa(i),
		})
		if err != nil {
			log.Fatalf("Error while sending location update: %v", err)
		}
		time.Sleep(2 * time.Second)
	}

	err = biStream.CloseSend()
	if err != nil {
		log.Fatalf("Error while closing client end of bidirectional stream: %v", err)
	}

	wg.Wait()
}
