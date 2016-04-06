package errcode

import (
	"net/http"

	"google.golang.org/grpc/codes"
)

// grpcToHTTP maps gRPC codes to HTTP status codes.
func grpcToHTTP(code codes.Code) int {
	if h, present := grpcToHTTPMap[code]; present {
		return h
	}
	return http.StatusInternalServerError
}

// HTTPToGRPC returns the most appropriate gRPC error code for the
// HTTP status. For example, HTTP 404 is mapped to codes.NotFound.
func HTTPToGRPC(statusCode int) codes.Code {
	if statusCode < 400 {
		return codes.OK
	}
	for g, h := range grpcToHTTPMap {
		if h == statusCode {
			return g
		}
	}
	return codes.Unknown
}

// grpcToHTTPMap is a 1-to-1 mapping of gRPC error codes to HTTP
// status codes. NOTE: If you change this so it's not 1-to-1, you will
// need to update the way that HTTP codes are mapped to gRPC error
// codes in code that uses this mapping, to ensure determinism.
//
// Callers should use the funcs grpcToHTTP or httpToGRPC to map error
// values. Those funcs properly handle the default and zero cases.
var grpcToHTTPMap = map[codes.Code]int{
	codes.OK:                 http.StatusOK,
	codes.Unknown:            http.StatusInternalServerError,
	codes.InvalidArgument:    http.StatusBadRequest,
	codes.NotFound:           http.StatusNotFound,
	codes.AlreadyExists:      http.StatusConflict,
	codes.PermissionDenied:   http.StatusForbidden,
	codes.Unauthenticated:    http.StatusUnauthorized,
	codes.FailedPrecondition: http.StatusPreconditionFailed,
	codes.OutOfRange:         http.StatusRequestedRangeNotSatisfiable,
	codes.Unimplemented:      http.StatusNotImplemented,
	codes.ResourceExhausted:  http.StatusTooManyRequests,
}
