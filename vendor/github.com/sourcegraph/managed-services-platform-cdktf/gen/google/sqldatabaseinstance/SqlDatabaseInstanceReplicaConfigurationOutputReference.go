package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceReplicaConfigurationOutputReference interface {
	cdktf.ComplexObject
	CaCertificate() *string
	SetCaCertificate(val *string)
	CaCertificateInput() *string
	ClientCertificate() *string
	SetClientCertificate(val *string)
	ClientCertificateInput() *string
	ClientKey() *string
	SetClientKey(val *string)
	ClientKeyInput() *string
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
	ConnectRetryInterval() *float64
	SetConnectRetryInterval(val *float64)
	ConnectRetryIntervalInput() *float64
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	DumpFilePath() *string
	SetDumpFilePath(val *string)
	DumpFilePathInput() *string
	FailoverTarget() interface{}
	SetFailoverTarget(val interface{})
	FailoverTargetInput() interface{}
	// Experimental.
	Fqn() *string
	InternalValue() *SqlDatabaseInstanceReplicaConfiguration
	SetInternalValue(val *SqlDatabaseInstanceReplicaConfiguration)
	MasterHeartbeatPeriod() *float64
	SetMasterHeartbeatPeriod(val *float64)
	MasterHeartbeatPeriodInput() *float64
	Password() *string
	SetPassword(val *string)
	PasswordInput() *string
	SslCipher() *string
	SetSslCipher(val *string)
	SslCipherInput() *string
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	Username() *string
	SetUsername(val *string)
	UsernameInput() *string
	VerifyServerCertificate() interface{}
	SetVerifyServerCertificate(val interface{})
	VerifyServerCertificateInput() interface{}
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
	ResetCaCertificate()
	ResetClientCertificate()
	ResetClientKey()
	ResetConnectRetryInterval()
	ResetDumpFilePath()
	ResetFailoverTarget()
	ResetMasterHeartbeatPeriod()
	ResetPassword()
	ResetSslCipher()
	ResetUsername()
	ResetVerifyServerCertificate()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceReplicaConfigurationOutputReference
type jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) CaCertificate() *string {
	var returns *string
	_jsii_.Get(
		j,
		"caCertificate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) CaCertificateInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"caCertificateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ClientCertificate() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientCertificate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ClientCertificateInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientCertificateInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ClientKey() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientKey",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ClientKeyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientKeyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ConnectRetryInterval() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectRetryInterval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ConnectRetryIntervalInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectRetryIntervalInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) DumpFilePath() *string {
	var returns *string
	_jsii_.Get(
		j,
		"dumpFilePath",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) DumpFilePathInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"dumpFilePathInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) FailoverTarget() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"failoverTarget",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) FailoverTargetInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"failoverTargetInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) InternalValue() *SqlDatabaseInstanceReplicaConfiguration {
	var returns *SqlDatabaseInstanceReplicaConfiguration
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) MasterHeartbeatPeriod() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"masterHeartbeatPeriod",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) MasterHeartbeatPeriodInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"masterHeartbeatPeriodInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) Password() *string {
	var returns *string
	_jsii_.Get(
		j,
		"password",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) PasswordInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"passwordInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) SslCipher() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslCipher",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) SslCipherInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslCipherInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) Username() *string {
	var returns *string
	_jsii_.Get(
		j,
		"username",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) UsernameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"usernameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) VerifyServerCertificate() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"verifyServerCertificate",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) VerifyServerCertificateInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"verifyServerCertificateInput",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceReplicaConfigurationOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceReplicaConfigurationOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceReplicaConfigurationOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceReplicaConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceReplicaConfigurationOutputReference_Override(s SqlDatabaseInstanceReplicaConfigurationOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceReplicaConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetCaCertificate(val *string) {
	if err := j.validateSetCaCertificateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"caCertificate",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetClientCertificate(val *string) {
	if err := j.validateSetClientCertificateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"clientCertificate",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetClientKey(val *string) {
	if err := j.validateSetClientKeyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"clientKey",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetConnectRetryInterval(val *float64) {
	if err := j.validateSetConnectRetryIntervalParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connectRetryInterval",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetDumpFilePath(val *string) {
	if err := j.validateSetDumpFilePathParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"dumpFilePath",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetFailoverTarget(val interface{}) {
	if err := j.validateSetFailoverTargetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"failoverTarget",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetInternalValue(val *SqlDatabaseInstanceReplicaConfiguration) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetMasterHeartbeatPeriod(val *float64) {
	if err := j.validateSetMasterHeartbeatPeriodParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"masterHeartbeatPeriod",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetPassword(val *string) {
	if err := j.validateSetPasswordParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"password",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetSslCipher(val *string) {
	if err := j.validateSetSslCipherParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sslCipher",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetUsername(val *string) {
	if err := j.validateSetUsernameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"username",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference)SetVerifyServerCertificate(val interface{}) {
	if err := j.validateSetVerifyServerCertificateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"verifyServerCertificate",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetCaCertificate() {
	_jsii_.InvokeVoid(
		s,
		"resetCaCertificate",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetClientCertificate() {
	_jsii_.InvokeVoid(
		s,
		"resetClientCertificate",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetClientKey() {
	_jsii_.InvokeVoid(
		s,
		"resetClientKey",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetConnectRetryInterval() {
	_jsii_.InvokeVoid(
		s,
		"resetConnectRetryInterval",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetDumpFilePath() {
	_jsii_.InvokeVoid(
		s,
		"resetDumpFilePath",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetFailoverTarget() {
	_jsii_.InvokeVoid(
		s,
		"resetFailoverTarget",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetMasterHeartbeatPeriod() {
	_jsii_.InvokeVoid(
		s,
		"resetMasterHeartbeatPeriod",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetPassword() {
	_jsii_.InvokeVoid(
		s,
		"resetPassword",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetSslCipher() {
	_jsii_.InvokeVoid(
		s,
		"resetSslCipher",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetUsername() {
	_jsii_.InvokeVoid(
		s,
		"resetUsername",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ResetVerifyServerCertificate() {
	_jsii_.InvokeVoid(
		s,
		"resetVerifyServerCertificate",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceReplicaConfigurationOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

