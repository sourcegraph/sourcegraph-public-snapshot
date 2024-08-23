package pubsubsubscription

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/pubsubsubscription/internal"
)

type PubsubSubscriptionPushConfigOutputReference interface {
	cdktf.ComplexObject
	Attributes() *map[string]*string
	SetAttributes(val *map[string]*string)
	AttributesInput() *map[string]*string
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
	InternalValue() *PubsubSubscriptionPushConfig
	SetInternalValue(val *PubsubSubscriptionPushConfig)
	NoWrapper() PubsubSubscriptionPushConfigNoWrapperOutputReference
	NoWrapperInput() *PubsubSubscriptionPushConfigNoWrapper
	OidcToken() PubsubSubscriptionPushConfigOidcTokenOutputReference
	OidcTokenInput() *PubsubSubscriptionPushConfigOidcToken
	PushEndpoint() *string
	SetPushEndpoint(val *string)
	PushEndpointInput() *string
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
	PutNoWrapper(value *PubsubSubscriptionPushConfigNoWrapper)
	PutOidcToken(value *PubsubSubscriptionPushConfigOidcToken)
	ResetAttributes()
	ResetNoWrapper()
	ResetOidcToken()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for PubsubSubscriptionPushConfigOutputReference
type jsiiProxy_PubsubSubscriptionPushConfigOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) Attributes() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"attributes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) AttributesInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"attributesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) InternalValue() *PubsubSubscriptionPushConfig {
	var returns *PubsubSubscriptionPushConfig
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) NoWrapper() PubsubSubscriptionPushConfigNoWrapperOutputReference {
	var returns PubsubSubscriptionPushConfigNoWrapperOutputReference
	_jsii_.Get(
		j,
		"noWrapper",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) NoWrapperInput() *PubsubSubscriptionPushConfigNoWrapper {
	var returns *PubsubSubscriptionPushConfigNoWrapper
	_jsii_.Get(
		j,
		"noWrapperInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) OidcToken() PubsubSubscriptionPushConfigOidcTokenOutputReference {
	var returns PubsubSubscriptionPushConfigOidcTokenOutputReference
	_jsii_.Get(
		j,
		"oidcToken",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) OidcTokenInput() *PubsubSubscriptionPushConfigOidcToken {
	var returns *PubsubSubscriptionPushConfigOidcToken
	_jsii_.Get(
		j,
		"oidcTokenInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) PushEndpoint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pushEndpoint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) PushEndpointInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pushEndpointInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewPubsubSubscriptionPushConfigOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) PubsubSubscriptionPushConfigOutputReference {
	_init_.Initialize()

	if err := validateNewPubsubSubscriptionPushConfigOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_PubsubSubscriptionPushConfigOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscriptionPushConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewPubsubSubscriptionPushConfigOutputReference_Override(p PubsubSubscriptionPushConfigOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.pubsubSubscription.PubsubSubscriptionPushConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		p,
	)
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference)SetAttributes(val *map[string]*string) {
	if err := j.validateSetAttributesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"attributes",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference)SetInternalValue(val *PubsubSubscriptionPushConfig) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference)SetPushEndpoint(val *string) {
	if err := j.validateSetPushEndpointParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"pushEndpoint",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_PubsubSubscriptionPushConfigOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		p,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := p.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := p.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := p.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		p,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := p.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		p,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := p.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		p,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := p.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		p,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := p.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		p,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := p.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		p,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := p.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		p,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := p.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) PutNoWrapper(value *PubsubSubscriptionPushConfigNoWrapper) {
	if err := p.validatePutNoWrapperParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putNoWrapper",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) PutOidcToken(value *PubsubSubscriptionPushConfigOidcToken) {
	if err := p.validatePutOidcTokenParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"putOidcToken",
		[]interface{}{value},
	)
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) ResetAttributes() {
	_jsii_.InvokeVoid(
		p,
		"resetAttributes",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) ResetNoWrapper() {
	_jsii_.InvokeVoid(
		p,
		"resetNoWrapper",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) ResetOidcToken() {
	_jsii_.InvokeVoid(
		p,
		"resetOidcToken",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := p.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		p,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PubsubSubscriptionPushConfigOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		p,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

