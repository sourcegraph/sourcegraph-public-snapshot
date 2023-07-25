package frontend

import (
	proto "github.com/sourcegraph/sourcegraph/internal/frontend/v1"
)

type ExternalServiceConfigsRequest struct {
	Kind    string `json:"kind"`
	Limit   int    `json:"limit"`
	AfterID int    `json:"after_id"`
}

func (r *ExternalServiceConfigsRequest) ToProto() *proto.ExternalServiceConfigsRequest {
	return &proto.ExternalServiceConfigsRequest{
		Kind:    r.Kind,
		Limit:   int64(r.Limit),
		AfterId: int64(r.AfterID),
	}
}

func (r *ExternalServiceConfigsRequest) FromProto(req *proto.ExternalServiceConfigsRequest) {
	*r = ExternalServiceConfigsRequest{
		Kind:    req.GetKind(),
		Limit:   int(req.GetLimit()),
		AfterID: int(req.GetAfterId()),
	}
}
