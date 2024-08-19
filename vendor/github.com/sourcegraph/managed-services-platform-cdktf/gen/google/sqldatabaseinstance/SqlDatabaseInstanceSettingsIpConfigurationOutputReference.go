package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceSettingsIpConfigurationOutputReference interface {
	cdktf.ComplexObject
	AllocatedIpRange() *string
	SetAllocatedIpRange(val *string)
	AllocatedIpRangeInput() *string
	AuthorizedNetworks() SqlDatabaseInstanceSettingsIpConfigurationAuthorizedNetworksList
	AuthorizedNetworksInput() interface{}
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
	EnablePrivatePathForGoogleCloudServices() interface{}
	SetEnablePrivatePathForGoogleCloudServices(val interface{})
	EnablePrivatePathForGoogleCloudServicesInput() interface{}
	// Experimental.
	Fqn() *string
	InternalValue() *SqlDatabaseInstanceSettingsIpConfiguration
	SetInternalValue(val *SqlDatabaseInstanceSettingsIpConfiguration)
	Ipv4Enabled() interface{}
	SetIpv4Enabled(val interface{})
	Ipv4EnabledInput() interface{}
	PrivateNetwork() *string
	SetPrivateNetwork(val *string)
	PrivateNetworkInput() *string
	PscConfig() SqlDatabaseInstanceSettingsIpConfigurationPscConfigList
	PscConfigInput() interface{}
	RequireSsl() interface{}
	SetRequireSsl(val interface{})
	RequireSslInput() interface{}
	SslMode() *string
	SetSslMode(val *string)
	SslModeInput() *string
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
	PutAuthorizedNetworks(value interface{})
	PutPscConfig(value interface{})
	ResetAllocatedIpRange()
	ResetAuthorizedNetworks()
	ResetEnablePrivatePathForGoogleCloudServices()
	ResetIpv4Enabled()
	ResetPrivateNetwork()
	ResetPscConfig()
	ResetRequireSsl()
	ResetSslMode()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceSettingsIpConfigurationOutputReference
type jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) AllocatedIpRange() *string {
	var returns *string
	_jsii_.Get(
		j,
		"allocatedIpRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) AllocatedIpRangeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"allocatedIpRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) AuthorizedNetworks() SqlDatabaseInstanceSettingsIpConfigurationAuthorizedNetworksList {
	var returns SqlDatabaseInstanceSettingsIpConfigurationAuthorizedNetworksList
	_jsii_.Get(
		j,
		"authorizedNetworks",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) AuthorizedNetworksInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"authorizedNetworksInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) EnablePrivatePathForGoogleCloudServices() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enablePrivatePathForGoogleCloudServices",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) EnablePrivatePathForGoogleCloudServicesInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enablePrivatePathForGoogleCloudServicesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) InternalValue() *SqlDatabaseInstanceSettingsIpConfiguration {
	var returns *SqlDatabaseInstanceSettingsIpConfiguration
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) Ipv4Enabled() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"ipv4Enabled",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) Ipv4EnabledInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"ipv4EnabledInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) PrivateNetwork() *string {
	var returns *string
	_jsii_.Get(
		j,
		"privateNetwork",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) PrivateNetworkInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"privateNetworkInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) PscConfig() SqlDatabaseInstanceSettingsIpConfigurationPscConfigList {
	var returns SqlDatabaseInstanceSettingsIpConfigurationPscConfigList
	_jsii_.Get(
		j,
		"pscConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) PscConfigInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"pscConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) RequireSsl() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requireSsl",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) RequireSslInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"requireSslInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) SslMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) SslModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceSettingsIpConfigurationOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceSettingsIpConfigurationOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceSettingsIpConfigurationOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsIpConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceSettingsIpConfigurationOutputReference_Override(s SqlDatabaseInstanceSettingsIpConfigurationOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsIpConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetAllocatedIpRange(val *string) {
	if err := j.validateSetAllocatedIpRangeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"allocatedIpRange",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetEnablePrivatePathForGoogleCloudServices(val interface{}) {
	if err := j.validateSetEnablePrivatePathForGoogleCloudServicesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enablePrivatePathForGoogleCloudServices",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetInternalValue(val *SqlDatabaseInstanceSettingsIpConfiguration) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetIpv4Enabled(val interface{}) {
	if err := j.validateSetIpv4EnabledParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ipv4Enabled",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetPrivateNetwork(val *string) {
	if err := j.validateSetPrivateNetworkParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"privateNetwork",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetRequireSsl(val interface{}) {
	if err := j.validateSetRequireSslParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"requireSsl",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetSslMode(val *string) {
	if err := j.validateSetSslModeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sslMode",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := s.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		s,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := s.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := s.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		s,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := s.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		s,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := s.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		s,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := s.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		s,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := s.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		s,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := s.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		s,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := s.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		s,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := s.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) PutAuthorizedNetworks(value interface{}) {
	if err := s.validatePutAuthorizedNetworksParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putAuthorizedNetworks",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) PutPscConfig(value interface{}) {
	if err := s.validatePutPscConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putPscConfig",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetAllocatedIpRange() {
	_jsii_.InvokeVoid(
		s,
		"resetAllocatedIpRange",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetAuthorizedNetworks() {
	_jsii_.InvokeVoid(
		s,
		"resetAuthorizedNetworks",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetEnablePrivatePathForGoogleCloudServices() {
	_jsii_.InvokeVoid(
		s,
		"resetEnablePrivatePathForGoogleCloudServices",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetIpv4Enabled() {
	_jsii_.InvokeVoid(
		s,
		"resetIpv4Enabled",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetPrivateNetwork() {
	_jsii_.InvokeVoid(
		s,
		"resetPrivateNetwork",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetPscConfig() {
	_jsii_.InvokeVoid(
		s,
		"resetPscConfig",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetRequireSsl() {
	_jsii_.InvokeVoid(
		s,
		"resetRequireSsl",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ResetSslMode() {
	_jsii_.InvokeVoid(
		s,
		"resetSslMode",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := s.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		s,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsIpConfigurationOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

