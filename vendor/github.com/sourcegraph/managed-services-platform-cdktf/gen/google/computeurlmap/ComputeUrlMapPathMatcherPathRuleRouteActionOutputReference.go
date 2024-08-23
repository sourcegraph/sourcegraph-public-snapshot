package computeurlmap

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap/internal"
)

type ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference interface {
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
	CorsPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicyOutputReference
	CorsPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicy
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	FaultInjectionPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicyOutputReference
	FaultInjectionPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicy
	// Experimental.
	Fqn() *string
	InternalValue() *ComputeUrlMapPathMatcherPathRuleRouteAction
	SetInternalValue(val *ComputeUrlMapPathMatcherPathRuleRouteAction)
	RequestMirrorPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicyOutputReference
	RequestMirrorPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicy
	RetryPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicyOutputReference
	RetryPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicy
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	Timeout() ComputeUrlMapPathMatcherPathRuleRouteActionTimeoutOutputReference
	TimeoutInput() *ComputeUrlMapPathMatcherPathRuleRouteActionTimeout
	UrlRewrite() ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewriteOutputReference
	UrlRewriteInput() *ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewrite
	WeightedBackendServices() ComputeUrlMapPathMatcherPathRuleRouteActionWeightedBackendServicesList
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
	PutCorsPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicy)
	PutFaultInjectionPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicy)
	PutRequestMirrorPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicy)
	PutRetryPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicy)
	PutTimeout(value *ComputeUrlMapPathMatcherPathRuleRouteActionTimeout)
	PutUrlRewrite(value *ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewrite)
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

// The jsii proxy struct for ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference
type jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) CorsPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicyOutputReference {
	var returns ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicyOutputReference
	_jsii_.Get(
		j,
		"corsPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) CorsPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicy {
	var returns *ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicy
	_jsii_.Get(
		j,
		"corsPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) FaultInjectionPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicyOutputReference {
	var returns ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicyOutputReference
	_jsii_.Get(
		j,
		"faultInjectionPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) FaultInjectionPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicy {
	var returns *ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicy
	_jsii_.Get(
		j,
		"faultInjectionPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) InternalValue() *ComputeUrlMapPathMatcherPathRuleRouteAction {
	var returns *ComputeUrlMapPathMatcherPathRuleRouteAction
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) RequestMirrorPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicyOutputReference {
	var returns ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicyOutputReference
	_jsii_.Get(
		j,
		"requestMirrorPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) RequestMirrorPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicy {
	var returns *ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicy
	_jsii_.Get(
		j,
		"requestMirrorPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) RetryPolicy() ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicyOutputReference {
	var returns ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicyOutputReference
	_jsii_.Get(
		j,
		"retryPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) RetryPolicyInput() *ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicy {
	var returns *ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicy
	_jsii_.Get(
		j,
		"retryPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) Timeout() ComputeUrlMapPathMatcherPathRuleRouteActionTimeoutOutputReference {
	var returns ComputeUrlMapPathMatcherPathRuleRouteActionTimeoutOutputReference
	_jsii_.Get(
		j,
		"timeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) TimeoutInput() *ComputeUrlMapPathMatcherPathRuleRouteActionTimeout {
	var returns *ComputeUrlMapPathMatcherPathRuleRouteActionTimeout
	_jsii_.Get(
		j,
		"timeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) UrlRewrite() ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewriteOutputReference {
	var returns ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewriteOutputReference
	_jsii_.Get(
		j,
		"urlRewrite",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) UrlRewriteInput() *ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewrite {
	var returns *ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewrite
	_jsii_.Get(
		j,
		"urlRewriteInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) WeightedBackendServices() ComputeUrlMapPathMatcherPathRuleRouteActionWeightedBackendServicesList {
	var returns ComputeUrlMapPathMatcherPathRuleRouteActionWeightedBackendServicesList
	_jsii_.Get(
		j,
		"weightedBackendServices",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) WeightedBackendServicesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"weightedBackendServicesInput",
		&returns,
	)
	return returns
}


func NewComputeUrlMapPathMatcherPathRuleRouteActionOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference {
	_init_.Initialize()

	if err := validateNewComputeUrlMapPathMatcherPathRuleRouteActionOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeUrlMapPathMatcherPathRuleRouteActionOutputReference_Override(c ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference)SetInternalValue(val *ComputeUrlMapPathMatcherPathRuleRouteAction) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) PutCorsPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionCorsPolicy) {
	if err := c.validatePutCorsPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putCorsPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) PutFaultInjectionPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionFaultInjectionPolicy) {
	if err := c.validatePutFaultInjectionPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putFaultInjectionPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) PutRequestMirrorPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionRequestMirrorPolicy) {
	if err := c.validatePutRequestMirrorPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRequestMirrorPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) PutRetryPolicy(value *ComputeUrlMapPathMatcherPathRuleRouteActionRetryPolicy) {
	if err := c.validatePutRetryPolicyParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRetryPolicy",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) PutTimeout(value *ComputeUrlMapPathMatcherPathRuleRouteActionTimeout) {
	if err := c.validatePutTimeoutParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeout",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) PutUrlRewrite(value *ComputeUrlMapPathMatcherPathRuleRouteActionUrlRewrite) {
	if err := c.validatePutUrlRewriteParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putUrlRewrite",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) PutWeightedBackendServices(value interface{}) {
	if err := c.validatePutWeightedBackendServicesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putWeightedBackendServices",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ResetCorsPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetCorsPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ResetFaultInjectionPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetFaultInjectionPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ResetRequestMirrorPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetRequestMirrorPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ResetRetryPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetRetryPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ResetTimeout() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeout",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ResetUrlRewrite() {
	_jsii_.InvokeVoid(
		c,
		"resetUrlRewrite",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ResetWeightedBackendServices() {
	_jsii_.InvokeVoid(
		c,
		"resetWeightedBackendServices",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherPathRuleRouteActionOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

