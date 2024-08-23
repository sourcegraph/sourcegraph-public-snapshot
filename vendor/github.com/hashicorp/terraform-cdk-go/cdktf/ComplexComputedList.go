// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cdktf

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/hashicorp/terraform-cdk-go/cdktf/jsii"
)

// Deprecated: Going to be replaced by Array of ComplexListItem
// and will be removed in the future.
type ComplexComputedList interface {
	IInterpolatingParent
	IResolvable
	ITerraformAddressable
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	ComplexComputedListIndex() *string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	SetComplexComputedListIndex(val *string)
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	CreationStack() *[]*string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	Fqn() *string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	TerraformAttribute() *string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	SetTerraformAttribute(val *string)
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	TerraformResource() IInterpolatingParent
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	SetTerraformResource(val IInterpolatingParent)
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	WrapsSet() *bool
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	SetWrapsSet(val *bool)
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	ComputeFqn() *string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{}
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetBooleanAttribute(terraformAttribute *string) IResolvable
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetListAttribute(terraformAttribute *string) *[]*string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetNumberAttribute(terraformAttribute *string) *float64
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetNumberListAttribute(terraformAttribute *string) *[]*float64
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetStringAttribute(terraformAttribute *string) *string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	GetStringMapAttribute(terraformAttribute *string) *map[string]*string
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	InterpolationForAttribute(property *string) IResolvable
	// Produce the Token's value at resolution time.
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	Resolve(_context IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Deprecated: Going to be replaced by Array of ComplexListItem
	// and will be removed in the future.
	ToString() *string
}

// The jsii proxy struct for ComplexComputedList
type jsiiProxy_ComplexComputedList struct {
	jsiiProxy_IInterpolatingParent
	jsiiProxy_IResolvable
	jsiiProxy_ITerraformAddressable
}

func (j *jsiiProxy_ComplexComputedList) ComplexComputedListIndex() *string {
	var returns *string
	_jsii_.Get(
		j,
		"complexComputedListIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComplexComputedList) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComplexComputedList) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComplexComputedList) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComplexComputedList) TerraformResource() IInterpolatingParent {
	var returns IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComplexComputedList) WrapsSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"wrapsSet",
		&returns,
	)
	return returns
}


// Deprecated: Going to be replaced by Array of ComplexListItem
// and will be removed in the future.
func NewComplexComputedList(terraformResource IInterpolatingParent, terraformAttribute *string, complexComputedListIndex *string, wrapsSet *bool) ComplexComputedList {
	_init_.Initialize()

	if err := validateNewComplexComputedListParameters(terraformResource, terraformAttribute, complexComputedListIndex); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComplexComputedList{}

	_jsii_.Create(
		"cdktf.ComplexComputedList",
		[]interface{}{terraformResource, terraformAttribute, complexComputedListIndex, wrapsSet},
		&j,
	)

	return &j
}

// Deprecated: Going to be replaced by Array of ComplexListItem
// and will be removed in the future.
func NewComplexComputedList_Override(c ComplexComputedList, terraformResource IInterpolatingParent, terraformAttribute *string, complexComputedListIndex *string, wrapsSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"cdktf.ComplexComputedList",
		[]interface{}{terraformResource, terraformAttribute, complexComputedListIndex, wrapsSet},
		c,
	)
}

func (j *jsiiProxy_ComplexComputedList)SetComplexComputedListIndex(val *string) {
	if err := j.validateSetComplexComputedListIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexComputedListIndex",
		val,
	)
}

func (j *jsiiProxy_ComplexComputedList)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComplexComputedList)SetTerraformResource(val IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_ComplexComputedList)SetWrapsSet(val *bool) {
	_jsii_.Set(
		j,
		"wrapsSet",
		val,
	)
}

func (c *jsiiProxy_ComplexComputedList) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := c.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetBooleanAttribute(terraformAttribute *string) IResolvable {
	if err := c.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		c,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := c.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		c,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := c.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		c,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := c.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		c,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := c.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		c,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := c.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		c,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetStringAttribute(terraformAttribute *string) *string {
	if err := c.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		c,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := c.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		c,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) InterpolationForAttribute(property *string) IResolvable {
	if err := c.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns IResolvable

	_jsii_.Invoke(
		c,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) Resolve(_context IResolveContext) interface{} {
	if err := c.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		c,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComplexComputedList) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

