package appliance

import (
	"context"

	pb "github.com/sourcegraph/sourcegraph/internal/appliance/v1"
)

type GRPCServer struct {
	// Embed the UnimplementedApplianceServiceServer structs to ensure forwards compatibility (if the service is
	// compiled against a newer version of the proto file, the server will still have default implementations of any new
	// RPCs).
	pb.UnimplementedApplianceServiceServer
}

func (s *GRPCServer) GetApplianceVersion(ctx context.Context, request *pb.GetApplianceVersionRequest) (*pb.GetApplianceVersionResponse, error) {
	return nil, nil
}

func (s *GRPCServer) GetApplianceStage(ctx context.Context, request *pb.GetApplianceStageRequest) (*pb.GetApplianceStageResponse, error) {
	return nil, nil
}
