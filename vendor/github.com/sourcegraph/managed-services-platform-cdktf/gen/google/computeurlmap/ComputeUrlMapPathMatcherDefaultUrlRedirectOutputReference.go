package computeurlmap

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap/internal"
)

type ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference interface {
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
	HostRedirect() *string
	SetHostRedirect(val *string)
	HostRedirectInput() *string
	HttpsRedirect() interface{}
	SetHttpsRedirect(val interface{})
	HttpsRedirectInput() interface{}
	InternalValue() *ComputeUrlMapPathMatcherDefaultUrlRedirect
	SetInternalValue(val *ComputeUrlMapPathMatcherDefaultUrlRedirect)
	PathRedirect() *string
	SetPathRedirect(val *string)
	PathRedirectInput() *string
	PrefixRedirect() *string
	SetPrefixRedirect(val *string)
	PrefixRedirectInput() *string
	RedirectResponseCode() *string
	SetRedirectResponseCode(val *string)
	RedirectResponseCodeInput() *string
	StripQuery() interface{}
	SetStripQuery(val interface{})
	StripQueryInput() interface{}
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
	ResetHostRedirect()
	ResetHttpsRedirect()
	ResetPathRedirect()
	ResetPrefixRedirect()
	ResetRedirectResponseCode()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference
type jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) HostRedirect() *string {
	var returns *string
	_jsii_.Get(
		j,
		"hostRedirect",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) HostRedirectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"hostRedirectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) HttpsRedirect() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"httpsRedirect",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) HttpsRedirectInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"httpsRedirectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) InternalValue() *ComputeUrlMapPathMatcherDefaultUrlRedirect {
	var returns *ComputeUrlMapPathMatcherDefaultUrlRedirect
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) PathRedirect() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pathRedirect",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) PathRedirectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pathRedirectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) PrefixRedirect() *string {
	var returns *string
	_jsii_.Get(
		j,
		"prefixRedirect",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) PrefixRedirectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"prefixRedirectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) RedirectResponseCode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"redirectResponseCode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) RedirectResponseCodeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"redirectResponseCodeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) StripQuery() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"stripQuery",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) StripQueryInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"stripQueryInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference {
	_init_.Initialize()

	if err := validateNewComputeUrlMapPathMatcherDefaultUrlRedirectOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference_Override(c ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetHostRedirect(val *string) {
	if err := j.validateSetHostRedirectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"hostRedirect",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetHttpsRedirect(val interface{}) {
	if err := j.validateSetHttpsRedirectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"httpsRedirect",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetInternalValue(val *ComputeUrlMapPathMatcherDefaultUrlRedirect) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetPathRedirect(val *string) {
	if err := j.validateSetPathRedirectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"pathRedirect",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetPrefixRedirect(val *string) {
	if err := j.validateSetPrefixRedirectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"prefixRedirect",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetRedirectResponseCode(val *string) {
	if err := j.validateSetRedirectResponseCodeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"redirectResponseCode",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetStripQuery(val interface{}) {
	if err := j.validateSetStripQueryParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"stripQuery",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ResetHostRedirect() {
	_jsii_.InvokeVoid(
		c,
		"resetHostRedirect",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ResetHttpsRedirect() {
	_jsii_.InvokeVoid(
		c,
		"resetHttpsRedirect",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ResetPathRedirect() {
	_jsii_.InvokeVoid(
		c,
		"resetPathRedirect",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ResetPrefixRedirect() {
	_jsii_.InvokeVoid(
		c,
		"resetPrefixRedirect",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ResetRedirectResponseCode() {
	_jsii_.InvokeVoid(
		c,
		"resetRedirectResponseCode",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultUrlRedirectOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

