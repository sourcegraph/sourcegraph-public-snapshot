// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type S3BackendAssumeRoleConfig struct {
	// (Required) Amazon Resource Name (ARN) of the IAM Role to assume.
	// Experimental.
	RoleArn *string `field:"required" json:"roleArn" yaml:"roleArn"`
	// (Optional) The duration individual credentials will be valid.
	//
	// Credentials are automatically renewed up to the maximum defined by the AWS account.
	// Specified using the format <hours>h<minutes>m<seconds>s with any unit being optional.
	// For example, an hour and a half can be specified as 1h30m or 90m.
	// Must be between 15 minutes (15m) and 12 hours (12h).
	// Experimental.
	Duration *string `field:"optional" json:"duration" yaml:"duration"`
	// (Optional) External identifier to use when assuming the role.
	// Experimental.
	ExternalId *string `field:"optional" json:"externalId" yaml:"externalId"`
	// (Optional) IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.
	// Experimental.
	Policy *string `field:"optional" json:"policy" yaml:"policy"`
	// (Optional) Set of Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.
	// Experimental.
	PolicyArns *[]*string `field:"optional" json:"policyArns" yaml:"policyArns"`
	// (Optional) Session name to use when assuming the role.
	// Experimental.
	SessionName *string `field:"optional" json:"sessionName" yaml:"sessionName"`
	// (Optional) Source identity specified by the principal assuming the.
	// Experimental.
	SourceIdentity *string `field:"optional" json:"sourceIdentity" yaml:"sourceIdentity"`
	// (Optional) Map of assume role session tags.
	// Experimental.
	Tags *map[string]*string `field:"optional" json:"tags" yaml:"tags"`
	// (Optional) Set of assume role session tag keys to pass to any subsequent sessions.
	// Experimental.
	TransitiveTagKeys *[]*string `field:"optional" json:"transitiveTagKeys" yaml:"transitiveTagKeys"`
}

