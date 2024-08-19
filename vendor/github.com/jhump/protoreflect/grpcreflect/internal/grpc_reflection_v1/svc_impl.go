package grpc_reflection_v1

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

func Register(svr reflection.GRPCServer) {
	reflection.Register(registrarInterceptor{svr})
}

type registrarInterceptor struct {
	svr reflection.GRPCServer
}

func (r registrarInterceptor) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	r.svr.RegisterService(&ServerReflection_ServiceDesc, reflectImpl{svr: impl.(grpc_reflection_v1alpha.ServerReflectionServer)})
}

func (r registrarInterceptor) GetServiceInfo() map[string]grpc.ServiceInfo {
	// HACK: We're using generated code for a proto file where we hacked the proto package
	// to avoid init-time issues (for a future where the grpc module also provides the same
	// protos/types). But we've rewritten the service names in the generated code, so that
	// we expose the expected service (e.g. w/out the hacked package name). That will lead
	// to issues trying to load/resolve descriptors for the hacked service. So we remove
	// it from the service info.
	info := r.svr.GetServiceInfo()
	delete(info, "grpc.reflection.v1.ServerReflection")
	return info
}

type reflectImpl struct {
	svr grpc_reflection_v1alpha.ServerReflectionServer
	UnimplementedServerReflectionServer
}

func (r reflectImpl) ServerReflectionInfo(stream ServerReflection_ServerReflectionInfoServer) error {
	return r.svr.ServerReflectionInfo(streamImpl{stream})
}

type streamImpl struct {
	ServerReflection_ServerReflectionInfoServer
}

func (s streamImpl) Send(response *grpc_reflection_v1alpha.ServerReflectionResponse) error {
	return s.ServerReflection_ServerReflectionInfoServer.Send(ToV1Response(response))
}

func (s streamImpl) Recv() (*grpc_reflection_v1alpha.ServerReflectionRequest, error) {
	resp, err := s.ServerReflection_ServerReflectionInfoServer.Recv()
	if err != nil {
		return nil, err
	}
	return ToV1AlphaRequest(resp), nil
}

func ToV1Request(v1alpha *grpc_reflection_v1alpha.ServerReflectionRequest) *ServerReflectionRequest {
	var v1 ServerReflectionRequest
	v1.Host = v1alpha.Host
	switch mr := v1alpha.MessageRequest.(type) {
	case *grpc_reflection_v1alpha.ServerReflectionRequest_FileByFilename:
		v1.MessageRequest = &ServerReflectionRequest_FileByFilename{
			FileByFilename: mr.FileByFilename,
		}
	case *grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol:
		v1.MessageRequest = &ServerReflectionRequest_FileContainingSymbol{
			FileContainingSymbol: mr.FileContainingSymbol,
		}
	case *grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingExtension:
		if mr.FileContainingExtension != nil {
			v1.MessageRequest = &ServerReflectionRequest_FileContainingExtension{
				FileContainingExtension: &ExtensionRequest{
					ContainingType:  mr.FileContainingExtension.GetContainingType(),
					ExtensionNumber: mr.FileContainingExtension.GetExtensionNumber(),
				},
			}
		}
	case *grpc_reflection_v1alpha.ServerReflectionRequest_AllExtensionNumbersOfType:
		v1.MessageRequest = &ServerReflectionRequest_AllExtensionNumbersOfType{
			AllExtensionNumbersOfType: mr.AllExtensionNumbersOfType,
		}
	case *grpc_reflection_v1alpha.ServerReflectionRequest_ListServices:
		v1.MessageRequest = &ServerReflectionRequest_ListServices{
			ListServices: mr.ListServices,
		}
	default:
		// no value set
	}
	return &v1
}

func ToV1AlphaRequest(v1 *ServerReflectionRequest) *grpc_reflection_v1alpha.ServerReflectionRequest {
	var v1alpha grpc_reflection_v1alpha.ServerReflectionRequest
	v1alpha.Host = v1.Host
	switch mr := v1.MessageRequest.(type) {
	case *ServerReflectionRequest_FileByFilename:
		if mr != nil {
			v1alpha.MessageRequest = &grpc_reflection_v1alpha.ServerReflectionRequest_FileByFilename{
				FileByFilename: mr.FileByFilename,
			}
		}
	case *ServerReflectionRequest_FileContainingSymbol:
		if mr != nil {
			v1alpha.MessageRequest = &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingSymbol{
				FileContainingSymbol: mr.FileContainingSymbol,
			}
		}
	case *ServerReflectionRequest_FileContainingExtension:
		if mr != nil {
			v1alpha.MessageRequest = &grpc_reflection_v1alpha.ServerReflectionRequest_FileContainingExtension{
				FileContainingExtension: &grpc_reflection_v1alpha.ExtensionRequest{
					ContainingType:  mr.FileContainingExtension.GetContainingType(),
					ExtensionNumber: mr.FileContainingExtension.GetExtensionNumber(),
				},
			}
		}
	case *ServerReflectionRequest_AllExtensionNumbersOfType:
		if mr != nil {
			v1alpha.MessageRequest = &grpc_reflection_v1alpha.ServerReflectionRequest_AllExtensionNumbersOfType{
				AllExtensionNumbersOfType: mr.AllExtensionNumbersOfType,
			}
		}
	case *ServerReflectionRequest_ListServices:
		if mr != nil {
			v1alpha.MessageRequest = &grpc_reflection_v1alpha.ServerReflectionRequest_ListServices{
				ListServices: mr.ListServices,
			}
		}
	default:
		// no value set
	}
	return &v1alpha
}

