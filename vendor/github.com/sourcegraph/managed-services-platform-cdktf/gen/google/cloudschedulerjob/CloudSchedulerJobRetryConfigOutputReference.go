package cloudschedulerjob

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudschedulerjob/internal"
)

type CloudSchedulerJobRetryConfigOutputReference interface {
	cdktf.ComplexObject
	// the index of the complex object in a list.
	// Experimental.
	ComplexObjectIndex() interface{}
	// Experimental.
	SetComplexObjectIndex(val interface{})
	// set to true if this item is from inside a set and needs tolist() for accessing it set to "0" for single list items.
	// Experimental.
	ComplexObjectIsFromSet() *bool
	// Experimental.
	SetComplexObjectIsFromSet(val *bool)
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	// Experimental.
	Fqn() *string
	InternalValue() *CloudSchedulerJobRetryConfig
	SetInternalValue(val *CloudSchedulerJobRetryConfig)
	MaxBackoffDuration() *string
	SetMaxBackoffDuration(val *string)
	MaxBackoffDurationInput() *string
	MaxDoublings() *float64
	SetMaxDoublings(val *float64)
	MaxDoublingsInput() *float64
	MaxRetryDuration() *string
	SetMaxRetryDuration(val *string)
	MaxRetryDurationInput() *string
	MinBackoffDuration() *string
	SetMinBackoffDuration(val *string)
	MinBackoffDurationInput() *string
	RetryCount() *float64
	SetRetryCount(val *float64)
	RetryCountInput() *float64
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	// Experimental.
	ComputeFqn() *string
	// Experimental.
	GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{}
	// Experimental.
	GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable
	// Experimental.
	GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool
	// Experimental.
	GetListAttribute(terraformAttribute *string) *[]*string
	// Experimental.
	GetNumberAttribute(terraformAttribute *string) *float64
	// Experimental.
	GetNumberListAttribute(terraformAttribute *string) *[]*float64
	// Experimental.
	GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64
	// Experimental.
	GetStringAttribute(terraformAttribute *string) *string
	// Experimental.
	GetStringMapAttribute(terraformAttribute *string) *map[string]*string
	// Experimental.
	InterpolationAsList() cdktf.IResolvable
	// Experimental.
	InterpolationForAttribute(property *string) cdktf.IResolvable
	ResetMaxBackoffDuration()
	ResetMaxDoublings()
	ResetMaxRetryDuration()
	ResetMinBackoffDuration()
	ResetRetryCount()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for CloudSchedulerJobRetryConfigOutputReference
type jsiiProxy_CloudSchedulerJobRetryConfigOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) InternalValue() *CloudSchedulerJobRetryConfig {
	var returns *CloudSchedulerJobRetryConfig
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MaxBackoffDuration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxBackoffDuration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MaxBackoffDurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxBackoffDurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MaxDoublings() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxDoublings",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MaxDoublingsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxDoublingsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MaxRetryDuration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxRetryDuration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MaxRetryDurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"maxRetryDurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MinBackoffDuration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"minBackoffDuration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) MinBackoffDurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"minBackoffDurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) RetryCount() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryCount",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) RetryCountInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryCountInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewCloudSchedulerJobRetryConfigOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) CloudSchedulerJobRetryConfigOutputReference {
	_init_.Initialize()

	if err := validateNewCloudSchedulerJobRetryConfigOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_CloudSchedulerJobRetryConfigOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJobRetryConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewCloudSchedulerJobRetryConfigOutputReference_Override(c CloudSchedulerJobRetryConfigOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJobRetryConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetInternalValue(val *CloudSchedulerJobRetryConfig) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetMaxBackoffDuration(val *string) {
	if err := j.validateSetMaxBackoffDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxBackoffDuration",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetMaxDoublings(val *float64) {
	if err := j.validateSetMaxDoublingsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxDoublings",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetMaxRetryDuration(val *string) {
	if err := j.validateSetMaxRetryDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maxRetryDuration",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetMinBackoffDuration(val *string) {
	if err := j.validateSetMinBackoffDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"minBackoffDuration",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetRetryCount(val *float64) {
	if err := j.validateSetRetryCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryCount",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := c.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := c.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ResetMaxBackoffDuration() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxBackoffDuration",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ResetMaxDoublings() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxDoublings",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ResetMaxRetryDuration() {
	_jsii_.InvokeVoid(
		c,
		"resetMaxRetryDuration",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ResetMinBackoffDuration() {
	_jsii_.InvokeVoid(
		c,
		"resetMinBackoffDuration",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ResetRetryCount() {
	_jsii_.InvokeVoid(
		c,
		"resetRetryCount",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_CloudSchedulerJobRetryConfigOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

