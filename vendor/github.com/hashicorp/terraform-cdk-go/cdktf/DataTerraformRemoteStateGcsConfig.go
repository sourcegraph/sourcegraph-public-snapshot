// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type DataTerraformRemoteStateGcsConfig struct {
	// Experimental.
	Defaults *map[string]interface{} `field:"optional" json:"defaults" yaml:"defaults"`
	// Experimental.
	Workspace *string `field:"optional" json:"workspace" yaml:"workspace"`
	// (Required) The name of the GCS bucket.
	//
	// This name must be globally unique.
	// Experimental.
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// (Optional) A temporary [OAuth 2.0 access token] obtained from the Google Authorization server, i.e. the Authorization: Bearer token used to authenticate HTTP requests to GCP APIs. This is an alternative to credentials. If both are specified, access_token will be used over the credentials field.
	// Experimental.
	AccessToken *string `field:"optional" json:"accessToken" yaml:"accessToken"`
	// (Optional) Local path to Google Cloud Platform account credentials in JSON format.
	//
	// If unset, Google Application Default Credentials are used.
	// The provided credentials must have Storage Object Admin role on the bucket.
	//
	// Warning: if using the Google Cloud Platform provider as well,
	// it will also pick up the GOOGLE_CREDENTIALS environment variable.
	// Experimental.
	Credentials *string `field:"optional" json:"credentials" yaml:"credentials"`
	// (Optional) A 32 byte base64 encoded 'customer supplied encryption key' used to encrypt all state.
	// Experimental.
	EncryptionKey *string `field:"optional" json:"encryptionKey" yaml:"encryptionKey"`
	// (Optional) The service account to impersonate for accessing the State Bucket.
	//
	// You must have roles/iam.serviceAccountTokenCreator role on that account for the impersonation to succeed.
	// If you are using a delegation chain, you can specify that using the impersonate_service_account_delegates field.
	// Alternatively, this can be specified using the GOOGLE_IMPERSONATE_SERVICE_ACCOUNT environment variable.
	// Experimental.
	ImpersonateServiceAccount *string `field:"optional" json:"impersonateServiceAccount" yaml:"impersonateServiceAccount"`
	// (Optional) The delegation chain for an impersonating a service account.
	// Experimental.
	ImpersonateServiceAccountDelegates *[]*string `field:"optional" json:"impersonateServiceAccountDelegates" yaml:"impersonateServiceAccountDelegates"`
	// (Optional) A Cloud KMS key ('customer-managed encryption key') used when reading and writing state files in the bucket.
	//
	// Format should be projects/{{project}}/locations/{{location}}/keyRings/{{keyRing}}/cryptoKeys/{{name}}.
	// For more information, including IAM requirements, see {@link https://cloud.google.com/storage/docs/encryption/customer-managed-keys Customer-managed Encryption Keys}.
	// Experimental.
	KmsEncryptionKey *string `field:"optional" json:"kmsEncryptionKey" yaml:"kmsEncryptionKey"`
	// (Optional) GCS prefix inside the bucket.
	//
	// Named states for workspaces are stored in an object called <prefix>/<name>.tfstate.
	// Experimental.
	Prefix *string `field:"optional" json:"prefix" yaml:"prefix"`
	// (Optional) A URL containing three parts: the protocol, the DNS name pointing to a Private Service Connect endpoint, and the path for the Cloud Storage API (/storage/v1/b).
	//
	// {@link https://developer.hashicorp.com/terraform/language/settings/backends/gcs#storage_custom_endpoint See here for more details}
	// Experimental.
	StoreageCustomEndpoint *string `field:"optional" json:"storeageCustomEndpoint" yaml:"storeageCustomEndpoint"`
}

