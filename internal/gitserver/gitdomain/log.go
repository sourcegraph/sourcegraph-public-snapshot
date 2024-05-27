package gitdomain

import (
	"fmt"

	proto "github.com/sourcegraph/sourcegraph/internal/gitserver/v1"
)

type PathStatus struct {
	Path   string
	Status Status
}

func (s *PathStatus) String() string {
	return fmt.Sprintf("%s %q", string(s.Status), s.Path)
}

func PathStatusFromProto(p *proto.ChangedFile) PathStatus {
	return PathStatus{
		Path:   string(p.Path),
		Status: StatusFromProto(p.Status),
	}
}

func (p *PathStatus) ToProto() *proto.ChangedFile {
	return &proto.ChangedFile{
		Path:   []byte(p.Path),
		Status: p.Status.ToProto(),
	}
}

type Status byte

const (
	StatusUnspecified Status = 'X'
	StatusAdded       Status = 'A'
	StatusModified    Status = 'M'
	StatusDeleted     Status = 'D'
	StatusTypeChanged Status = 'T'
	// TODO: add other possible statuses here (see man git-diff-tree)
)

func StatusFromProto(s proto.ChangedFile_Status) Status {
	switch s {
	case proto.ChangedFile_STATUS_ADDED:
		return StatusAdded
	case proto.ChangedFile_STATUS_MODIFIED:
		return StatusModified
	case proto.ChangedFile_STATUS_DELETED:
		return StatusDeleted
	case proto.ChangedFile_STATUS_TYPE_CHANGED:
		return StatusTypeChanged
	default:
		return StatusUnspecified
	}
}

func (s Status) ToProto() proto.ChangedFile_Status {
	switch s {
	case StatusAdded:
		return proto.ChangedFile_STATUS_ADDED
	case StatusModified:
		return proto.ChangedFile_STATUS_MODIFIED
	case StatusDeleted:
		return proto.ChangedFile_STATUS_DELETED
	case StatusTypeChanged:
		return proto.ChangedFile_STATUS_TYPE_CHANGED
	default:
		return proto.ChangedFile_STATUS_UNSPECIFIED
	}
}
