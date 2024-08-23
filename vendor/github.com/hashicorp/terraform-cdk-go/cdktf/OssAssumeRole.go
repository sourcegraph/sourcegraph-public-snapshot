// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type OssAssumeRole struct {
	// Experimental.
	RoleArn *string `field:"required" json:"roleArn" yaml:"roleArn"`
	// Experimental.
	Policy *string `field:"optional" json:"policy" yaml:"policy"`
	// Experimental.
	SessionExpiration *float64 `field:"optional" json:"sessionExpiration" yaml:"sessionExpiration"`
	// Experimental.
	SessionName *string `field:"optional" json:"sessionName" yaml:"sessionName"`
}

