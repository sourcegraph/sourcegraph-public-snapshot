package computebackendservice

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computebackendservice/internal"
)

type ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference interface {
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
	IncludeHost() interface{}
	SetIncludeHost(val interface{})
	IncludeHostInput() interface{}
	IncludeHttpHeaders() *[]*string
	SetIncludeHttpHeaders(val *[]*string)
	IncludeHttpHeadersInput() *[]*string
	IncludeNamedCookies() *[]*string
	SetIncludeNamedCookies(val *[]*string)
	IncludeNamedCookiesInput() *[]*string
	IncludeProtocol() interface{}
	SetIncludeProtocol(val interface{})
	IncludeProtocolInput() interface{}
	IncludeQueryString() interface{}
	SetIncludeQueryString(val interface{})
	IncludeQueryStringInput() interface{}
	InternalValue() *ComputeBackendServiceCdnPolicyCacheKeyPolicy
	SetInternalValue(val *ComputeBackendServiceCdnPolicyCacheKeyPolicy)
	QueryStringBlacklist() *[]*string
	SetQueryStringBlacklist(val *[]*string)
	QueryStringBlacklistInput() *[]*string
	QueryStringWhitelist() *[]*string
	SetQueryStringWhitelist(val *[]*string)
	QueryStringWhitelistInput() *[]*string
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
	ResetIncludeHost()
	ResetIncludeHttpHeaders()
	ResetIncludeNamedCookies()
	ResetIncludeProtocol()
	ResetIncludeQueryString()
	ResetQueryStringBlacklist()
	ResetQueryStringWhitelist()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference
type jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeHost() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"includeHost",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeHostInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"includeHostInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeHttpHeaders() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"includeHttpHeaders",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeHttpHeadersInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"includeHttpHeadersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeNamedCookies() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"includeNamedCookies",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeNamedCookiesInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"includeNamedCookiesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeProtocol() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"includeProtocol",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeProtocolInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"includeProtocolInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeQueryString() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"includeQueryString",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) IncludeQueryStringInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"includeQueryStringInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) InternalValue() *ComputeBackendServiceCdnPolicyCacheKeyPolicy {
	var returns *ComputeBackendServiceCdnPolicyCacheKeyPolicy
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) QueryStringBlacklist() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"queryStringBlacklist",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) QueryStringBlacklistInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"queryStringBlacklistInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) QueryStringWhitelist() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"queryStringWhitelist",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) QueryStringWhitelistInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"queryStringWhitelistInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference {
	_init_.Initialize()

	if err := validateNewComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference_Override(c ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeBackendService.ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetIncludeHost(val interface{}) {
	if err := j.validateSetIncludeHostParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"includeHost",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetIncludeHttpHeaders(val *[]*string) {
	if err := j.validateSetIncludeHttpHeadersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"includeHttpHeaders",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetIncludeNamedCookies(val *[]*string) {
	if err := j.validateSetIncludeNamedCookiesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"includeNamedCookies",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetIncludeProtocol(val interface{}) {
	if err := j.validateSetIncludeProtocolParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"includeProtocol",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetIncludeQueryString(val interface{}) {
	if err := j.validateSetIncludeQueryStringParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"includeQueryString",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetInternalValue(val *ComputeBackendServiceCdnPolicyCacheKeyPolicy) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetQueryStringBlacklist(val *[]*string) {
	if err := j.validateSetQueryStringBlacklistParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"queryStringBlacklist",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetQueryStringWhitelist(val *[]*string) {
	if err := j.validateSetQueryStringWhitelistParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"queryStringWhitelist",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ResetIncludeHost() {
	_jsii_.InvokeVoid(
		c,
		"resetIncludeHost",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ResetIncludeHttpHeaders() {
	_jsii_.InvokeVoid(
		c,
		"resetIncludeHttpHeaders",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ResetIncludeNamedCookies() {
	_jsii_.InvokeVoid(
		c,
		"resetIncludeNamedCookies",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ResetIncludeProtocol() {
	_jsii_.InvokeVoid(
		c,
		"resetIncludeProtocol",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ResetIncludeQueryString() {
	_jsii_.InvokeVoid(
		c,
		"resetIncludeQueryString",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ResetQueryStringBlacklist() {
	_jsii_.InvokeVoid(
		c,
		"resetQueryStringBlacklist",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ResetQueryStringWhitelist() {
	_jsii_.InvokeVoid(
		c,
		"resetQueryStringWhitelist",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeBackendServiceCdnPolicyCacheKeyPolicyOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

