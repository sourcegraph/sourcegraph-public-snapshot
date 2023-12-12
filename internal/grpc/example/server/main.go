package main

import (
	"context"
	"io"
	"net"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/grpc/example/server/service"
	pb "github.com/sourcegraph/sourcegraph/internal/grpc/example/weather/v1"
	"github.com/sourcegraph/sourcegraph/internal/grpc/streamio"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type weatherGRPCServer struct {
	logger log.Logger

	// The actual business logic is implemented in the service package, which we're wrapping here.
	service *service.WeatherService

	// All gRPC services should embed the Unimplemented*Server structs to ensure forwards compatibility (if the service is
	// compiled against a newer version of the proto file, the server will still have default implementations of any new
	// RPCs).
	pb.UnimplementedWeatherServiceServer
}

// GetCurrentWeather is a Unary RPC (single request, single response) that returns the current weather for the requested location.
func (s *weatherGRPCServer) GetCurrentWeather(ctx context.Context, req *pb.GetCurrentWeatherRequest) (*pb.GetCurrentWeatherResponse, error) {
	// We use the generated getter method to safety access the location since there are no required fields in Protobuf messages:
	// The getters return the zero value for the type if the field is not set.
	//
	// See https://protobuf.dev/programming-guides/field_presence/ and https://stackoverflow.com/a/42634681 for more information.
	location := req.GetLocation()

	response, err := s.service.GetCurrentWeather(ctx, location)
	if err != nil {
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

		var invalidPlaceErr *service.InvalidPlaceError
		if errors.As(err, &invalidPlaceErr) {
			// The client requested a location that doesn't exist.
			//
			// We can use the status.Errorf function to create a Status object with the appropriate
			// error code and message.
			return nil, status.Errorf(codes.InvalidArgument, "invalid place: %s", invalidPlaceErr.Place)
		}

		var sensorOfflineErr *service.SensorOfflineError
		if errors.As(err, &sensorOfflineErr) {
			// The client requested a location that doesn't exist.
			//
			// We can use the status.Errorf function to create a Status object with the appropriate
			// error code and message.
			s, err := status.New(codes.Internal, "The resonance cascade has begun.").WithDetails(SensorOfflineErrorToProto(sensorOfflineErr))
			if err != nil {
				return nil, errors.Wrap(err, "failed to create gRPC status")
			}

			return nil, s.Err()
		}

		// If we don't recognize the error, we can return it directly. The gRPC library will massage it into a Status object
		// with code "Unknown" and handle it appropriately.
		return nil, err
	}

	return WeatherResponseToGetCurrentWeatherProto(response), nil
}

// SubscribeWeatherAlerts is a Server Streaming (single request, multiple responses) RPC that returns a stream of relevant weather alerts.
func (s *weatherGRPCServer) SubscribeWeatherAlerts(req *pb.SubscribeWeatherAlertsRequest, stream pb.WeatherService_SubscribeWeatherAlertsServer) error {
	ctx := stream.Context()
	callback := func(a *service.WeatherAlert) error {
		// Send a message to the client
		err := stream.Send(WeatherAlertToProto(a))
		if err != nil {
			// We don't need to explicitly assign a gRPC status code to issues that occur while sending,
			// since the gRPC library generates the error and will most likely already have set the appropriate error code.
			return errors.Wrap(err, "failed to send alert across gRPC stream")
		}

		return nil
	}

	err := s.service.SubscribeWeatherAlerts(ctx, req.GetRegion(), callback)
	if err != nil {
		if ctx.Err() != nil {
			// status.FromContextError is a convenience function that converts a context error
			// to a Status object (with code codes.Canceled or codes.DeadlineExceeded).
			//
			// Returning a proper status (instead of a nil error) here makes it clearer to the caller / service-wide observability tools that look at
			// response codes what exactly happened with this RPC call.
			return status.FromContextError(ctx.Err()).Err()
		}

		return err
	}

	return nil
}

// UploadWeatherData is a long-running Client Streaming RPC (multiple request messages) is that is used to receive weather sensor data from a client.
func (s *weatherGRPCServer) UploadWeatherData(stream pb.WeatherService_UploadWeatherDataServer) error {
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
			return stream.SendAndClose(&pb.UploadWeatherDataResponse{
				Message: "Data received successfully",
			})
		}

		if err != nil {
			return errors.Wrap(err, "Failed to receive data from sensor")
		}

		s.logger.Info("Received data from sensor", log.String("sensorID", data.SensorId))

		if err := s.service.StoreSensorData(ctx, SensorDataFromProto(data)); err != nil {
			return status.Errorf(codes.Internal, "failed to store sensor data:%v", err)
		}
	}
}

