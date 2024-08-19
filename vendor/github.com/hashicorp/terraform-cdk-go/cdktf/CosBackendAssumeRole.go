// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type CosBackendAssumeRole struct {
	// (Required) The ARN of the role to assume.
	//
	// It can be sourced from the TENCENTCLOUD_ASSUME_ROLE_ARN.
	// Experimental.
	RoleArn *string `field:"required" json:"roleArn" yaml:"roleArn"`
	// (Required) The duration of the session when making the AssumeRole call.
	//
	// Its value ranges from 0 to 43200(seconds), and default is 7200 seconds.
	// It can be sourced from the TENCENTCLOUD_ASSUME_ROLE_SESSION_DURATION.
	// Experimental.
	SessionDuration *float64 `field:"required" json:"sessionDuration" yaml:"sessionDuration"`
	// (Required) The session name to use when making the AssumeRole call.
	//
	// It can be sourced from the TENCENTCLOUD_ASSUME_ROLE_SESSION_NAME.
	// Experimental.
	SessionName *string `field:"required" json:"sessionName" yaml:"sessionName"`
	// (Optional) A more restrictive policy when making the AssumeRole call.
	//
	// Its content must not contains principal elements.
	// Please refer to {@link https://www.tencentcloud.com/document/product/598/10603 policies syntax logic}.
	// Experimental.
	Policy interface{} `field:"optional" json:"policy" yaml:"policy"`
}