func ToV1Response(v1alpha *grpc_reflection_v1alpha.ServerReflectionResponse) *ServerReflectionResponse {
	var v1 ServerReflectionResponse
	v1.ValidHost = v1alpha.ValidHost
	if v1alpha.OriginalRequest != nil {
		v1.OriginalRequest = ToV1Request(v1alpha.OriginalRequest)
	}
	switch mr := v1alpha.MessageResponse.(type) {
	case *grpc_reflection_v1alpha.ServerReflectionResponse_FileDescriptorResponse:
		if mr != nil {
			v1.MessageResponse = &ServerReflectionResponse_FileDescriptorResponse{
				FileDescriptorResponse: &FileDescriptorResponse{
					FileDescriptorProto: mr.FileDescriptorResponse.GetFileDescriptorProto(),
				},
			}
		}
	case *grpc_reflection_v1alpha.ServerReflectionResponse_AllExtensionNumbersResponse:
		if mr != nil {
			v1.MessageResponse = &ServerReflectionResponse_AllExtensionNumbersResponse{
				AllExtensionNumbersResponse: &ExtensionNumberResponse{
					BaseTypeName:    mr.AllExtensionNumbersResponse.GetBaseTypeName(),
					ExtensionNumber: mr.AllExtensionNumbersResponse.GetExtensionNumber(),
				},
			}
		}
	case *grpc_reflection_v1alpha.ServerReflectionResponse_ListServicesResponse:
		if mr != nil {
			svcs := make([]*ServiceResponse, len(mr.ListServicesResponse.GetService()))
			for i, svc := range mr.ListServicesResponse.GetService() {
				svcs[i] = &ServiceResponse{
					Name: svc.GetName(),
				}
			}
			v1.MessageResponse = &ServerReflectionResponse_ListServicesResponse{
				ListServicesResponse: &ListServiceResponse{
					Service: svcs,
				},
			}
		}
	case *grpc_reflection_v1alpha.ServerReflectionResponse_ErrorResponse:
		if mr != nil {
			v1.MessageResponse = &ServerReflectionResponse_ErrorResponse{
				ErrorResponse: &ErrorResponse{
					ErrorCode:    mr.ErrorResponse.GetErrorCode(),
					ErrorMessage: mr.ErrorResponse.GetErrorMessage(),
				},
			}
		}
	default:
		// no value set
	}
	return &v1
}

func ToV1AlphaResponse(v1 *ServerReflectionResponse) *grpc_reflection_v1alpha.ServerReflectionResponse {
	var v1alpha grpc_reflection_v1alpha.ServerReflectionResponse
	v1alpha.ValidHost = v1.ValidHost
	if v1.OriginalRequest != nil {
		v1alpha.OriginalRequest = ToV1AlphaRequest(v1.OriginalRequest)
	}
	switch mr := v1.MessageResponse.(type) {
	case *ServerReflectionResponse_FileDescriptorResponse:
		if mr != nil {
			v1alpha.MessageResponse = &grpc_reflection_v1alpha.ServerReflectionResponse_FileDescriptorResponse{
				FileDescriptorResponse: &grpc_reflection_v1alpha.FileDescriptorResponse{
					FileDescriptorProto: mr.FileDescriptorResponse.GetFileDescriptorProto(),
				},
			}
		}
	case *ServerReflectionResponse_AllExtensionNumbersResponse:
		if mr != nil {
			v1alpha.MessageResponse = &grpc_reflection_v1alpha.ServerReflectionResponse_AllExtensionNumbersResponse{
				AllExtensionNumbersResponse: &grpc_reflection_v1alpha.ExtensionNumberResponse{
					BaseTypeName:    mr.AllExtensionNumbersResponse.GetBaseTypeName(),
					ExtensionNumber: mr.AllExtensionNumbersResponse.GetExtensionNumber(),
				},
			}
		}
	case *ServerReflectionResponse_ListServicesResponse:
		if mr != nil {
			svcs := make([]*grpc_reflection_v1alpha.ServiceResponse, len(mr.ListServicesResponse.GetService()))
			for i, svc := range mr.ListServicesResponse.GetService() {
				svcs[i] = &grpc_reflection_v1alpha.ServiceResponse{
					Name: svc.GetName(),
				}
			}
			v1alpha.MessageResponse = &grpc_reflection_v1alpha.ServerReflectionResponse_ListServicesResponse{
				ListServicesResponse: &grpc_reflection_v1alpha.ListServiceResponse{
					Service: svcs,
				},
			}
		}
	case *ServerReflectionResponse_ErrorResponse:
		if mr != nil {
			v1alpha.MessageResponse = &grpc_reflection_v1alpha.ServerReflectionResponse_ErrorResponse{
				ErrorResponse: &grpc_reflection_v1alpha.ErrorResponse{
					ErrorCode:    mr.ErrorResponse.GetErrorCode(),
					ErrorMessage: mr.ErrorResponse.GetErrorMessage(),
				},
			}
		}
	default:
		// no value set
	}
	return &v1alpha
}