// RealTimeWeather is a Bidirectional streaming RPC (multiple request messages, multiple response messages) that is used to
// receive location data from a client and respond with the current weather for the requested location.
func (s *weatherGRPCServer) RealTimeWeather(stream pb.WeatherService_RealTimeWeatherServer) error {
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

		location := locUpdate.GetLocation()

		weather, err := s.service.GetCurrentWeather(ctx, location)
		if err != nil {
			// On an error, we stop the bidi-stream.
			// Alternative patterns could be:
			// - To collect all the errors in a multierror and report the failed calls to service.GetCurrentWeather at the
			//   end of the stream.
			// - Make the response type have a oneOf field to encode a failed attempt.
			return status.Errorf(codes.Internal, "failed to get weather for %s: %v", location, err)
		}

		// Send a message back to the client with the current weather for the requested location.
		err = stream.Send(WeatherResponseToRealTimeWeatherProto(weather))
		if err != nil {
			return err
		}
	}
}

func (w *weatherGRPCServer) UploadWeatherPhoto(stream pb.WeatherService_UploadWeatherPhotoServer) error {
	msg, err := stream.Recv()
	if err != nil {
		return errors.Wrap(err, "failed to receive screenshot metadata")
	}

	// First, we pull the first message off the stream, which should be the image metadata.
	switch v := msg.GetContent().(type) {
	case *pb.UploadWeatherPhotoRequest_Metadata_:
		w.logger.Info("Received screenshot metadata",
			log.String("sensorID", v.Metadata.GetSensorId()),
			log.String("filename", v.Metadata.GetFileName()),
			log.String("location", v.Metadata.GetLocation()))
	default:
		return errors.Errorf("expected first message to be image metadata, instead got unexpected message type %T", v)
	}

	// Next, we read the image data from the stream.
	// We have an abstraction called streamio.Reader that allows us to read from a gRPC stream that contains bytes as if it were an io.Reader.
	// This is useful because it allows us to hide the mechanics of fetching more gRPC messages from users of the reader.
	reader := streamio.NewReader(func() ([]byte, error) {
		msg, err := stream.Recv()
		if err != nil {
			return nil, err
		}

		switch v := msg.GetContent().(type) {
		case *pb.UploadWeatherPhotoRequest_Payload_:
			return v.Payload.GetData(), nil
		default:
			// Note: Depending on your use-case, you could choose to ignore unknown message types and just continue reading.
			// This could be beneficial if you want to add new message types in the future without breaking existing clients.
			return nil, errors.Errorf("unexpected message type %T", v)
		}
	})

	err = w.service.StoreWeatherPhoto(reader)
	if err != nil {
		return errors.Wrap(err, "failed to store weather screenshot")
	}

	// Once we've read all the data from the stream, we can send a final message to the client and close our end of the stream.
	return stream.SendAndClose(&pb.UploadWeatherPhotoResponse{
		Message: "Screenshot received successfully.",
	})
}

func main() {
	log.Init(log.Resource{
		Name: "weather-server",
	})
	logger := log.Scoped("weather-server")

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("Failed to listen", log.String("error", err.Error()))
	}

	s := grpc.NewServer()
	pb.RegisterWeatherServiceServer(s, &weatherGRPCServer{
		logger: logger,
	})
	logger.Info("Server listening", log.String("address", lis.Addr().String()))

	if err := s.Serve(lis); err != nil {
		logger.Fatal("Failed to serve", log.String("error", err.Error()))
	}
}
