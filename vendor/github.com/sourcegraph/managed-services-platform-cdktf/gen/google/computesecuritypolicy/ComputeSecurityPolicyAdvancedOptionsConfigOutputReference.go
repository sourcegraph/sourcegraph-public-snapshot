package computesecuritypolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesecuritypolicy/internal"
)

type ComputeSecurityPolicyAdvancedOptionsConfigOutputReference interface {
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
	InternalValue() *ComputeSecurityPolicyAdvancedOptionsConfig
	SetInternalValue(val *ComputeSecurityPolicyAdvancedOptionsConfig)
	JsonCustomConfig() ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfigOutputReference
	JsonCustomConfigInput() *ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfig
	JsonParsing() *string
	SetJsonParsing(val *string)
	JsonParsingInput() *string
	LogLevel() *string
	SetLogLevel(val *string)
	LogLevelInput() *string
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	UserIpRequestHeaders() *[]*string
	SetUserIpRequestHeaders(val *[]*string)
	UserIpRequestHeadersInput() *[]*string
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
	PutJsonCustomConfig(value *ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfig)
	ResetJsonCustomConfig()
	ResetJsonParsing()
	ResetLogLevel()
	ResetUserIpRequestHeaders()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for ComputeSecurityPolicyAdvancedOptionsConfigOutputReference
type jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) InternalValue() *ComputeSecurityPolicyAdvancedOptionsConfig {
	var returns *ComputeSecurityPolicyAdvancedOptionsConfig
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) JsonCustomConfig() ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfigOutputReference {
	var returns ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfigOutputReference
	_jsii_.Get(
		j,
		"jsonCustomConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) JsonCustomConfigInput() *ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfig {
	var returns *ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfig
	_jsii_.Get(
		j,
		"jsonCustomConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) JsonParsing() *string {
	var returns *string
	_jsii_.Get(
		j,
		"jsonParsing",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) JsonParsingInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"jsonParsingInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) LogLevel() *string {
	var returns *string
	_jsii_.Get(
		j,
		"logLevel",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) LogLevelInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"logLevelInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) UserIpRequestHeaders() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"userIpRequestHeaders",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) UserIpRequestHeadersInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"userIpRequestHeadersInput",
		&returns,
	)
	return returns
}


func NewComputeSecurityPolicyAdvancedOptionsConfigOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) ComputeSecurityPolicyAdvancedOptionsConfigOutputReference {
	_init_.Initialize()

	if err := validateNewComputeSecurityPolicyAdvancedOptionsConfigOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.computeSecurityPolicy.ComputeSecurityPolicyAdvancedOptionsConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewComputeSecurityPolicyAdvancedOptionsConfigOutputReference_Override(c ComputeSecurityPolicyAdvancedOptionsConfigOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeSecurityPolicy.ComputeSecurityPolicyAdvancedOptionsConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		c,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetInternalValue(val *ComputeSecurityPolicyAdvancedOptionsConfig) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetJsonParsing(val *string) {
	if err := j.validateSetJsonParsingParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"jsonParsing",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetLogLevel(val *string) {
	if err := j.validateSetLogLevelParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"logLevel",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference)SetUserIpRequestHeaders(val *[]*string) {
	if err := j.validateSetUserIpRequestHeadersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"userIpRequestHeaders",
		val,
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) PutJsonCustomConfig(value *ComputeSecurityPolicyAdvancedOptionsConfigJsonCustomConfig) {
	if err := c.validatePutJsonCustomConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putJsonCustomConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ResetJsonCustomConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetJsonCustomConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ResetJsonParsing() {
	_jsii_.InvokeVoid(
		c,
		"resetJsonParsing",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ResetLogLevel() {
	_jsii_.InvokeVoid(
		c,
		"resetLogLevel",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ResetUserIpRequestHeaders() {
	_jsii_.InvokeVoid(
		c,
		"resetUserIpRequestHeaders",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (c *jsiiProxy_ComputeSecurityPolicyAdvancedOptionsConfigOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

