// Copyright 2022-2023 Buf Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package grpcreflect enables any net/http server, including those built with
// Connect, to handle gRPC's server reflection API. This lets ad-hoc debugging
// tools call your Protobuf services and print the responses without a copy of
// the schema.
//
// The exposed reflection API is wire compatible with Google's gRPC
// implementations, so it works with grpcurl, grpcui, BloomRPC, and many other
// tools.
//
// The core Connect package is connectrpc.com/connect. Documentation is
// available at https://connectrpc.com.
package grpcreflect

import (
	"context"
	_ "embed" // required for go:embed directive
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"

	"connectrpc.com/connect"
	reflectionv1 "connectrpc.com/grpcreflect/internal/gen/go/connectext/grpc/reflection/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protodesc"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/descriptorpb"
)

const (
	// ReflectV1ServiceName is the fully-qualified name of the v1 version of the reflection service.
	ReflectV1ServiceName = "grpc.reflection.v1.ServerReflection"
	// ReflectV1AlphaServiceName is the fully-qualified name of the v1alpha version of the reflection service.
	ReflectV1AlphaServiceName = "grpc.reflection.v1alpha.ServerReflection"

	serviceURLPathV1      = "/" + ReflectV1ServiceName + "/"
	serviceURLPathV1Alpha = "/" + ReflectV1AlphaServiceName + "/"
	methodName            = "ServerReflectionInfo"
)

//nolint:gochecknoglobals
var (
	//go:embed services.bin
	embeddedDescriptors []byte

	globalFiles = resolverHackForConnectext(embeddedDescriptors)
)

// NewHandlerV1 constructs an implementation of v1 of the gRPC server reflection
// API. It returns an HTTP handler and the path on which to mount it.
//
// Note that because the reflection API requires bidirectional streaming, the
// returned handler doesn't support HTTP/1.1. If your server must also support
// older tools that use the v1alpha server reflection API, see NewHandlerV1Alpha.
func NewHandlerV1(reflector *Reflector, options ...connect.HandlerOption) (string, http.Handler) {
	return newHandler(reflector, serviceURLPathV1, options)
}

// NewHandlerV1Alpha constructs an implementation of v1alpha of the gRPC server
// reflection API, which is useful to support tools that haven't updated to the
// v1 API. It returns an HTTP handler and the path on which to mount it.
//
// Like NewHandlerV1, the returned handler doesn't support HTTP/1.1.
func NewHandlerV1Alpha(reflector *Reflector, options ...connect.HandlerOption) (string, http.Handler) {
	// v1 is binary-compatible with v1alpha, so we only need to change paths.
	return newHandler(reflector, serviceURLPathV1Alpha, options)
}

// Reflector implements the underlying logic for gRPC's protobuf server
// reflection. They're configurable, so they can support straightforward
// process-local reflection or more complex proxying.
//
// Keep in mind that by default, Reflectors expose every protobuf type and
// extension compiled into your binary. Think twice before including the
// default Reflector in a public API.
//
// For more information, see
// https://github.com/grpc/grpc-go/blob/master/Documentation/server-reflection-tutorial.md,
// https://github.com/grpc/grpc/blob/master/doc/server-reflection.md, and
// https://github.com/fullstorydev/grpcurl.
type Reflector struct {
	namer              Namer
	extensionResolver  ExtensionResolver
	descriptorResolver protodesc.Resolver
}

// NewReflector constructs a highly configurable Reflector: it can serve a
// dynamic list of services, proxy reflection requests to other backends, or
// use an alternate source of extension information.
//
// To build a simpler Reflector that supports a static list of services using
// information from the package-global Protobuf registry, use
// NewStaticReflector.
func NewReflector(namer Namer, options ...Option) *Reflector {
	reflector := &Reflector{
		namer:              namer,
		extensionResolver:  protoregistry.GlobalTypes,
		descriptorResolver: globalFiles,
	}
	for _, option := range options {
		option.apply(reflector)
	}
	return reflector
}

