// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type S3BackendAssumeRoleWithWebIdentityConfig struct {
	// (Optional) The duration individual credentials will be valid.
	//
	// Credentials are automatically renewed up to the maximum defined by the AWS account.
	// Specified using the format <hours>h<minutes>m<seconds>s with any unit being optional.
	// For example, an hour and a half can be specified as 1h30m or 90m.
	// Must be between 15 minutes (15m) and 12 hours (12h).
	// Experimental.
	Duration *string `field:"optional" json:"duration" yaml:"duration"`
	// (Optional) IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.
	// Experimental.
	Policy *string `field:"optional" json:"policy" yaml:"policy"`
	// (Optional) Set of Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.
	// Experimental.
	PolicyArns *[]*string `field:"optional" json:"policyArns" yaml:"policyArns"`
	// (Required) Amazon Resource Name (ARN) of the IAM Role to assume.
	//
	// Can also be set with the AWS_ROLE_ARN environment variable.
	// Experimental.
	RoleArn *string `field:"optional" json:"roleArn" yaml:"roleArn"`
	// (Optional) Session name to use when assuming the role.
	//
	// Can also be set with the AWS_ROLE_SESSION_NAME environment variable.
	// Experimental.
	SessionName *string `field:"optional" json:"sessionName" yaml:"sessionName"`
	// (Optional) The value of a web identity token from an OpenID Connect (OIDC) or OAuth provider.
	//
	// One of web_identity_token or web_identity_token_file is required.
	// Experimental.
	WebIdentityToken *string `field:"optional" json:"webIdentityToken" yaml:"webIdentityToken"`
	// (Optional) File containing a web identity token from an OpenID Connect (OIDC) or OAuth provider.
	//
	// One of web_identity_token_file or web_identity_token is required.
	// Can also be set with the AWS_WEB_IDENTITY_TOKEN_FILE environment variable.
	// Experimental.
	WebIdentityTokenFile *string `field:"optional" json:"webIdentityTokenFile" yaml:"webIdentityTokenFile"`
}

