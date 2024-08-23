// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Stores the state using a simple REST client.
//
// State will be fetched via GET, updated via POST, and purged with DELETE.
// The method used for updating is configurable.
//
// This backend optionally supports state locking.
// When locking support is enabled it will use LOCK and UNLOCK requests providing the lock info in the body.
// The endpoint should return a 423: Locked or 409: Conflict with the holding lock info when
// it's already taken, 200: OK for success. Any other status will be considered an error.
// The ID of the holding lock info will be added as a query parameter to state updates requests.
//
// Read more about this backend in the Terraform docs:
// https://developer.hashicorp.com/terraform/language/settings/backends/http
// Experimental.
type HttpBackendConfig struct {
	// (Required) The address of the REST endpoint.
	// Experimental.
	Address *string `field:"required" json:"address" yaml:"address"`
	// (Optional) A PEM-encoded CA certificate chain used by the client to verify server certificates during TLS authentication.
	// Experimental.
	ClientCaCertificatePem *string `field:"optional" json:"clientCaCertificatePem" yaml:"clientCaCertificatePem"`
	// (Optional) A PEM-encoded certificate used by the server to verify the client during mutual TLS (mTLS) authentication.
	// Experimental.
	ClientCertificatePem *string `field:"optional" json:"clientCertificatePem" yaml:"clientCertificatePem"`
	// (Optional) A PEM-encoded private key, required if client_certificate_pem is specified.
	// Experimental.
	ClientPrivateKeyPem *string `field:"optional" json:"clientPrivateKeyPem" yaml:"clientPrivateKeyPem"`
	// (Optional) The address of the lock REST endpoint.
	//
	// Defaults to disabled.
	// Experimental.
	LockAddress *string `field:"optional" json:"lockAddress" yaml:"lockAddress"`
	// (Optional) The HTTP method to use when locking.
	//
	// Defaults to LOCK.
	// Experimental.
	LockMethod *string `field:"optional" json:"lockMethod" yaml:"lockMethod"`
	// (Optional) The password for HTTP basic authentication.
	// Experimental.
	Password *string `field:"optional" json:"password" yaml:"password"`
	// (Optional) The number of HTTP request retries.
	//
	// Defaults to 2.
	// Experimental.
	RetryMax *float64 `field:"optional" json:"retryMax" yaml:"retryMax"`
	// (Optional) The maximum time in seconds to wait between HTTP request attempts.
	//
	// Defaults to 30.
	// Experimental.
	RetryWaitMax *float64 `field:"optional" json:"retryWaitMax" yaml:"retryWaitMax"`
	// (Optional) The minimum time in seconds to wait between HTTP request attempts.
	//
	// Defaults to 1.
	// Experimental.
	RetryWaitMin *float64 `field:"optional" json:"retryWaitMin" yaml:"retryWaitMin"`
	// (Optional) Whether to skip TLS verification.
	//
	// Defaults to false.
	// Experimental.
	SkipCertVerification *bool `field:"optional" json:"skipCertVerification" yaml:"skipCertVerification"`
	// (Optional) The address of the unlock REST endpoint.
	//
	// Defaults to disabled.
	// Experimental.
	UnlockAddress *string `field:"optional" json:"unlockAddress" yaml:"unlockAddress"`
	// (Optional) The HTTP method to use when unlocking.
	//
	// Defaults to UNLOCK.
	// Experimental.
	UnlockMethod *string `field:"optional" json:"unlockMethod" yaml:"unlockMethod"`
	// (Optional) HTTP method to use when updating state.
	//
	// Defaults to POST.
	// Experimental.
	UpdateMethod *string `field:"optional" json:"updateMethod" yaml:"updateMethod"`
	// (Optional) The username for HTTP basic authentication.
	// Experimental.
	Username *string `field:"optional" json:"username" yaml:"username"`
}