// NewStaticReflector constructs a simple Reflector that supports a static list
// of services using information from the package-global Protobuf registry. For
// a more configurable Reflector, use NewReflector.
//
// The supplied strings should be fully-qualified Protobuf service names (for
// example, "acme.user.v1.UserService"). Generated Connect service files
// have this declared as a constant.
func NewStaticReflector(services ...string) *Reflector {
	namer := &staticNames{names: services}
	return NewReflector(namer)
}

// serverReflectionInfo implements the gRPC server reflection API.
func (r *Reflector) serverReflectionInfo(
	_ context.Context,
	stream *connect.BidiStream[
		reflectionv1.ServerReflectionRequest,
		reflectionv1.ServerReflectionResponse,
	],
) error {
	fileDescriptorsSent := &fileDescriptorNameSet{}
	for {
		request, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return err
		}
		// The server reflection API sends file descriptors as uncompressed
		// Protobuf-serialized bytes.
		response := &reflectionv1.ServerReflectionResponse{
			ValidHost:       request.Host,
			OriginalRequest: request,
		}
		switch messageRequest := request.MessageRequest.(type) {
		case *reflectionv1.ServerReflectionRequest_FileByFilename:
			data, err := r.getFileByFilename(messageRequest.FileByFilename, fileDescriptorsSent)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectionv1.FileDescriptorResponse{FileDescriptorProto: data},
				}
			}
		case *reflectionv1.ServerReflectionRequest_FileContainingSymbol:
			data, err := r.getFileContainingSymbol(
				messageRequest.FileContainingSymbol,
				fileDescriptorsSent,
			)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectionv1.FileDescriptorResponse{FileDescriptorProto: data},
				}
			}
		case *reflectionv1.ServerReflectionRequest_FileContainingExtension:
			msgFQN := messageRequest.FileContainingExtension.ContainingType
			extNumber := messageRequest.FileContainingExtension.ExtensionNumber
			data, err := r.getFileContainingExtension(msgFQN, extNumber, fileDescriptorsSent)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_FileDescriptorResponse{
					FileDescriptorResponse: &reflectionv1.FileDescriptorResponse{FileDescriptorProto: data},
				}
			}
		case *reflectionv1.ServerReflectionRequest_AllExtensionNumbersOfType:
			nums, err := r.getAllExtensionNumbersOfType(messageRequest.AllExtensionNumbersOfType)
			if err != nil {
				response.MessageResponse = newNotFoundResponse(err)
			} else {
				response.MessageResponse = &reflectionv1.ServerReflectionResponse_AllExtensionNumbersResponse{
					AllExtensionNumbersResponse: &reflectionv1.ExtensionNumberResponse{
						BaseTypeName:    messageRequest.AllExtensionNumbersOfType,
						ExtensionNumber: nums,
					},
				}
			}
		case *reflectionv1.ServerReflectionRequest_ListServices:
			services := r.namer.Names()
			serviceResponses := make([]*reflectionv1.ServiceResponse, len(services))
			for i, name := range services {
				serviceResponses[i] = &reflectionv1.ServiceResponse{Name: name}
			}
			response.MessageResponse = &reflectionv1.ServerReflectionResponse_ListServicesResponse{
				ListServicesResponse: &reflectionv1.ListServiceResponse{Service: serviceResponses},
			}
		default:
			return connect.NewError(connect.CodeInvalidArgument, fmt.Errorf(
				"invalid MessageRequest: %v",
				request.MessageRequest,
			))
		}
		if err := stream.Send(response); err != nil {
			return err
		}
	}
}

func (r *Reflector) getFileByFilename(fname string, sent *fileDescriptorNameSet) ([][]byte, error) {
	fd, err := r.descriptorResolver.FindFileByPath(fname)
	if err != nil {
		return nil, err
	}
	return fileDescriptorWithDependencies(fd, sent)
}

