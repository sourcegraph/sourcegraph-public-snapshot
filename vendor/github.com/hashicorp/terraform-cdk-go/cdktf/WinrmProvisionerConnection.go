// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Most provisioners require access to the remote resource via SSH or WinRM and expect a nested connection block with details about how to connect.
//
// See {@link https://developer.hashicorp.com/terraform/language/resources/provisioners/connection connection}
// Experimental.
type WinrmProvisionerConnection struct {
	// The address of the resource to connect to.
	// Experimental.
	Host *string `field:"required" json:"host" yaml:"host"`
	// The connection type.
	//
	// Valid values are "ssh" and "winrm".
	// Provisioners typically assume that the remote system runs Microsoft Windows when using WinRM.
	// Behaviors based on the SSH target_platform will force Windows-specific behavior for WinRM, unless otherwise specified.
	// Experimental.
	Type *string `field:"required" json:"type" yaml:"type"`
	// The CA certificate to validate against.
	// Experimental.
	Cacert *string `field:"optional" json:"cacert" yaml:"cacert"`
	// Set to true to connect using HTTPS instead of HTTP.
	// Experimental.
	Https *bool `field:"optional" json:"https" yaml:"https"`
	// Set to true to skip validating the HTTPS certificate chain.
	// Experimental.
	Insecure *bool `field:"optional" json:"insecure" yaml:"insecure"`
	// The password to use for the connection.
	// Experimental.
	Password *string `field:"optional" json:"password" yaml:"password"`
	// The port to connect to.
	// Default: 22.
	//
	// Experimental.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// The path used to copy scripts meant for remote execution.
	//
	// Refer to {@link https://developer.hashicorp.com/terraform/language/resources/provisioners/connection#how-provisioners-execute-remote-scripts How Provisioners Execute Remote Scripts below for more details}
	// Experimental.
	ScriptPath *string `field:"optional" json:"scriptPath" yaml:"scriptPath"`
	// The timeout to wait for the connection to become available.
	//
	// Should be provided as a string (e.g., "30s" or "5m".)
	// Default: 5m.
	//
	// Experimental.
	Timeout *string `field:"optional" json:"timeout" yaml:"timeout"`
	// Set to true to use NTLM authentication rather than default (basic authentication), removing the requirement for basic authentication to be enabled within the target guest.
	//
	// Refer to Authentication for Remote Connections in the Windows App Development documentation for more details.
	// Experimental.
	UseNtlm *bool `field:"optional" json:"useNtlm" yaml:"useNtlm"`
	// The user to use for the connection.
	// Default: root.
	//
	// Experimental.
	User *string `field:"optional" json:"user" yaml:"user"`
}

