package main

import (
	"context"
	"io"
	"net"
	"time"

	logger "github.com/sourcegraph/log"
	pb "github.com/sourcegraph/sourcegraph/internal/grpc/example/weather/v1"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type weatherServer struct {
	logger logger.Logger

	// All gRPC services should embed the Unimplemented*Server structs to ensure forwards compatibility (if the service is
	// compiled against a newer version of the proto file, the server will still have default implementations of any new
	// RPCs).
	pb.UnimplementedWeatherServiceServer
}

// GetCurrentWeather is a Unary RPC (single request, single response) that returns the current weather for the requested location.
func (s *weatherServer) GetCurrentWeather(ctx context.Context, req *pb.LocationRequest) (*pb.WeatherResponse, error) {
	// We use the generated getter method to safety access the location since there are no required fields in Protobuf messages:
	// The getters return the zero value for the type if the field is not set.
	//
	// See https://protobuf.dev/programming-guides/field_presence/ and https://stackoverflow.com/a/42634681 for more information.
	location := req.GetLocation()

	if location == "Ravenholm" { // Let's pretend this location is not real to demonstrate error handling.

		// gRPC errors are Status objects, which contain an error code (akin to HTTP status codes: ), a message, and optional details.
		//
		// For well-known error cases, you can use the status.Errorf function to create a Status object with the appropriate
		// error code and message. Otherwise, any "anonymous" errors that don't implement the Status interface will be massaged
		// into a Status object with code "Unknown" and handled appropriately by the gRPC library.
		//
		// See the following for more background and information:
		// - https://avi.im/grpc-errors/#go
		// - https://godoc.org/google.golang.org/grpc/codes
		// - https://cloud.google.com/apis/design/errors (intended for Google developers, but generally applicable advice)
		err := status.Errorf(codes.InvalidArgument, "We don't go to Ravenholm.")
		return nil, err
	}

	// Now that we know that we have a "valid" location, we can safely return the response.
	response := &pb.WeatherResponse{
		Description: "It's always Sunny!",
		Temperature: 25.0,
	}

	// Note: The gRPC library handles serializing the response to the client - we can return the response struct directly.
	return response, nil
}

// SubscribeWeatherAlerts is a Server Streaming (aingle request, multiple responses) RPC that returns a stream of relevant weather alerts.
func (s *weatherServer) SubscribeWeatherAlerts(req *pb.AlertRequest, stream pb.WeatherService_SubscribeWeatherAlertsServer) error {
	ctx := stream.Context()
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done(): // The client either explicitly canceled the operation or the context timed out.

			// status.FromContextError is a convenience function that converts a context error
			// to a Status object (with code codes.Canceled or codes.DeadlineExceeded).
			//
			// Returning a proper status (instead of a nil error) here makes it clearer to the caller / service-wide observability tools that look at
			// response codes what exactly happened with this RPC call.
			return status.FromContextError(ctx.Err()).Err()

		case <-time.After(2 * time.Second): // Wait 2 seconds between the fake alerts we're sending.

			// Send a message to the client
			err := stream.Send(&pb.AlertResponse{
				Alert: "Severe weather alert!",
			})
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// UploadWeatherData is a long-running Client Streaming RPC (multiple request messages) is that is used to receive weather sensor data from a client.
func (s *weatherServer) UploadWeatherData(stream pb.WeatherService_UploadWeatherDataServer) error {
	ctx := stream.Context()
	for {
		if err := ctx.Err(); err != nil {
			// The client either explicitly canceled the operation, or the deadline expired.
			return status.FromContextError(err).Err()
		}

		data, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			// io.EOF is a sentinel value that indicates that the client has explicitly closed its end of the stream, which signals the end of the RPC.

			// We can use SendAndClose to send a final message to the client and close our end of the stream.
			return stream.SendAndClose(&pb.UploadStatus{
				Message: "Data received successfully",
			})
		}

		if err != nil {
			return errors.Wrap(err, "Failed to receive data from sensor")
		}

		s.logger.Info("Received data from sensor", logger.String("sensorID", data.SensorId))

		// (Extra logic to actually process the weather data goes here ...)
	}
}

// RealTimeWeather is a Bidirectional streaming RPC (multiple request messages, multiple response messages) that is used to
// receive location data from a client and respond with the current weather for the requested location.
func (s *weatherServer) RealTimeWeather(stream pb.WeatherService_RealTimeWeatherServer) error {
	ctx := stream.Context()
	for { // Loop until the client closes its end of the stream, or we encounter an error.

		if err := ctx.Err(); err != nil {
			// The client either explicitly canceled the operation, or the deadline expired.
			return status.FromContextError(err).Err()
		}

		locUpdate, err := stream.Recv()
		if errors.Is(err, io.EOF) { // The client has closed its end of the stream, so we can close our end as well.
			return nil
		}
		if err != nil {
			return err
		}

		description := "Cloudy"
		location := locUpdate.GetLocation()
		if location == "Philadelphia" {
			description = "Sunny!"
		}

		// Send a message back to the client with the current weather for the requested location.
		err = stream.Send(&pb.WeatherResponse{
			Description: description,
			Temperature: 22.0,
		})
		if err != nil {
			return err
		}
	}
}

func main() {
	l := logger.Scoped("weather-server")
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		l.Fatal("Failed to listen", logger.String("error", err.Error()))
	}

	s := grpc.NewServer()
	pb.RegisterWeatherServiceServer(s, &weatherServer{
		logger: l,
	})
	l.Info("Server listening", logger.String("address", lis.Addr().String()))

	if err := s.Serve(lis); err != nil {
		l.Fatal("Failed to serve", logger.String("error", err.Error()))
	}
}
