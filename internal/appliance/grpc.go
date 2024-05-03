package appliance

import (
	"context"

	pb "github.com/sourcegraph/sourcegraph/internal/appliance/v1"
)

func (a *Appliance) GetApplianceVersion(ctx context.Context, request *pb.GetApplianceVersionRequest) (*pb.GetApplianceVersionResponse, error) {
	version := a.GetCurrentVersion(ctx)

	return &pb.GetApplianceVersionResponse{Version: version.String()}, nil
}

func (a *Appliance) GetApplianceStage(ctx context.Context, request *pb.GetApplianceStageRequest) (*pb.GetApplianceStageResponse, error) {
	stage := a.GetCurrentStage(ctx)

	return &pb.GetApplianceStageResponse{Stage: stage.String()}, nil
}
