// Copyright 2021-2024 The Connect Authors
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

package connect

import (
	"encoding/base64"
	"net/http"
)

// EncodeBinaryHeader base64-encodes the data. It always emits unpadded values.
//
// In the Connect, gRPC, and gRPC-Web protocols, binary headers must have keys
// ending in "-Bin".
func EncodeBinaryHeader(data []byte) string {
	// gRPC specification says that implementations should emit unpadded values.
	return base64.RawStdEncoding.EncodeToString(data)
}

// DecodeBinaryHeader base64-decodes the data. It can decode padded or unpadded
// values. Following usual HTTP semantics, multiple base64-encoded values may
// be joined with a comma. When receiving such comma-separated values, split
// them with [strings.Split] before calling DecodeBinaryHeader.
//
// Binary headers sent using the Connect, gRPC, and gRPC-Web protocols have
// keys ending in "-Bin".
func DecodeBinaryHeader(data string) ([]byte, error) {
	if len(data)%4 != 0 {
		// Data definitely isn't padded.
		return base64.RawStdEncoding.DecodeString(data)
	}
	// Either the data was padded, or padding wasn't necessary. In both cases,
	// the padding-aware decoder works.
	return base64.StdEncoding.DecodeString(data)
}

func mergeHeaders(into, from http.Header) {
	for key, vals := range from {
		if len(vals) == 0 {
			// For response trailers, net/http will pre-populate entries
			// with nil values based on the "Trailer" header. But if there
			// are no actual values for those keys, we skip them.
			continue
		}
		into[key] = append(into[key], vals...)
	}
}

// mergeMetdataHeaders merges the metadata headers from the "from" header into
// the "into" header. It skips over non metadata headers that should not be
// propagated from the server to the client.
func mergeMetadataHeaders(into, from http.Header) {
	for key, vals := range from {
		if len(vals) == 0 {
			// For response trailers, net/http will pre-populate entries
			// with nil values based on the "Trailer" header. But if there
			// are no actual values for those keys, we skip them.
			continue
		}
		switch http.CanonicalHeaderKey(key) {
		case headerContentType,
			headerContentLength,
			headerContentEncoding,
			headerHost,
			headerUserAgent,
			headerTrailer,
			headerDate:
			// HTTP headers.
		case connectUnaryHeaderAcceptCompression,
			connectUnaryTrailerPrefix,
			connectStreamingHeaderCompression,
			connectStreamingHeaderAcceptCompression,
			connectHeaderTimeout,
			connectHeaderProtocolVersion:
			// Connect headers.
		case grpcHeaderCompression,
			grpcHeaderAcceptCompression,
			grpcHeaderTimeout,
			grpcHeaderStatus,
			grpcHeaderMessage,
			grpcHeaderDetails:
			// gRPC headers.
		default:
			into[key] = append(into[key], vals...)
		}
	}
}

// getHeaderCanonical is a shortcut for Header.Get() which
// bypasses the CanonicalMIMEHeaderKey operation when we
// know the key is already in canonical form.
func getHeaderCanonical(h http.Header, key string) string {
	if h == nil {
		return ""
	}
	v := h[key]
	if len(v) == 0 {
		return ""
	}
	return v[0]
}

// setHeaderCanonical is a shortcut for Header.Set() which
// bypasses the CanonicalMIMEHeaderKey operation when we
// know the key is already in canonical form.
func setHeaderCanonical(h http.Header, key, value string) {
	h[key] = []string{value}
}

// delHeaderCanonical is a shortcut for Header.Del() which
// bypasses the CanonicalMIMEHeaderKey operation when we
// know the key is already in canonical form.
func delHeaderCanonical(h http.Header, key string) {
	delete(h, key)
}
