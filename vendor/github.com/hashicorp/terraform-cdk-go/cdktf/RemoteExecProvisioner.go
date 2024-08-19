// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// The remote-exec provisioner invokes a script on a remote resource after it is created.
//
// This can be used to run a configuration management tool, bootstrap into a cluster, etc
// The remote-exec provisioner requires a connection and supports both ssh and winrm.
//
// See {@link https://developer.hashicorp.com/terraform/language/resources/provisioners/remote-exec remote-exec}
// Experimental.
type RemoteExecProvisioner struct {
	// Experimental.
	Type *string `field:"required" json:"type" yaml:"type"`
	// Most provisioners require access to the remote resource via SSH or WinRM and expect a nested connection block with details about how to connect.
	//
	// A connection must be provided here or in the parent resource.
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// This is a list of command strings.
	//
	// They are executed in the order they are provided.
	// This cannot be provided with script or scripts.
	// Experimental.
	Inline *[]*string `field:"optional" json:"inline" yaml:"inline"`
	// This is a path (relative or absolute) to a local script that will be copied to the remote resource and then executed.
	//
	// This cannot be provided with inline or scripts.
	// Experimental.
	Script *string `field:"optional" json:"script" yaml:"script"`
	// This is a list of paths (relative or absolute) to local scripts that will be copied to the remote resource and then executed.
	//
	// They are executed in the order they are provided.
	// This cannot be provided with inline or script.
	// Experimental.
	Scripts *[]*string `field:"optional" json:"scripts" yaml:"scripts"`
}

