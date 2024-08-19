package computeurlmap

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap/internal"
)

type ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference interface {
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
	InternalValue() *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicy
	SetInternalValue(val *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicy)
	NumRetries() *float64
	SetNumRetries(val *float64)
	NumRetriesInput() *float64
	PerTryTimeout() ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeoutOutputReference
	PerTryTimeoutInput() *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeout
	RetryConditions() *[]*string
	SetRetryConditions(val *[]*string)
	RetryConditionsInput() *[]*string
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
	PutPerTryTimeout(value *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeout)
	ResetNumRetries()
	ResetPerTryTimeout()
	ResetRetryConditions()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference
type jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) InternalValue() *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicy {
	var returns *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicy
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) NumRetries() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"numRetries",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) NumRetriesInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"numRetriesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) PerTryTimeout() ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeoutOutputReference {
	var returns ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeoutOutputReference
	_jsii_.Get(
		j,
		"perTryTimeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) PerTryTimeoutInput() *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeout {
	var returns *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeout
	_jsii_.Get(
		j,
		"perTryTimeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) RetryConditions() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"retryConditions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) RetryConditionsInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"retryConditionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference {
	_init_.Initialize()

	if err := validateNewComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference_Override(c ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference)SetInternalValue(val *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicy) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference)SetNumRetries(val *float64) {
	if err := j.validateSetNumRetriesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"numRetries",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference)SetRetryConditions(val *[]*string) {
	if err := j.validateSetRetryConditionsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryConditions",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) PutPerTryTimeout(value *ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyPerTryTimeout) {
	if err := c.validatePutPerTryTimeoutParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putPerTryTimeout",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) ResetNumRetries() {
	_jsii_.InvokeVoid(
		c,
		"resetNumRetries",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) ResetPerTryTimeout() {
	_jsii_.InvokeVoid(
		c,
		"resetPerTryTimeout",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) ResetRetryConditions() {
	_jsii_.InvokeVoid(
		c,
		"resetRetryConditions",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionRetryPolicyOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

