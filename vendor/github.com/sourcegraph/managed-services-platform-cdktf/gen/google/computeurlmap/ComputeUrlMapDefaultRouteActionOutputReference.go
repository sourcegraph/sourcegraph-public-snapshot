package computeurlmap

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap/internal"
)

type ComputeUrlMapDefaultRouteActionOutputReference interface {
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
	CorsPolicy() ComputeUrlMapDefaultRouteActionCorsPolicyOutputReference
	CorsPolicyInput() *ComputeUrlMapDefaultRouteActionCorsPolicy
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	FaultInjectionPolicy() ComputeUrlMapDefaultRouteActionFaultInjectionPolicyOutputReference
	FaultInjectionPolicyInput() *ComputeUrlMapDefaultRouteActionFaultInjectionPolicy
	// Experimental.
	Fqn() *string
	InternalValue() *ComputeUrlMapDefaultRouteAction
	SetInternalValue(val *ComputeUrlMapDefaultRouteAction)
	RequestMirrorPolicy() ComputeUrlMapDefaultRouteActionRequestMirrorPolicyOutputReference
	RequestMirrorPolicyInput() *ComputeUrlMapDefaultRouteActionRequestMirrorPolicy
	RetryPolicy() ComputeUrlMapDefaultRouteActionRetryPolicyOutputReference
	RetryPolicyInput() *ComputeUrlMapDefaultRouteActionRetryPolicy
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	Timeout() ComputeUrlMapDefaultRouteActionTimeoutOutputReference
	TimeoutInput() *ComputeUrlMapDefaultRouteActionTimeout
	UrlRewrite() ComputeUrlMapDefaultRouteActionUrlRewriteOutputReference
	UrlRewriteInput() *ComputeUrlMapDefaultRouteActionUrlRewrite
	WeightedBackendServices() ComputeUrlMapDefaultRouteActionWeightedBackendServicesList
	WeightedBackendServicesInput() interface{}
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
	PutCorsPolicy(value *ComputeUrlMapDefaultRouteActionCorsPolicy)
	PutFaultInjectionPolicy(value *ComputeUrlMapDefaultRouteActionFaultInjectionPolicy)
	PutRequestMirrorPolicy(value *ComputeUrlMapDefaultRouteActionRequestMirrorPolicy)
	PutRetryPolicy(value *ComputeUrlMapDefaultRouteActionRetryPolicy)
	PutTimeout(value *ComputeUrlMapDefaultRouteActionTimeout)
	PutUrlRewrite(value *ComputeUrlMapDefaultRouteActionUrlRewrite)
	PutWeightedBackendServices(value interface{})
	ResetCorsPolicy()
	ResetFaultInjectionPolicy()
	ResetRequestMirrorPolicy()
	ResetRetryPolicy()
	ResetTimeout()
	ResetUrlRewrite()
	ResetWeightedBackendServices()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeUrlMapDefaultRouteActionOutputReference
type jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) CorsPolicy() ComputeUrlMapDefaultRouteActionCorsPolicyOutputReference {
	var returns ComputeUrlMapDefaultRouteActionCorsPolicyOutputReference
	_jsii_.Get(
		j,
		"corsPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) CorsPolicyInput() *ComputeUrlMapDefaultRouteActionCorsPolicy {
	var returns *ComputeUrlMapDefaultRouteActionCorsPolicy
	_jsii_.Get(
		j,
		"corsPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) FaultInjectionPolicy() ComputeUrlMapDefaultRouteActionFaultInjectionPolicyOutputReference {
	var returns ComputeUrlMapDefaultRouteActionFaultInjectionPolicyOutputReference
	_jsii_.Get(
		j,
		"faultInjectionPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) FaultInjectionPolicyInput() *ComputeUrlMapDefaultRouteActionFaultInjectionPolicy {
	var returns *ComputeUrlMapDefaultRouteActionFaultInjectionPolicy
	_jsii_.Get(
		j,
		"faultInjectionPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) InternalValue() *ComputeUrlMapDefaultRouteAction {
	var returns *ComputeUrlMapDefaultRouteAction
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) RequestMirrorPolicy() ComputeUrlMapDefaultRouteActionRequestMirrorPolicyOutputReference {
	var returns ComputeUrlMapDefaultRouteActionRequestMirrorPolicyOutputReference
	_jsii_.Get(
		j,
		"requestMirrorPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) RequestMirrorPolicyInput() *ComputeUrlMapDefaultRouteActionRequestMirrorPolicy {
	var returns *ComputeUrlMapDefaultRouteActionRequestMirrorPolicy
	_jsii_.Get(
		j,
		"requestMirrorPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) RetryPolicy() ComputeUrlMapDefaultRouteActionRetryPolicyOutputReference {
	var returns ComputeUrlMapDefaultRouteActionRetryPolicyOutputReference
	_jsii_.Get(
		j,
		"retryPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) RetryPolicyInput() *ComputeUrlMapDefaultRouteActionRetryPolicy {
	var returns *ComputeUrlMapDefaultRouteActionRetryPolicy
	_jsii_.Get(
		j,
		"retryPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) Timeout() ComputeUrlMapDefaultRouteActionTimeoutOutputReference {
	var returns ComputeUrlMapDefaultRouteActionTimeoutOutputReference
	_jsii_.Get(
		j,
		"timeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) TimeoutInput() *ComputeUrlMapDefaultRouteActionTimeout {
	var returns *ComputeUrlMapDefaultRouteActionTimeout
	_jsii_.Get(
		j,
		"timeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) UrlRewrite() ComputeUrlMapDefaultRouteActionUrlRewriteOutputReference {
	var returns ComputeUrlMapDefaultRouteActionUrlRewriteOutputReference
	_jsii_.Get(
		j,
		"urlRewrite",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) UrlRewriteInput() *ComputeUrlMapDefaultRouteActionUrlRewrite {
	var returns *ComputeUrlMapDefaultRouteActionUrlRewrite
	_jsii_.Get(
		j,
		"urlRewriteInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) WeightedBackendServices() ComputeUrlMapDefaultRouteActionWeightedBackendServicesList {
	var returns ComputeUrlMapDefaultRouteActionWeightedBackendServicesList
	_jsii_.Get(
		j,
		"weightedBackendServices",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) WeightedBackendServicesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"weightedBackendServicesInput",
		&returns,
	)
	return returns
}


func NewComputeUrlMapDefaultRouteActionOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeUrlMapDefaultRouteActionOutputReference {
	_init_.Initialize()

	if err := validateNewComputeUrlMapDefaultRouteActionOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapDefaultRouteActionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeUrlMapDefaultRouteActionOutputReference_Override(c ComputeUrlMapDefaultRouteActionOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapDefaultRouteActionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference)SetInternalValue(val *ComputeUrlMapDefaultRouteAction) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) PutCorsPolicy(value *ComputeUrlMapDefaultRouteActionCorsPolicy) {
	if err := c.validatePutCorsPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putCorsPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) PutFaultInjectionPolicy(value *ComputeUrlMapDefaultRouteActionFaultInjectionPolicy) {
	if err := c.validatePutFaultInjectionPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putFaultInjectionPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) PutRequestMirrorPolicy(value *ComputeUrlMapDefaultRouteActionRequestMirrorPolicy) {
	if err := c.validatePutRequestMirrorPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRequestMirrorPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) PutRetryPolicy(value *ComputeUrlMapDefaultRouteActionRetryPolicy) {
	if err := c.validatePutRetryPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRetryPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) PutTimeout(value *ComputeUrlMapDefaultRouteActionTimeout) {
	if err := c.validatePutTimeoutParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeout",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) PutUrlRewrite(value *ComputeUrlMapDefaultRouteActionUrlRewrite) {
	if err := c.validatePutUrlRewriteParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putUrlRewrite",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) PutWeightedBackendServices(value interface{}) {
	if err := c.validatePutWeightedBackendServicesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putWeightedBackendServices",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ResetCorsPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetCorsPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ResetFaultInjectionPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetFaultInjectionPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ResetRequestMirrorPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetRequestMirrorPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ResetRetryPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetRetryPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ResetTimeout() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeout",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ResetUrlRewrite() {
	_jsii_.InvokeVoid(
		c,
		"resetUrlRewrite",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ResetWeightedBackendServices() {
	_jsii_.InvokeVoid(
		c,
		"resetWeightedBackendServices",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeUrlMapDefaultRouteActionOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

