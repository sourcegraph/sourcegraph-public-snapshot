package appliance

import (
	"context"

	pb "github.com/sourcegraph/sourcegraph/internal/appliance/v1"
)

func (a *Appliance) GetApplianceVersion(ctx context.Context, request *pb.GetApplianceVersionRequest) (*pb.GetApplianceVersionResponse, error) {
	return &pb.GetApplianceVersionResponse{Version: a.GetCurrentVersion(ctx)}, nil
}

func (a *Appliance) GetApplianceStatus(ctx context.Context, request *pb.GetApplianceStatusRequest) (*pb.GetApplianceStatusResponse, error) {
	status := a.GetCurrentStatus(ctx)

	return &pb.GetApplianceStatusResponse{Status: status.String()}, nil
}
