package monitoringuptimecheckconfig

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringuptimecheckconfig/internal"
)

type MonitoringUptimeCheckConfigHttpCheckOutputReference interface {
	cdktf.ComplexObject
	AcceptedResponseStatusCodes() MonitoringUptimeCheckConfigHttpCheckAcceptedResponseStatusCodesList
	AcceptedResponseStatusCodesInput() interface{}
	AuthInfo() MonitoringUptimeCheckConfigHttpCheckAuthInfoOutputReference
	AuthInfoInput() *MonitoringUptimeCheckConfigHttpCheckAuthInfo
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
	ContentType() *string
	SetContentType(val *string)
	ContentTypeInput() *string
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	CustomContentType() *string
	SetCustomContentType(val *string)
	CustomContentTypeInput() *string
	// Experimental.
	Fqn() *string
	Headers() *map[string]*string
	SetHeaders(val *map[string]*string)
	HeadersInput() *map[string]*string
	InternalValue() *MonitoringUptimeCheckConfigHttpCheck
	SetInternalValue(val *MonitoringUptimeCheckConfigHttpCheck)
	MaskHeaders() interface{}
	SetMaskHeaders(val interface{})
	MaskHeadersInput() interface{}
	Path() *string
	SetPath(val *string)
	PathInput() *string
	PingConfig() MonitoringUptimeCheckConfigHttpCheckPingConfigOutputReference
	PingConfigInput() *MonitoringUptimeCheckConfigHttpCheckPingConfig
	Port() *float64
	SetPort(val *float64)
	PortInput() *float64
	RequestMethod() *string
	SetRequestMethod(val *string)
	RequestMethodInput() *string
	ServiceAgentAuthentication() MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthenticationOutputReference
	ServiceAgentAuthenticationInput() *MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthentication
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	UseSsl() interface{}
	SetUseSsl(val interface{})
	UseSslInput() interface{}
	ValidateSsl() interface{}
	SetValidateSsl(val interface{})
	ValidateSslInput() interface{}
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
	PutAcceptedResponseStatusCodes(value interface{})
	PutAuthInfo(value *MonitoringUptimeCheckConfigHttpCheckAuthInfo)
	PutPingConfig(value *MonitoringUptimeCheckConfigHttpCheckPingConfig)
	PutServiceAgentAuthentication(value *MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthentication)
	ResetAcceptedResponseStatusCodes()
	ResetAuthInfo()
	ResetBody()
	ResetContentType()
	ResetCustomContentType()
	ResetHeaders()
	ResetMaskHeaders()
	ResetPath()
	ResetPingConfig()
	ResetPort()
	ResetRequestMethod()
	ResetServiceAgentAuthentication()
	ResetUseSsl()
	ResetValidateSsl()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for MonitoringUptimeCheckConfigHttpCheckOutputReference
type jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) AcceptedResponseStatusCodes() MonitoringUptimeCheckConfigHttpCheckAcceptedResponseStatusCodesList {
	var returns MonitoringUptimeCheckConfigHttpCheckAcceptedResponseStatusCodesList
	_jsii_.Get(
		j,
		"acceptedResponseStatusCodes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) AcceptedResponseStatusCodesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"acceptedResponseStatusCodesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) AuthInfo() MonitoringUptimeCheckConfigHttpCheckAuthInfoOutputReference {
	var returns MonitoringUptimeCheckConfigHttpCheckAuthInfoOutputReference
	_jsii_.Get(
		j,
		"authInfo",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) AuthInfoInput() *MonitoringUptimeCheckConfigHttpCheckAuthInfo {
	var returns *MonitoringUptimeCheckConfigHttpCheckAuthInfo
	_jsii_.Get(
		j,
		"authInfoInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) Body() *string {
	var returns *string
	_jsii_.Get(
		j,
		"body",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) BodyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"bodyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ContentType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"contentType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ContentTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"contentTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) CustomContentType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"customContentType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) CustomContentTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"customContentTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) Headers() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"headers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) HeadersInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"headersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) InternalValue() *MonitoringUptimeCheckConfigHttpCheck {
	var returns *MonitoringUptimeCheckConfigHttpCheck
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) MaskHeaders() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"maskHeaders",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) MaskHeadersInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"maskHeadersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) Path() *string {
	var returns *string
	_jsii_.Get(
		j,
		"path",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PathInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pathInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PingConfig() MonitoringUptimeCheckConfigHttpCheckPingConfigOutputReference {
	var returns MonitoringUptimeCheckConfigHttpCheckPingConfigOutputReference
	_jsii_.Get(
		j,
		"pingConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PingConfigInput() *MonitoringUptimeCheckConfigHttpCheckPingConfig {
	var returns *MonitoringUptimeCheckConfigHttpCheckPingConfig
	_jsii_.Get(
		j,
		"pingConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) Port() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"port",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PortInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"portInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) RequestMethod() *string {
	var returns *string
	_jsii_.Get(
		j,
		"requestMethod",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) RequestMethodInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"requestMethodInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ServiceAgentAuthentication() MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthenticationOutputReference {
	var returns MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthenticationOutputReference
	_jsii_.Get(
		j,
		"serviceAgentAuthentication",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ServiceAgentAuthenticationInput() *MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthentication {
	var returns *MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthentication
	_jsii_.Get(
		j,
		"serviceAgentAuthenticationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) UseSsl() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"useSsl",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) UseSslInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"useSslInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ValidateSsl() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"validateSsl",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ValidateSslInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"validateSslInput",
		&returns,
	)
	return returns
}


func NewMonitoringUptimeCheckConfigHttpCheckOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) MonitoringUptimeCheckConfigHttpCheckOutputReference {
	_init_.Initialize()

	if err := validateNewMonitoringUptimeCheckConfigHttpCheckOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfigHttpCheckOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewMonitoringUptimeCheckConfigHttpCheckOutputReference_Override(m MonitoringUptimeCheckConfigHttpCheckOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.monitoringUptimeCheckConfig.MonitoringUptimeCheckConfigHttpCheckOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		m,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetBody(val *string) {
	if err := j.validateSetBodyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"body",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetContentType(val *string) {
	if err := j.validateSetContentTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"contentType",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetCustomContentType(val *string) {
	if err := j.validateSetCustomContentTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"customContentType",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetHeaders(val *map[string]*string) {
	if err := j.validateSetHeadersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"headers",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetInternalValue(val *MonitoringUptimeCheckConfigHttpCheck) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetMaskHeaders(val interface{}) {
	if err := j.validateSetMaskHeadersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"maskHeaders",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetPath(val *string) {
	if err := j.validateSetPathParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"path",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetPort(val *float64) {
	if err := j.validateSetPortParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"port",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetRequestMethod(val *string) {
	if err := j.validateSetRequestMethodParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"requestMethod",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetUseSsl(val interface{}) {
	if err := j.validateSetUseSslParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"useSsl",
		val,
	)
}

func (j *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference)SetValidateSsl(val interface{}) {
	if err := j.validateSetValidateSslParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"validateSsl",
		val,
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := m.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		m,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := m.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := m.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		m,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := m.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		m,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := m.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		m,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := m.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		m,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := m.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		m,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := m.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		m,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := m.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		m,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := m.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PutAcceptedResponseStatusCodes(value interface{}) {
	if err := m.validatePutAcceptedResponseStatusCodesParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putAcceptedResponseStatusCodes",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PutAuthInfo(value *MonitoringUptimeCheckConfigHttpCheckAuthInfo) {
	if err := m.validatePutAuthInfoParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putAuthInfo",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PutPingConfig(value *MonitoringUptimeCheckConfigHttpCheckPingConfig) {
	if err := m.validatePutPingConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putPingConfig",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) PutServiceAgentAuthentication(value *MonitoringUptimeCheckConfigHttpCheckServiceAgentAuthentication) {
	if err := m.validatePutServiceAgentAuthenticationParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putServiceAgentAuthentication",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetAcceptedResponseStatusCodes() {
	_jsii_.InvokeVoid(
		m,
		"resetAcceptedResponseStatusCodes",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetAuthInfo() {
	_jsii_.InvokeVoid(
		m,
		"resetAuthInfo",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetBody() {
	_jsii_.InvokeVoid(
		m,
		"resetBody",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetContentType() {
	_jsii_.InvokeVoid(
		m,
		"resetContentType",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetCustomContentType() {
	_jsii_.InvokeVoid(
		m,
		"resetCustomContentType",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetHeaders() {
	_jsii_.InvokeVoid(
		m,
		"resetHeaders",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetMaskHeaders() {
	_jsii_.InvokeVoid(
		m,
		"resetMaskHeaders",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetPath() {
	_jsii_.InvokeVoid(
		m,
		"resetPath",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetPingConfig() {
	_jsii_.InvokeVoid(
		m,
		"resetPingConfig",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetPort() {
	_jsii_.InvokeVoid(
		m,
		"resetPort",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetRequestMethod() {
	_jsii_.InvokeVoid(
		m,
		"resetRequestMethod",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetServiceAgentAuthentication() {
	_jsii_.InvokeVoid(
		m,
		"resetServiceAgentAuthentication",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetUseSsl() {
	_jsii_.InvokeVoid(
		m,
		"resetUseSsl",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ResetValidateSsl() {
	_jsii_.InvokeVoid(
		m,
		"resetValidateSsl",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := m.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		m,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringUptimeCheckConfigHttpCheckOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

