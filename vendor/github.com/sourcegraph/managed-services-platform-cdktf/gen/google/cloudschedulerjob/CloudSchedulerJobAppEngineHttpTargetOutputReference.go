package cloudschedulerjob

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/cloudschedulerjob/internal"
)

type CloudSchedulerJobAppEngineHttpTargetOutputReference interface {
	cdktf.ComplexObject
	AppEngineRouting() CloudSchedulerJobAppEngineHttpTargetAppEngineRoutingOutputReference
	AppEngineRoutingInput() *CloudSchedulerJobAppEngineHttpTargetAppEngineRouting
	Body() *string
	SetBody(val *string)
	BodyInput() *string
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
	Headers() *map[string]*string
	SetHeaders(val *map[string]*string)
	HeadersInput() *map[string]*string
	HttpMethod() *string
	SetHttpMethod(val *string)
	HttpMethodInput() *string
	InternalValue() *CloudSchedulerJobAppEngineHttpTarget
	SetInternalValue(val *CloudSchedulerJobAppEngineHttpTarget)
	RelativeUri() *string
	SetRelativeUri(val *string)
	RelativeUriInput() *string
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
	PutAppEngineRouting(value *CloudSchedulerJobAppEngineHttpTargetAppEngineRouting)
	ResetAppEngineRouting()
	ResetBody()
	ResetHeaders()
	ResetHttpMethod()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for CloudSchedulerJobAppEngineHttpTargetOutputReference
type jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) AppEngineRouting() CloudSchedulerJobAppEngineHttpTargetAppEngineRoutingOutputReference {
	var returns CloudSchedulerJobAppEngineHttpTargetAppEngineRoutingOutputReference
	_jsii_.Get(
		j,
		"appEngineRouting",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) AppEngineRoutingInput() *CloudSchedulerJobAppEngineHttpTargetAppEngineRouting {
	var returns *CloudSchedulerJobAppEngineHttpTargetAppEngineRouting
	_jsii_.Get(
		j,
		"appEngineRoutingInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) Body() *string {
	var returns *string
	_jsii_.Get(
		j,
		"body",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) BodyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"bodyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) Headers() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"headers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) HeadersInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"headersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) HttpMethod() *string {
	var returns *string
	_jsii_.Get(
		j,
		"httpMethod",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) HttpMethodInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"httpMethodInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) InternalValue() *CloudSchedulerJobAppEngineHttpTarget {
	var returns *CloudSchedulerJobAppEngineHttpTarget
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) RelativeUri() *string {
	var returns *string
	_jsii_.Get(
		j,
		"relativeUri",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) RelativeUriInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"relativeUriInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewCloudSchedulerJobAppEngineHttpTargetOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) CloudSchedulerJobAppEngineHttpTargetOutputReference {
	_init_.Initialize()

	if err := validateNewCloudSchedulerJobAppEngineHttpTargetOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJobAppEngineHttpTargetOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewCloudSchedulerJobAppEngineHttpTargetOutputReference_Override(c CloudSchedulerJobAppEngineHttpTargetOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.cloudSchedulerJob.CloudSchedulerJobAppEngineHttpTargetOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetBody(val *string) {
	if err := j.validateSetBodyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"body",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetHeaders(val *map[string]*string) {
	if err := j.validateSetHeadersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"headers",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetHttpMethod(val *string) {
	if err := j.validateSetHttpMethodParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"httpMethod",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetInternalValue(val *CloudSchedulerJobAppEngineHttpTarget) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetRelativeUri(val *string) {
	if err := j.validateSetRelativeUriParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"relativeUri",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) PutAppEngineRouting(value *CloudSchedulerJobAppEngineHttpTargetAppEngineRouting) {
	if err := c.validatePutAppEngineRoutingParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putAppEngineRouting",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ResetAppEngineRouting() {
	_jsii_.InvokeVoid(
		c,
		"resetAppEngineRouting",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ResetBody() {
	_jsii_.InvokeVoid(
		c,
		"resetBody",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ResetHeaders() {
	_jsii_.InvokeVoid(
		c,
		"resetHeaders",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ResetHttpMethod() {
	_jsii_.InvokeVoid(
		c,
		"resetHttpMethod",
		nil, // no parameters
	)
}

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_CloudSchedulerJobAppEngineHttpTargetOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

