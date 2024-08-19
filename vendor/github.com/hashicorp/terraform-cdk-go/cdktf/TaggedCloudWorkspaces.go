// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// A set of Terraform Cloud workspace tags.
//
// You will be able to use this working directory with any workspaces that have all of the specified tags, and can use the terraform workspace commands to switch between them or create new workspaces. New workspaces will automatically have the specified tags. This option conflicts with name.
// Experimental.
type TaggedCloudWorkspaces interface {
	CloudWorkspace
	// Experimental.
	Project() *string
	// Experimental.
	SetProject(val *string)
	// Experimental.
	Tags() *[]*string
	// Experimental.
	SetTags(val *[]*string)
	// Experimental.
	ToHclTerraform() interface{}
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for TaggedCloudWorkspaces
type jsiiProxy_TaggedCloudWorkspaces struct {
	jsiiProxy_CloudWorkspace
}

func (j *jsiiProxy_TaggedCloudWorkspaces) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_TaggedCloudWorkspaces) Tags() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"tags",
		&returns,
	)
	return returns
}


// Experimental.
func NewTaggedCloudWorkspaces(tags *[]*string, project *string) TaggedCloudWorkspaces {
	_init_.Initialize()

	if err := validateNewTaggedCloudWorkspacesParameters(tags); err != nil {
		panic(err)
	}
	j := jsiiProxy_TaggedCloudWorkspaces{}

	_jsii_.Create(
		"cdktf.TaggedCloudWorkspaces",
		[]interface{}{tags, project},
		&j,
	)

	return &j
}

// Experimental.
func NewTaggedCloudWorkspaces_Override(t TaggedCloudWorkspaces, tags *[]*string, project *string) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.TaggedCloudWorkspaces",
		[]interface{}{tags, project},
		t,
	)
}

func (j *jsiiProxy_TaggedCloudWorkspaces)SetProject(val *string) {
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_TaggedCloudWorkspaces)SetTags(val *[]*string) {
	if err := j.validateSetTagsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"tags",
		val,
	)
}

func (t *jsiiProxy_TaggedCloudWorkspaces) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (t *jsiiProxy_TaggedCloudWorkspaces) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		t,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

