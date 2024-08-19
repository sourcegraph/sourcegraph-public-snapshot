// Copyright 2024 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go_gapic. DO NOT EDIT.

// Package secretmanager is an auto-generated package for the
// Secret Manager API.
//
// Stores sensitive data such as API keys, passwords, and certificates.
// Provides convenience while improving security.
//
// # General documentation
//
// For information that is relevant for all client libraries please reference
// https://pkg.go.dev/cloud.google.com/go#pkg-overview. Some information on this
// page includes:
//
//   - [Authentication and Authorization]
//   - [Timeouts and Cancellation]
//   - [Testing against Client Libraries]
//   - [Debugging Client Libraries]
//   - [Inspecting errors]
//
// # Example usage
//
// To get started with this package, create a client.
//
//	ctx := context.Background()
//	// This snippet has been automatically generated and should be regarded as a code template only.
//	// It will require modifications to work:
//	// - It may require correct/in-range values for request initialization.
//	// - It may require specifying regional endpoints when creating the service client as shown in:
//	//   https://pkg.go.dev/cloud.google.com/go#hdr-Client_Options
//	c, err := secretmanager.NewClient(ctx)
//	if err != nil {
//		// TODO: Handle error.
//	}
//	defer c.Close()
//
// The client will use your default application credentials. Clients should be reused instead of created as needed.
// The methods of Client are safe for concurrent use by multiple goroutines.
// The returned client must be Closed when it is done being used.
//
// # Using the Client
//
// The following is an example of making an API call with the newly created client.
//
//	ctx := context.Background()
//	// This snippet has been automatically generated and should be regarded as a code template only.
//	// It will require modifications to work:
//	// - It may require correct/in-range values for request initialization.
//	// - It may require specifying regional endpoints when creating the service client as shown in:
//	//   https://pkg.go.dev/cloud.google.com/go#hdr-Client_Options
//	c, err := secretmanager.NewClient(ctx)
//	if err != nil {
//		// TODO: Handle error.
//	}
//	defer c.Close()
//
//	req := &secretmanagerpb.AccessSecretVersionRequest{
//		// TODO: Fill request struct fields.
//		// See https://pkg.go.dev/cloud.google.com/go/secretmanager/apiv1/secretmanagerpb#AccessSecretVersionRequest.
//	}
//	resp, err := c.AccessSecretVersion(ctx, req)
//	if err != nil {
//		// TODO: Handle error.
//	}
//	// TODO: Use resp.
//	_ = resp
//
// # Use of Context
//
// The ctx passed to NewClient is used for authentication requests and
// for creating the underlying connection, but is not used for subsequent calls.
// Individual methods on the client use the ctx given to them.
//
// To close the open connection, use the Close() method.
//
// [Authentication and Authorization]: https://pkg.go.dev/cloud.google.com/go#hdr-Authentication_and_Authorization
// [Timeouts and Cancellation]: https://pkg.go.dev/cloud.google.com/go#hdr-Timeouts_and_Cancellation
// [Testing against Client Libraries]: https://pkg.go.dev/cloud.google.com/go#hdr-Testing
// [Debugging Client Libraries]: https://pkg.go.dev/cloud.google.com/go#hdr-Debugging
// [Inspecting errors]: https://pkg.go.dev/cloud.google.com/go#hdr-Inspecting_errors
package secretmanager // import "cloud.google.com/go/secretmanager/apiv1"

import (
	"context"

	"google.golang.org/api/option"
)

// For more information on implementing a client constructor hook, see
// https://github.com/googleapis/google-cloud-go/wiki/Customizing-constructors.
type clientHookParams struct{}
type clientHook func(context.Context, clientHookParams) ([]option.ClientOption, error)

var versionClient string

func getVersionClient() string {
	if versionClient == "" {
		return "UNKNOWN"
	}
	return versionClient
}

// DefaultAuthScopes reports the default set of authentication scopes to use with this package.
func DefaultAuthScopes() []string {
	return []string{
		"https://www.googleapis.com/auth/cloud-platform",
	}
}
