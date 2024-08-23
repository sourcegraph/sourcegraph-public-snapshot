// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// Most provisioners require access to the remote resource via SSH or WinRM and expect a nested connection block with details about how to connect.
//
// Refer to {@link https://developer.hashicorp.com/terraform/language/resources/provisioners/connection connection}
// Experimental.
type SSHProvisionerConnection struct {
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
	// Set to false to disable using ssh-agent to authenticate.
	//
	// On Windows the only supported SSH authentication agent is Pageant.
	// Experimental.
	Agent *string `field:"optional" json:"agent" yaml:"agent"`
	// The preferred identity from the ssh agent for authentication.
	// Experimental.
	AgentIdentity *string `field:"optional" json:"agentIdentity" yaml:"agentIdentity"`
	// The contents of a signed CA Certificate.
	//
	// The certificate argument must be used in conjunction with a bastion_private_key.
	// These can be loaded from a file on disk using the the file function.
	// Experimental.
	BastionCertificate *string `field:"optional" json:"bastionCertificate" yaml:"bastionCertificate"`
	// Setting this enables the bastion Host connection.
	//
	// The provisioner will connect to bastion_host first, and then connect from there to host.
	// Experimental.
	BastionHost *string `field:"optional" json:"bastionHost" yaml:"bastionHost"`
	// The public key from the remote host or the signing CA, used to verify the host connection.
	// Experimental.
	BastionHostKey *string `field:"optional" json:"bastionHostKey" yaml:"bastionHostKey"`
	// The password to use for the bastion host.
	// Experimental.
	BastionPassword *string `field:"optional" json:"bastionPassword" yaml:"bastionPassword"`
	// The port to use connect to the bastion host.
	// Experimental.
	BastionPort *float64 `field:"optional" json:"bastionPort" yaml:"bastionPort"`
	// The contents of an SSH key file to use for the bastion host.
	//
	// These can be loaded from a file on disk using the file function.
	// Experimental.
	BastionPrivateKey *string `field:"optional" json:"bastionPrivateKey" yaml:"bastionPrivateKey"`
	// The user for the connection to the bastion host.
	// Experimental.
	BastionUser *string `field:"optional" json:"bastionUser" yaml:"bastionUser"`
	// The contents of a signed CA Certificate.
	//
	// The certificate argument must be used in conjunction with a private_key.
	// These can be loaded from a file on disk using the the file function.
	// Experimental.
	Certificate *string `field:"optional" json:"certificate" yaml:"certificate"`
	// The public key from the remote host or the signing CA, used to verify the connection.
	// Experimental.
	HostKey *string `field:"optional" json:"hostKey" yaml:"hostKey"`
	// The password to use for the connection.
	// Experimental.
	Password *string `field:"optional" json:"password" yaml:"password"`
	// The port to connect to.
	// Default: 22.
	//
	// Experimental.
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// The contents of an SSH key to use for the connection.
	//
	// These can be loaded from a file on disk using the file function.
	// This takes preference over password if provided.
	// Experimental.
	PrivateKey *string `field:"optional" json:"privateKey" yaml:"privateKey"`
	// Setting this enables the SSH over HTTP connection.
	//
	// This host will be connected to first, and then the host or bastion_host connection will be made from there.
	// Experimental.
	ProxyHost *string `field:"optional" json:"proxyHost" yaml:"proxyHost"`
	// The port to use connect to the proxy host.
	// Experimental.
	ProxyPort *float64 `field:"optional" json:"proxyPort" yaml:"proxyPort"`
	// The ssh connection also supports the following fields to facilitate connections by SSH over HTTP proxy.
	// Experimental.
	ProxyScheme *string `field:"optional" json:"proxyScheme" yaml:"proxyScheme"`
	// The username to use connect to the private proxy host.
	//
	// This argument should be specified only if authentication is required for the HTTP Proxy server.
	// Experimental.
	ProxyUserName *string `field:"optional" json:"proxyUserName" yaml:"proxyUserName"`
	// The password to use connect to the private proxy host.
	//
	// This argument should be specified only if authentication is required for the HTTP Proxy server.
	// Experimental.
	ProxyUserPassword *string `field:"optional" json:"proxyUserPassword" yaml:"proxyUserPassword"`
	// The path used to copy scripts meant for remote execution.
	//
	// Refer to {@link https://developer.hashicorp.com/terraform/language/resources/provisioners/connection#how-provisioners-execute-remote-scripts How Provisioners Execute Remote Scripts below for more details}
	// Experimental.
	ScriptPath *string `field:"optional" json:"scriptPath" yaml:"scriptPath"`
	// The target platform to connect to.
	//
	// Valid values are "windows" and "unix".
	// If the platform is set to windows, the default script_path is c:\windows\temp\terraform_%RAND%.cmd, assuming the SSH default shell is cmd.exe.
	// If the SSH default shell is PowerShell, set script_path to "c:/windows/temp/terraform_%RAND%.ps1"
	// Default: unix.
	//
	// Experimental.
	TargetPlatform *string `field:"optional" json:"targetPlatform" yaml:"targetPlatform"`
	// The timeout to wait for the connection to become available.
	//
	// Should be provided as a string (e.g., "30s" or "5m".)
	// Default: 5m.
	//
	// Experimental.
	Timeout *string `field:"optional" json:"timeout" yaml:"timeout"`
	// The user to use for the connection.
	// Default: root.
	//
	// Experimental.
	User *string `field:"optional" json:"user" yaml:"user"`
}