func (r *Reflector) getFileContainingSymbol(fqn string, sent *fileDescriptorNameSet) ([][]byte, error) {
	desc, err := r.descriptorResolver.FindDescriptorByName(protoreflect.FullName(fqn))
	if err != nil {
		return nil, err
	}
	fd := desc.ParentFile()
	if fd == nil {
		return nil, fmt.Errorf("no file for symbol %s", fqn)
	}
	return fileDescriptorWithDependencies(fd, sent)
}

func (r *Reflector) getFileContainingExtension(
	msgFQN string,
	extNumber int32,
	sent *fileDescriptorNameSet,
) ([][]byte, error) {
	extension, err := r.extensionResolver.FindExtensionByNumber(
		protoreflect.FullName(msgFQN),
		protoreflect.FieldNumber(extNumber),
	)
	if err != nil {
		return nil, err
	}
	fd := extension.TypeDescriptor().ParentFile()
	if fd == nil {
		return nil, fmt.Errorf("no file for extension %d of message %s", extNumber, msgFQN)
	}
	return fileDescriptorWithDependencies(fd, sent)
}

func (r *Reflector) getAllExtensionNumbersOfType(fqn string) ([]int32, error) {
	nums := []int32{}
	name := protoreflect.FullName(fqn)
	r.extensionResolver.RangeExtensionsByMessage(name, func(ext protoreflect.ExtensionType) bool {
		num := int32(ext.TypeDescriptor().Number())
		nums = append(nums, num)
		return true
	})
	if len(nums) == 0 {
		if _, err := r.descriptorResolver.FindDescriptorByName(name); err != nil {
			return nil, err
		}
	}
	sort.Slice(nums, func(i, j int) bool {
		return nums[i] < nums[j]
	})
	return nums, nil
}

// A Namer lists the fully-qualified Protobuf service names available for
// reflection (for example, "acme.user.v1.UserService"). Namers must be safe to
// call concurrently.
type Namer interface {
	Names() []string
}

// An Option configures a Reflector.
type Option interface {
	apply(*Reflector)
}

// WithExtensionResolver sets the resolver used to find Protobuf extensions. By
// default, Reflectors use protoregistry.GlobalTypes.
func WithExtensionResolver(resolver ExtensionResolver) Option {
	return &extensionResolverOption{resolver: resolver}
}

// WithDescriptorResolver sets the resolver used to find Protobuf type
// information (typically called a "descriptor"). By default, Reflectors use
// protoregistry.GlobalFiles.
func WithDescriptorResolver(resolver protodesc.Resolver) Option {
	return &descriptorResolverOption{resolver: resolver}
}

// An ExtensionResolver lets server reflection implementations query details
// about the registered Protobuf extensions. protoregistry.GlobalTypes
// implements ExtensionResolver.
//
// ExtensionResolvers must be safe to call concurrently.
type ExtensionResolver interface {
	protoregistry.ExtensionTypeResolver

	RangeExtensionsByMessage(protoreflect.FullName, func(protoreflect.ExtensionType) bool)
}

type fileDescriptorNameSet struct {
	names map[protoreflect.FullName]struct{}
}

func (s *fileDescriptorNameSet) Insert(fd protoreflect.FileDescriptor) {
	if s.names == nil {
		s.names = make(map[protoreflect.FullName]struct{}, 1)
	}
	s.names[fd.FullName()] = struct{}{}
}

func (s *fileDescriptorNameSet) Contains(fd protoreflect.FileDescriptor) bool {
	_, ok := s.names[fd.FullName()]
	return ok
}

