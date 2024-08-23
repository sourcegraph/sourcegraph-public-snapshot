// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// The Cloud Backend synthesizes a {@link https://developer.hashicorp.com/terraform/cli/cloud/settings#the-cloud-block cloud block}. The cloud block is a nested block within the top-level terraform settings block. It specifies which Terraform Cloud workspaces to use for the current working directory. The cloud block only affects Terraform CLI's behavior. When Terraform Cloud uses a configuration that contains a cloud block - for example, when a workspace is configured to use a VCS provider directly - it ignores the block and behaves according to its own workspace settings.
//
// https://developer.hashicorp.com/terraform/cli/cloud/settings#arguments
// Experimental.
type CloudBackendConfig struct {
	// The name of the organization containing the workspace(s) the current configuration should use.
	// Experimental.
	Organization *string `field:"required" json:"organization" yaml:"organization"`
	// A nested block that specifies which remote Terraform Cloud workspaces to use for the current configuration.
	//
	// The workspaces block must contain exactly one of the following arguments, each denoting a strategy for how workspaces should be mapped:.
	// Experimental.
	Workspaces interface{} `field:"required" json:"workspaces" yaml:"workspaces"`
	// The hostname of a Terraform Enterprise installation, if using Terraform Enterprise.
	// Default: app.terraform.io
	//
	// Experimental.
	Hostname *string `field:"optional" json:"hostname" yaml:"hostname"`
	// The token used to authenticate with Terraform Cloud.
	//
	// We recommend omitting the token from the configuration, and instead using terraform login or manually configuring credentials in the CLI config file.
	// Experimental.
	Token *string `field:"optional" json:"token" yaml:"token"`
}

