// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Stores the state as an object in a configurable prefix in a given bucket on Tencent Cloud Object Storage (COS).
//
// This backend supports state locking.
//
// Warning! It is highly recommended that you enable Object Versioning on the COS bucket to allow for state recovery in the case of accidental deletions and human error.
//
// Read more about this backend in the Terraform docs:
// https://developer.hashicorp.com/terraform/language/settings/backends/cos
// Experimental.
type CosBackendConfig struct {
	// (Required) The name of the COS bucket.
	//
	// You shall manually create it first.
	// Experimental.
	Bucket *string `field:"required" json:"bucket" yaml:"bucket"`
	// (Optional) Whether to enable global Acceleration.
	//
	// Defaults to false.
	// Experimental.
	Accelerate *bool `field:"optional" json:"accelerate" yaml:"accelerate"`
	// (Optional) Object ACL to be applied to the state file, allows private and public-read.
	//
	// Defaults to private.
	// Experimental.
	Acl *string `field:"optional" json:"acl" yaml:"acl"`
	// (Optional) The assume_role block.
	//
	// If provided, terraform will attempt to assume this role using the supplied credentials.
	// Experimental.
	AssumeRole *CosBackendAssumeRole `field:"optional" json:"assumeRole" yaml:"assumeRole"`
	// (Optional) The root domain of the API request.
	//
	// Defaults to tencentcloudapi.com.
	// It supports the environment variable TENCENTCLOUD_DOMAIN.
	// Experimental.
	Domain *string `field:"optional" json:"domain" yaml:"domain"`
	// (Optional) Whether to enable server side encryption of the state file.
	//
	// If it is true, COS will use 'AES256' encryption algorithm to encrypt state file.
	// Experimental.
	Encrypt *bool `field:"optional" json:"encrypt" yaml:"encrypt"`
	// (Optional) The Custom Endpoint for the COS backend.
	//
	// It supports the environment variable TENCENTCLOUD_ENDPOINT.
	// Experimental.
	Endpoint *string `field:"optional" json:"endpoint" yaml:"endpoint"`
	// (Optional) The path for saving the state file in bucket.
	//
	// Defaults to terraform.tfstate.
	// Experimental.
	Key *string `field:"optional" json:"key" yaml:"key"`
	// (Optional) The directory for saving the state file in bucket.
	//
	// Default to "env:".
	// Experimental.
	Prefix *string `field:"optional" json:"prefix" yaml:"prefix"`
	// (Optional) The region of the COS bucket.
	//
	// It supports environment variables TENCENTCLOUD_REGION.
	// Experimental.
	Region *string `field:"optional" json:"region" yaml:"region"`
	// (Optional) Secret id of Tencent Cloud.
	//
	// It supports environment variables TENCENTCLOUD_SECRET_ID.
	// Experimental.
	SecretId *string `field:"optional" json:"secretId" yaml:"secretId"`
	// (Optional) Secret key of Tencent Cloud.
	//
	// It supports environment variables TENCENTCLOUD_SECRET_KEY.
	// Experimental.
	SecretKey *string `field:"optional" json:"secretKey" yaml:"secretKey"`
	// (Optional) TencentCloud Security Token of temporary access credentials.
	//
	// It supports environment variables TENCENTCLOUD_SECURITY_TOKEN.
	// Experimental.
	SecurityToken *string `field:"optional" json:"securityToken" yaml:"securityToken"`
}

