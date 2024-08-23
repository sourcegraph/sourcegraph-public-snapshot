// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf


// The local-exec provisioner invokes a local executable after a resource is created.
//
// This invokes a process on the machine running Terraform, not on the resource.
//
// See {@link https://developer.hashicorp.com/terraform/language/resources/provisioners/local-exec local-exec}
// Experimental.
type LocalExecProvisioner struct {
	// This is the command to execute.
	//
	// It can be provided as a relative path to the current working directory or as an absolute path.
	// It is evaluated in a shell, and can use environment variables or Terraform variables.
	// Experimental.
	Command *string `field:"required" json:"command" yaml:"command"`
	// Experimental.
	Type *string `field:"required" json:"type" yaml:"type"`
	// A record of key value pairs representing the environment of the executed command.
	//
	// It inherits the current process environment.
	// Experimental.
	Environment *map[string]*string `field:"optional" json:"environment" yaml:"environment"`
	// If provided, this is a list of interpreter arguments used to execute the command.
	//
	// The first argument is the interpreter itself.
	// It can be provided as a relative path to the current working directory or as an absolute path
	// The remaining arguments are appended prior to the command.
	// This allows building command lines of the form "/bin/bash", "-c", "echo foo".
	// If interpreter is unspecified, sensible defaults will be chosen based on the system OS.
	// Experimental.
	Interpreter *[]*string `field:"optional" json:"interpreter" yaml:"interpreter"`
	// If provided, specifies when Terraform will execute the command.
	//
	// For example, when = destroy specifies that the provisioner will run when the associated resource is destroyed.
	// Experimental.
	When *string `field:"optional" json:"when" yaml:"when"`
	// If provided, specifies the working directory where command will be executed.
	//
	// It can be provided as a relative path to the current working directory or as an absolute path.
	// The directory must exist.
	// Experimental.
	WorkingDir *string `field:"optional" json:"workingDir" yaml:"workingDir"`
}