func fileDescriptorWithDependencies(fd protoreflect.FileDescriptor, sent *fileDescriptorNameSet) ([][]byte, error) {
	results := make([][]byte, 0, 1)
	queue := []protoreflect.FileDescriptor{fd}
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		if len(results) == 0 || !sent.Contains(curr) { // always send root fd
			// Mark as sent immediately. If we hit an error marshaling below, there's
			// no point trying again later.
			sent.Insert(curr)
			encoded, err := proto.Marshal(protodesc.ToFileDescriptorProto(curr))
			if err != nil {
				return nil, err
			}
			results = append(results, encoded)
		}
		imports := curr.Imports()
		for i := 0; i < imports.Len(); i++ {
			queue = append(queue, imports.Get(i).FileDescriptor)
		}
	}
	return results, nil
}

func newNotFoundResponse(err error) *reflectionv1.ServerReflectionResponse_ErrorResponse {
	return &reflectionv1.ServerReflectionResponse_ErrorResponse{
		ErrorResponse: &reflectionv1.ErrorResponse{
			ErrorCode:    int32(connect.CodeNotFound),
			ErrorMessage: err.Error(),
		},
	}
}

func newHandler(reflector *Reflector, servicePath string, options []connect.HandlerOption) (string, http.Handler) {
	return servicePath, connect.NewBidiStreamHandler(
		servicePath+methodName,
		reflector.serverReflectionInfo,
		options...,
	)
}

type extensionResolverOption struct {
	resolver ExtensionResolver
}

func (o *extensionResolverOption) apply(reflector *Reflector) {
	reflector.extensionResolver = o.resolver
}

type descriptorResolverOption struct {
	resolver protodesc.Resolver
}

func (o *descriptorResolverOption) apply(reflector *Reflector) {
	reflector.descriptorResolver = o.resolver
}

type staticNames struct {
	names []string
}

func (n *staticNames) Names() []string {
	return n.names
}

// resolverHackForConnectext returns a resolver that can successfully resolve the descriptors
// for the gRPC health and reflection services. We need a work-around since this repo (and the
// connect-grpchealth-go repo) use "hacked" services that have a "connectext." package prefix.
// We don't use the "authoritative" packages for these descriptors because they depend on the
// gRPC runtime (ew!). We add a special prefix to the packages to avoid an init-time panic from
// duplicate registrations, in the event that the calling application _also_ imports the gRPC
// versions.
//
// This works by serving embedded descriptors (from "services.bin") for items not found in
// protoregistry.GlobalFiles. The only thing in the embedded descriptors are for the health
// and reflection services.
func resolverHackForConnectext(data []byte) protodesc.Resolver {
	var backupResolver protodesc.Resolver
	var fileSet descriptorpb.FileDescriptorSet
	if err := proto.Unmarshal(data, &fileSet); err != nil {
		backupResolver = &errResolver{err}
	} else if files, err := protodesc.NewFiles(&fileSet); err != nil {
		backupResolver = &errResolver{err}
	} else {
		backupResolver = files
	}

	return &combinedResolver{
		first:  protoregistry.GlobalFiles,
		second: backupResolver,
	}
}

type combinedResolver struct {
	first, second protodesc.Resolver
}

func (r *combinedResolver) FindFileByPath(s string) (protoreflect.FileDescriptor, error) {
	file, err := r.first.FindFileByPath(s)
	if errors.Is(err, protoregistry.NotFound) {
		file, err = r.second.FindFileByPath(s)
	}
	return file, err
}

func (r *combinedResolver) FindDescriptorByName(name protoreflect.FullName) (protoreflect.Descriptor, error) {
	desc, err := r.first.FindDescriptorByName(name)
	if errors.Is(err, protoregistry.NotFound) {
		desc, err = r.second.FindDescriptorByName(name)
	}
	return desc, err
}

type errResolver struct {
	err error
}

func (r *errResolver) FindFileByPath(_ string) (protoreflect.FileDescriptor, error) {
	return nil, r.err
}

func (r *errResolver) FindDescriptorByName(_ protoreflect.FullName) (protoreflect.Descriptor, error) {
	return nil, r.err
}
