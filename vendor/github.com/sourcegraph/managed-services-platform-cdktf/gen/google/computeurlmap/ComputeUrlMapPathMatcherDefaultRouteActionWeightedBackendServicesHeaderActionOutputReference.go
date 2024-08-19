package computeurlmap

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computeurlmap/internal"
)

type ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference interface {
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
	InternalValue() *ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderAction
	SetInternalValue(val *ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderAction)
	RequestHeadersToAdd() ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionRequestHeadersToAddList
	RequestHeadersToAddInput() interface{}
	RequestHeadersToRemove() *[]*string
	SetRequestHeadersToRemove(val *[]*string)
	RequestHeadersToRemoveInput() *[]*string
	ResponseHeadersToAdd() ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionResponseHeadersToAddList
	ResponseHeadersToAddInput() interface{}
	ResponseHeadersToRemove() *[]*string
	SetResponseHeadersToRemove(val *[]*string)
	ResponseHeadersToRemoveInput() *[]*string
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
	PutRequestHeadersToAdd(value interface{})
	PutResponseHeadersToAdd(value interface{})
	ResetRequestHeadersToAdd()
	ResetRequestHeadersToRemove()
	ResetResponseHeadersToAdd()
	ResetResponseHeadersToRemove()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference
type jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) InternalValue() *ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderAction {
	var returns *ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderAction
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) RequestHeadersToAdd() ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionRequestHeadersToAddList {
	var returns ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionRequestHeadersToAddList
	_jsii_.Get(
		j,
		"requestHeadersToAdd",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) RequestHeadersToAddInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requestHeadersToAddInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) RequestHeadersToRemove() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"requestHeadersToRemove",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) RequestHeadersToRemoveInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"requestHeadersToRemoveInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResponseHeadersToAdd() ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionResponseHeadersToAddList {
	var returns ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionResponseHeadersToAddList
	_jsii_.Get(
		j,
		"responseHeadersToAdd",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResponseHeadersToAddInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"responseHeadersToAddInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResponseHeadersToRemove() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"responseHeadersToRemove",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResponseHeadersToRemoveInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"responseHeadersToRemoveInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference {
	_init_.Initialize()

	if err := validateNewComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference_Override(c ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeUrlMap.ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference)SetInternalValue(val *ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderAction) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference)SetRequestHeadersToRemove(val *[]*string) {
	if err := j.validateSetRequestHeadersToRemoveParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"requestHeadersToRemove",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference)SetResponseHeadersToRemove(val *[]*string) {
	if err := j.validateSetResponseHeadersToRemoveParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"responseHeadersToRemove",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) PutRequestHeadersToAdd(value interface{}) {
	if err := c.validatePutRequestHeadersToAddParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putRequestHeadersToAdd",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) PutResponseHeadersToAdd(value interface{}) {
	if err := c.validatePutResponseHeadersToAddParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putResponseHeadersToAdd",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResetRequestHeadersToAdd() {
	_jsii_.InvokeVoid(
		c,
		"resetRequestHeadersToAdd",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResetRequestHeadersToRemove() {
	_jsii_.InvokeVoid(
		c,
		"resetRequestHeadersToRemove",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResetResponseHeadersToAdd() {
	_jsii_.InvokeVoid(
		c,
		"resetResponseHeadersToAdd",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ResetResponseHeadersToRemove() {
	_jsii_.InvokeVoid(
		c,
		"resetResponseHeadersToRemove",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeUrlMapPathMatcherDefaultRouteActionWeightedBackendServicesHeaderActionOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

