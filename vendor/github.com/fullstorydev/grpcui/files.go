package grpcui

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/grpcreflect"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	rpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/fullstorydev/grpcurl"
)

// AllFilesViaReflection returns a slice that contains the file descriptors
// for all methods exposed by the server on the other end of the given
// connection. This returns an error if the server does not support service
// reflection. (See "google.golang.org/grpc/reflection" for more on service
// reflection.)
func AllFilesViaReflection(ctx context.Context, cc grpc.ClientConnInterface) ([]*desc.FileDescriptor, error) {
	stub := rpb.NewServerReflectionClient(cc)
	cli := grpcreflect.NewClient(ctx, stub)
	source := grpcurl.DescriptorSourceFromServer(ctx, cli)
	return grpcurl.GetAllFiles(source)
}

// AllFilesViaInProcess returns a slice that contains all file descriptors
// known to this server process. This collects descriptors for all files
// registered with protoregistry.GlobalFiles, which includes all compiled
// proto files linked into the current program.
func AllFilesViaInProcess() ([]*desc.FileDescriptor, error) {
	var fds []*desc.FileDescriptor
	var err error
	protoregistry.GlobalFiles.RangeFiles(func(d protoreflect.FileDescriptor) bool {
		var fd *desc.FileDescriptor
		fd, err = desc.LoadFileDescriptor(d.Path())
		if err != nil {
			return false
		}
		fds = append(fds, fd)
		return true
	})
	if err != nil {
		return nil, err
	}
	return fds, nil
}
