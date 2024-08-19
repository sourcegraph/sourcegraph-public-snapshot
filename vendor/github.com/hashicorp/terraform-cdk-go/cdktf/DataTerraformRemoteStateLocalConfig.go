// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Experimental.
type DataTerraformRemoteStateLocalConfig struct {
	// Experimental.
	Defaults *map[string]interface{} `field:"optional" json:"defaults" yaml:"defaults"`
	// Experimental.
	Workspace *string `field:"optional" json:"workspace" yaml:"workspace"`
	// Path where the state file is stored.
	// Default: - defaults to terraform.${stackId}.tfstate
	//
	// Experimental.
	Path *string `field:"optional" json:"path" yaml:"path"`
	// (Optional) The path to non-default workspaces.
	// Experimental.
	WorkspaceDir *string `field:"optional" json:"workspaceDir" yaml:"workspaceDir"`
}

