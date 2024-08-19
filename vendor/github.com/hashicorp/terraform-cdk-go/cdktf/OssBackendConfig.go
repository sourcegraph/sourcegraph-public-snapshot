// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type OssBackendConfig struct {
	// (Required) The name of the OSS bucket.
	// Experimental.
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// (Optional) Alibaba Cloud access key.
	//
	// It supports environment variables ALICLOUD_ACCESS_KEY and ALICLOUD_ACCESS_KEY_ID.
	// Experimental.
	AccessKey *string `field:"optional" json:"accessKey" yaml:"accessKey"`
	// (Optional) Object ACL to be applied to the state file.
	// Experimental.
	Acl *string `field:"optional" json:"acl" yaml:"acl"`
	// Deprecated: Use flattened assume role options.
	AssumeRole *OssAssumeRole `field:"optional" json:"assumeRole" yaml:"assumeRole"`
	// (Optional, Available in 1.1.0+) A more restrictive policy to apply to the temporary credentials. This gives you a way to further restrict the permissions for the resulting temporary security credentials. You cannot use this policy to grant permissions that exceed those of the role that is being assumed.
	// Experimental.
	AssumeRolePolicy *string `field:"optional" json:"assumeRolePolicy" yaml:"assumeRolePolicy"`
	// (Optional, Available in 1.1.0+) The ARN of the role to assume. If ARN is set to an empty string, it does not perform role switching. It supports the environment variable ALICLOUD_ASSUME_ROLE_ARN. Terraform executes configuration on account with provided credentials.
	// Experimental.
	AssumeRoleRoleArn *string `field:"optional" json:"assumeRoleRoleArn" yaml:"assumeRoleRoleArn"`
	// (Optional, Available in 1.1.0+) The time after which the established session for assuming role expires. Valid value range: [900-3600] seconds. Default to 3600 (in this case Alibaba Cloud uses its own default value). It supports environment variable ALICLOUD_ASSUME_ROLE_SESSION_EXPIRATION.
	// Experimental.
	AssumeRoleSessionExpiration *float64 `field:"optional" json:"assumeRoleSessionExpiration" yaml:"assumeRoleSessionExpiration"`
	// (Optional, Available in 1.1.0+) The session name to use when assuming the role. If omitted, 'terraform' is passed to the AssumeRole call as session name. It supports environment variable ALICLOUD_ASSUME_ROLE_SESSION_NAME.
	// Experimental.
	AssumeRoleSessionName *string `field:"optional" json:"assumeRoleSessionName" yaml:"assumeRoleSessionName"`
	// (Optional, Available in 0.12.14+) The RAM Role Name attached on a ECS instance for API operations. You can retrieve this from the 'Access Control' section of the Alibaba Cloud console.
	// Experimental.
	EcsRoleName *string `field:"optional" json:"ecsRoleName" yaml:"ecsRoleName"`
	// (Optional) Whether to enable server side encryption of the state file.
	//
	// If it is true, OSS will use 'AES256' encryption algorithm to encrypt state file.
	// Experimental.
	Encrypt *bool `field:"optional" json:"encrypt" yaml:"encrypt"`
	// (Optional) A custom endpoint for the OSS API.
	//
	// It supports environment variables ALICLOUD_OSS_ENDPOINT and OSS_ENDPOINT.
	// Experimental.
	Endpoint *string `field:"optional" json:"endpoint" yaml:"endpoint"`
	// (Optional) The name of the state file.
	//
	// Defaults to terraform.tfstate.
	// Experimental.
	Key *string `field:"optional" json:"key" yaml:"key"`
	// (Optional) The path directory of the state file will be stored.
	//
	// Default to "env:".
	// Experimental.
	Prefix *string `field:"optional" json:"prefix" yaml:"prefix"`
	// (Optional, Available in 0.12.8+) This is the Alibaba Cloud profile name as set in the shared credentials file. It can also be sourced from the ALICLOUD_PROFILE environment variable.
	// Experimental.
	Profile *string `field:"optional" json:"profile" yaml:"profile"`
	// (Optional) The region of the OSS bucket.
	//
	// It supports environment variables ALICLOUD_REGION and ALICLOUD_DEFAULT_REGION.
	// Experimental.
	Region *string `field:"optional" json:"region" yaml:"region"`
	// (Optional) Alibaba Cloud secret access key.
	//
	// It supports environment variables ALICLOUD_SECRET_KEY and ALICLOUD_ACCESS_KEY_SECRET.
	// Experimental.
	SecretKey *string `field:"optional" json:"secretKey" yaml:"secretKey"`
	// (Optional) STS access token.
	//
	// It supports environment variable ALICLOUD_SECURITY_TOKEN.
	// Experimental.
	SecurityToken *string `field:"optional" json:"securityToken" yaml:"securityToken"`
	// (Optional, Available in 0.12.8+) This is the path to the shared credentials file. It can also be sourced from the ALICLOUD_SHARED_CREDENTIALS_FILE environment variable. If this is not set and a profile is specified, ~/.aliyun/config.json will be used.
	// Experimental.
	SharedCredentialsFile *string `field:"optional" json:"sharedCredentialsFile" yaml:"sharedCredentialsFile"`
	// (Optional, Available in 1.0.11+) Custom endpoint for the AliCloud Security Token Service (STS) API. It supports environment variable ALICLOUD_STS_ENDPOINT.
	// Experimental.
	StsEndpoint *string `field:"optional" json:"stsEndpoint" yaml:"stsEndpoint"`
	// (Optional) A custom endpoint for the TableStore API.
	// Experimental.
	TablestoreEndpoint *string `field:"optional" json:"tablestoreEndpoint" yaml:"tablestoreEndpoint"`
	// (Optional) A TableStore table for state locking and consistency.
	//
	// The table must have a primary key named LockID of type String.
	// Experimental.
	TablestoreTable *string `field:"optional" json:"tablestoreTable" yaml:"tablestoreTable"`
}

