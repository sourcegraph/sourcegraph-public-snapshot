// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type DataTerraformRemoteStateS3Config struct {
	// Experimental.
	Defaults *map[string]interface{} `field:"optional" json:"defaults" yaml:"defaults"`
	// Experimental.
	Workspace *string `field:"optional" json:"workspace" yaml:"workspace"`
	// Name of the S3 Bucket.
	// Experimental.
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// Path to the state file inside the S3 Bucket.
	//
	// When using a non-default workspace, the state path will be /workspace_key_prefix/workspace_name/key.
	// Experimental.
	Key *string `field:"required" json:"key" yaml:"key"`
	// (Optional) AWS access key.
	//
	// If configured, must also configure secret_key.
	// This can also be sourced from
	// the AWS_ACCESS_KEY_ID environment variable,
	// AWS shared credentials file (e.g. ~/.aws/credentials),
	// or AWS shared configuration file (e.g. ~/.aws/config).
	// Experimental.
	AccessKey *string `field:"optional" json:"accessKey" yaml:"accessKey"`
	// (Optional) Canned ACL to be applied to the state file.
	// Experimental.
	Acl *string `field:"optional" json:"acl" yaml:"acl"`
	// (Optional) List of allowed AWS account IDs to prevent potential destruction of a live environment.
	//
	// Conflicts with forbidden_account_ids.
	// Experimental.
	AllowedAccountIds *string `field:"optional" json:"allowedAccountIds" yaml:"allowedAccountIds"`
	// Assuming an IAM Role can be configured in two ways.
	//
	// The preferred way is to use the argument assume_role, the other, which is deprecated, is with arguments at the top level.
	// Experimental.
	AssumeRole *S3BackendAssumeRoleConfig `field:"optional" json:"assumeRole" yaml:"assumeRole"`
	// (Optional) IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.
	// Deprecated: Use assumeRole.policy instead.
	AssumeRolePolicy *string `field:"optional" json:"assumeRolePolicy" yaml:"assumeRolePolicy"`
	// (Optional) Set of Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.
	// Deprecated: Use assumeRole.policyArns instead.
	AssumeRolePolicyArns *[]*string `field:"optional" json:"assumeRolePolicyArns" yaml:"assumeRolePolicyArns"`
	// (Optional) Map of assume role session tags.
	// Deprecated: Use assumeRole.tags instead.
	AssumeRoleTags *map[string]*string `field:"optional" json:"assumeRoleTags" yaml:"assumeRoleTags"`
	// (Optional) Set of assume role session tag keys to pass to any subsequent sessions.
	// Deprecated: Use assumeRole.transitiveTagKeys instead.
	AssumeRoleTransitiveTagKeys *[]*string `field:"optional" json:"assumeRoleTransitiveTagKeys" yaml:"assumeRoleTransitiveTagKeys"`
	// Assume Role With Web Identity Configuration.
	// Experimental.
	AssumeRoleWithWebIdentity *S3BackendAssumeRoleWithWebIdentityConfig `field:"optional" json:"assumeRoleWithWebIdentity" yaml:"assumeRoleWithWebIdentity"`
	// (Optional) File containing custom root and intermediate certificates.
	//
	// Can also be set using the AWS_CA_BUNDLE environment variable.
	// Setting ca_bundle in the shared config file is not supported.
	// Experimental.
	CustomCaBundle *string `field:"optional" json:"customCaBundle" yaml:"customCaBundle"`
	// (Optional) Custom endpoint for the AWS DynamoDB API.
	//
	// This can also be sourced from the AWS_DYNAMODB_ENDPOINT environment variable.
	// Deprecated: Use endpoints.dynamodb instead
	DynamodbEndpoint *string `field:"optional" json:"dynamodbEndpoint" yaml:"dynamodbEndpoint"`
	// (Optional) Name of DynamoDB Table to use for state locking and consistency.
	//
	// The table must have a partition key named LockID with type of String.
	// If not configured, state locking will be disabled.
	// Experimental.
	DynamodbTable *string `field:"optional" json:"dynamodbTable" yaml:"dynamodbTable"`
	// Optional) Custom endpoint URL for the EC2 Instance Metadata Service (IMDS) API.
	//
	// Can also be set with the AWS_EC2_METADATA_SERVICE_ENDPOINT environment variable.
	// Experimental.
	Ec2MetadataServiceEndpoint *string `field:"optional" json:"ec2MetadataServiceEndpoint" yaml:"ec2MetadataServiceEndpoint"`
	// (Optional) Mode to use in communicating with the metadata service.
	//
	// Valid values are IPv4 and IPv6.
	// Can also be set with the AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE environment variable.
	// Experimental.
	Ec2MetadataServiceEndpointMode *string `field:"optional" json:"ec2MetadataServiceEndpointMode" yaml:"ec2MetadataServiceEndpointMode"`
	// (Optional) Enable server side encryption of the state file.
	// Experimental.
	Encrypt *bool `field:"optional" json:"encrypt" yaml:"encrypt"`
	// (Optional) Custom endpoint for the AWS S3 API.
	//
	// This can also be sourced from the AWS_S3_ENDPOINT environment variable.
	// Deprecated: Use endpoints.s3 instead
	Endpoint *string `field:"optional" json:"endpoint" yaml:"endpoint"`
	// (Optional) The endpoint configuration block.
	// Experimental.
	Endpoints *S3BackendEndpointConfig `field:"optional" json:"endpoints" yaml:"endpoints"`
	// (Optional) External identifier to use when assuming the role.
	// Deprecated: Use assume_role.external_id instead.
	ExternalId *string `field:"optional" json:"externalId" yaml:"externalId"`
	// (Optional) List of forbidden AWS account IDs to prevent potential destruction of a live environment.
	//
	// Conflicts with allowed_account_ids.
	// Experimental.
	ForbiddenAccountIds *string `field:"optional" json:"forbiddenAccountIds" yaml:"forbiddenAccountIds"`
	// (Optional) Enable path-style S3 URLs (https://<HOST>/<BUCKET> instead of https://<BUCKET>.<HOST>).
	// Deprecated: Use usePathStyle instead.
	ForcePathStyle *bool `field:"optional" json:"forcePathStyle" yaml:"forcePathStyle"`
	// (Optional) URL of a proxy to use for HTTP requests when accessing the AWS API.
	//
	// Can also be set using the HTTP_PROXY or http_proxy environment variables.
	// Experimental.
	HttpProxy *string `field:"optional" json:"httpProxy" yaml:"httpProxy"`
	// (Optional) URL of a proxy to use for HTTPS requests when accessing the AWS API.
	//
	// Can also be set using the HTTPS_PROXY or https_proxy environment variables.
	// Experimental.
	HttpsProxy *string `field:"optional" json:"httpsProxy" yaml:"httpsProxy"`
	// (Optional) Custom endpoint for the AWS Identity and Access Management (IAM) API.
	//
	// This can also be sourced from the AWS_IAM_ENDPOINT environment variable.
	// Deprecated: Use endpoints.iam instead
	IamEndpoint *string `field:"optional" json:"iamEndpoint" yaml:"iamEndpoint"`
	// Optional) Whether to explicitly allow the backend to perform "insecure" SSL requests.
	//
	// If omitted, the default value is false.
	// Experimental.
	Insecure *bool `field:"optional" json:"insecure" yaml:"insecure"`
	// (Optional) Amazon Resource Name (ARN) of a Key Management Service (KMS) Key to use for encrypting the state.
	//
	// Note that if this value is specified,
	// Terraform will need kms:Encrypt, kms:Decrypt and kms:GenerateDataKey permissions on this KMS key.
	// Experimental.
	KmsKeyId *string `field:"optional" json:"kmsKeyId" yaml:"kmsKeyId"`
	// (Optional) The maximum number of times an AWS API request is retried on retryable failure.
	//
	// Defaults to 5.
	// Experimental.
	MaxRetries *float64 `field:"optional" json:"maxRetries" yaml:"maxRetries"`
	// (Optional) Comma-separated list of hosts that should not use HTTP or HTTPS proxies.
	//
	// Each value can be one of:
	// - A domain name
	// - An IP address
	// - A CIDR address
	// - An asterisk (*), to indicate that no proxying should be performed Domain name and IP address values can also include a port number.
	// Can also be set using the NO_PROXY or no_proxy environment variables.
	// Experimental.
	NoProxy *string `field:"optional" json:"noProxy" yaml:"noProxy"`
	// (Optional) Name of AWS profile in AWS shared credentials file (e.g. ~/.aws/credentials) or AWS shared configuration file (e.g. ~/.aws/config) to use for credentials and/or configuration. This can also be sourced from the AWS_PROFILE environment variable.
	// Experimental.
	Profile *string `field:"optional" json:"profile" yaml:"profile"`
	// AWS Region of the S3 Bucket and DynamoDB Table (if used).
	//
	// This can also
	// be sourced from the AWS_DEFAULT_REGION and AWS_REGION environment
	// variables.
	// Experimental.
	Region *string `field:"optional" json:"region" yaml:"region"`
	// (Optional) Specifies how retries are attempted.
	//
	// Valid values are standard and adaptive.
	// Can also be configured using the AWS_RETRY_MODE environment variable or the shared config file parameter retry_mode.
	// Experimental.
	RetryMode *string `field:"optional" json:"retryMode" yaml:"retryMode"`
	// (Optional) Amazon Resource Name (ARN) of the IAM Role to assume.
	// Deprecated: Use assumeRole.roleArn instead.
	RoleArn *string `field:"optional" json:"roleArn" yaml:"roleArn"`
	// (Optional) AWS secret access key.
	//
	// If configured, must also configure access_key.
	// This can also be sourced from
	// the AWS_SECRET_ACCESS_KEY environment variable,
	// AWS shared credentials file (e.g. ~/.aws/credentials),
	// or AWS shared configuration file (e.g. ~/.aws/config)
	// Experimental.
	SecretKey *string `field:"optional" json:"secretKey" yaml:"secretKey"`
	// (Optional) Session name to use when assuming the role.
	// Deprecated: Use assumeRole.sessionName instead.
	SessionName *string `field:"optional" json:"sessionName" yaml:"sessionName"`
	// (Optional) List of paths to AWS shared configuration files.
	//
	// Defaults to ~/.aws/config.
	// Experimental.
	SharedConfigFiles *[]*string `field:"optional" json:"sharedConfigFiles" yaml:"sharedConfigFiles"`
	// (Optional) Path to the AWS shared credentials file.
	//
	// Defaults to ~/.aws/credentials.
	// Experimental.
	SharedCredentialsFile *string `field:"optional" json:"sharedCredentialsFile" yaml:"sharedCredentialsFile"`
	// (Optional) List of paths to AWS shared credentials files.
	//
	// Defaults to ~/.aws/credentials.
	// Experimental.
	SharedCredentialsFiles *[]*string `field:"optional" json:"sharedCredentialsFiles" yaml:"sharedCredentialsFiles"`
	// (Optional) Skip credentials validation via the STS API.
	// Experimental.
	SkipCredentialsValidation *bool `field:"optional" json:"skipCredentialsValidation" yaml:"skipCredentialsValidation"`
	// (Optional) Skip usage of EC2 Metadata API.
	// Experimental.
	SkipMetadataApiCheck *bool `field:"optional" json:"skipMetadataApiCheck" yaml:"skipMetadataApiCheck"`
	// (Optional) Skip validation of provided region name.
	// Experimental.
	SkipRegionValidation *bool `field:"optional" json:"skipRegionValidation" yaml:"skipRegionValidation"`
	// (Optional) Whether to skip requesting the account ID.
	//
	// Useful for AWS API implementations that do not have the IAM, STS API, or metadata API.
	// Experimental.
	SkipRequestingAccountId *bool `field:"optional" json:"skipRequestingAccountId" yaml:"skipRequestingAccountId"`
	// (Optional) Do not include checksum when uploading S3 Objects.
	//
	// Useful for some S3-Compatible APIs.
	// Experimental.
	SkipS3Checksum *bool `field:"optional" json:"skipS3Checksum" yaml:"skipS3Checksum"`
	// (Optional) The key to use for encrypting state with Server-Side Encryption with Customer-Provided Keys (SSE-C).
	//
	// This is the base64-encoded value of the key, which must decode to 256 bits.
	// This can also be sourced from the AWS_SSE_CUSTOMER_KEY environment variable,
	// which is recommended due to the sensitivity of the value.
	// Setting it inside a terraform file will cause it to be persisted to disk in terraform.tfstate.
	// Experimental.
	SseCustomerKey *string `field:"optional" json:"sseCustomerKey" yaml:"sseCustomerKey"`
	// (Optional) Custom endpoint for the AWS Security Token Service (STS) API.
	//
	// This can also be sourced from the AWS_STS_ENDPOINT environment variable.
	// Deprecated: Use endpoints.sts instead
	StsEndpoint *string `field:"optional" json:"stsEndpoint" yaml:"stsEndpoint"`
	// (Optional) AWS region for STS.
	//
	// If unset, AWS will use the same region for STS as other non-STS operations.
	// Experimental.
	StsRegion *string `field:"optional" json:"stsRegion" yaml:"stsRegion"`
	// (Optional) Multi-Factor Authentication (MFA) token.
	//
	// This can also be sourced from the AWS_SESSION_TOKEN environment variable.
	// Experimental.
	Token *string `field:"optional" json:"token" yaml:"token"`
	// (Optional) Use the legacy authentication workflow, preferring environment variables over backend configuration.
	//
	// Defaults to true.
	// This behavior does not align with the authentication flow of the AWS CLI or SDK's, and will be removed in the future.
	// Experimental.
	UseLegacyWorkflow *bool `field:"optional" json:"useLegacyWorkflow" yaml:"useLegacyWorkflow"`
	// (Optional) Enable path-style S3 URLs (https://<HOST>/<BUCKET> instead of https://<BUCKET>.<HOST>).
	// Experimental.
	UsePathStyle *bool `field:"optional" json:"usePathStyle" yaml:"usePathStyle"`
	// (Optional) Prefix applied to the state path inside the bucket.
	//
	// This is only relevant when using a non-default workspace. Defaults to env:
	// Experimental.
	WorkspaceKeyPrefix *string `field:"optional" json:"workspaceKeyPrefix" yaml:"workspaceKeyPrefix"`
}

