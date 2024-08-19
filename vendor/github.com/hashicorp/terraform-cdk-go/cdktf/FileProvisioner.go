// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// The file provisioner copies files or directories from the machine running Terraform to the newly created resource.
//
// The file provisioner supports both ssh and winrm type connections.
//
// See {@link https://developer.hashicorp.com/terraform/language/resources/provisioners/file file}
// Experimental.
type FileProvisioner struct {
	// The source file or directory.
	//
	// Specify it either relative to the current working directory or as an absolute path.
	// This argument cannot be combined with content.
	// Experimental.
	Destination *string `field:"required" json:"destination" yaml:"destination"`
	// Experimental.
	Type *string `field:"required" json:"type" yaml:"type"`
	// Most provisioners require access to the remote resource via SSH or WinRM and expect a nested connection block with details about how to connect.
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// The destination path to write to on the remote system.
	//
	// See Destination Paths below for more information.
	// Experimental.
	Content *string `field:"optional" json:"content" yaml:"content"`
	// The direct content to copy on the destination.
	//
	// If destination is a file, the content will be written on that file.
	// In case of a directory, a file named tf-file-content is created inside that directory.
	// We recommend using a file as the destination when using content.
	// This argument cannot be combined with source.
	// Experimental.
	Source *string `field:"optional" json:"source" yaml:"source"`
}

