package errcode

import (
	"net/http"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph/legacyerr"
)

// codeToHTTP maps error codes to HTTP status codes.
func codeToHTTP(code legacyerr.Code) int {
	if h, present := codeToHTTPMap[code]; present {
		return h
	}
	return http.StatusInternalServerError
}

// HTTPToCode returns the most appropriate gRPC/legacy error code for the
// HTTP status. For example, HTTP 404 is mapped to codes.NotFound.
func HTTPToCode(statusCode int) legacyerr.Code {
	for g, h := range codeToHTTPMap {
		if h == statusCode {
			return g
		}
	}
	return legacyerr.Unknown
}

// codeToHTTPMap is a 1-to-1 mapping of gRPC/legacy error codes to HTTP
// status codes. NOTE: If you change this so it's not 1-to-1, you will
// need to update the way that HTTP codes are mapped to gRPC/legacy error
// codes in code that uses this mapping, to ensure determinism.
//
// Callers should use the funcs grpcToHTTP or httpToCode to map error
// values. Those funcs properly handle the default and zero cases.
var codeToHTTPMap = map[legacyerr.Code]int{
	legacyerr.Unknown:            http.StatusInternalServerError,
	legacyerr.InvalidArgument:    http.StatusBadRequest,
	legacyerr.NotFound:           http.StatusNotFound,
	legacyerr.AlreadyExists:      http.StatusConflict,
	legacyerr.PermissionDenied:   http.StatusForbidden,
	legacyerr.Unauthenticated:    http.StatusUnauthorized,
	legacyerr.FailedPrecondition: http.StatusPreconditionFailed,
	legacyerr.Unimplemented:      http.StatusNotImplemented,
	legacyerr.ResourceExhausted:  http.StatusTooManyRequests,
}
