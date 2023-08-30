package v1

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (e *CreateCommitFromPatchError) ToStatus() *status.Status {
	s, err := status.New(codes.Internal, e.InternalError).WithDetails(e)
	if err != nil {
		return status.New(codes.Internal, errors.Wrap(err, "failed to add details to status").Error())
	}
	return s
}
