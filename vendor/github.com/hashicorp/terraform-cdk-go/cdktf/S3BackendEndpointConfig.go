// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type S3BackendEndpointConfig struct {
	// (Optional) Custom endpoint URL for the AWS DynamoDB API.
	//
	// This can also be sourced from the environment variable AWS_ENDPOINT_URL_DYNAMODB or the deprecated environment variable AWS_DYNAMODB_ENDPOINT.
	// Experimental.
	Dynamodb *string `field:"optional" json:"dynamodb" yaml:"dynamodb"`
	// (Optional) Custom endpoint URL for the AWS IAM API.
	//
	// This can also be sourced from the environment variable AWS_ENDPOINT_URL_IAM or the deprecated environment variable AWS_IAM_ENDPOINT.
	// Experimental.
	Iam *string `field:"optional" json:"iam" yaml:"iam"`
	// (Optional) Custom endpoint URL for the AWS S3 API.
	//
	// This can also be sourced from the environment variable AWS_ENDPOINT_URL_S3 or the deprecated environment variable AWS_S3_ENDPOINT.
	// Experimental.
	S3 *string `field:"optional" json:"s3" yaml:"s3"`
	// (Optional) Custom endpoint URL for the AWS IAM Identity Center (formerly known as AWS SSO) API.
	//
	// This can also be sourced from the environment variable AWS_ENDPOINT_URL_SSO.
	// Experimental.
	Sso *string `field:"optional" json:"sso" yaml:"sso"`
	// (Optional) Custom endpoint URL for the AWS STS API.
	//
	// This can also be sourced from the environment variable AWS_ENDPOINT_URL_STS or the deprecated environment variable AWS_STS_ENDPOINT.
	// Experimental.
	Sts *string `field:"optional" json:"sts" yaml:"sts"`
}

