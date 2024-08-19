package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceSettingsBackupConfigurationOutputReference interface {
	cdktf.ComplexObject
	BackupRetentionSettings() SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettingsOutputReference
	BackupRetentionSettingsInput() *SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettings
	BinaryLogEnabled() interface{}
	SetBinaryLogEnabled(val interface{})
	BinaryLogEnabledInput() interface{}
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
	Enabled() interface{}
	SetEnabled(val interface{})
	EnabledInput() interface{}
	// Experimental.
	Fqn() *string
	InternalValue() *SqlDatabaseInstanceSettingsBackupConfiguration
	SetInternalValue(val *SqlDatabaseInstanceSettingsBackupConfiguration)
	Location() *string
	SetLocation(val *string)
	LocationInput() *string
	PointInTimeRecoveryEnabled() interface{}
	SetPointInTimeRecoveryEnabled(val interface{})
	PointInTimeRecoveryEnabledInput() interface{}
	StartTime() *string
	SetStartTime(val *string)
	StartTimeInput() *string
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	TransactionLogRetentionDays() *float64
	SetTransactionLogRetentionDays(val *float64)
	TransactionLogRetentionDaysInput() *float64
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
	PutBackupRetentionSettings(value *SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettings)
	ResetBackupRetentionSettings()
	ResetBinaryLogEnabled()
	ResetEnabled()
	ResetLocation()
	ResetPointInTimeRecoveryEnabled()
	ResetStartTime()
	ResetTransactionLogRetentionDays()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceSettingsBackupConfigurationOutputReference
type jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) BackupRetentionSettings() SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettingsOutputReference {
	var returns SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettingsOutputReference
	_jsii_.Get(
		j,
		"backupRetentionSettings",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) BackupRetentionSettingsInput() *SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettings {
	var returns *SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettings
	_jsii_.Get(
		j,
		"backupRetentionSettingsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) BinaryLogEnabled() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"binaryLogEnabled",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) BinaryLogEnabledInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"binaryLogEnabledInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) Enabled() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enabled",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) EnabledInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enabledInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) InternalValue() *SqlDatabaseInstanceSettingsBackupConfiguration {
	var returns *SqlDatabaseInstanceSettingsBackupConfiguration
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) Location() *string {
	var returns *string
	_jsii_.Get(
		j,
		"location",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) LocationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"locationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) PointInTimeRecoveryEnabled() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"pointInTimeRecoveryEnabled",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) PointInTimeRecoveryEnabledInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"pointInTimeRecoveryEnabledInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) StartTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"startTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) StartTimeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"startTimeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) TransactionLogRetentionDays() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"transactionLogRetentionDays",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) TransactionLogRetentionDaysInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"transactionLogRetentionDaysInput",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceSettingsBackupConfigurationOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceSettingsBackupConfigurationOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceSettingsBackupConfigurationOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsBackupConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceSettingsBackupConfigurationOutputReference_Override(s SqlDatabaseInstanceSettingsBackupConfigurationOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsBackupConfigurationOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetBinaryLogEnabled(val interface{}) {
	if err := j.validateSetBinaryLogEnabledParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"binaryLogEnabled",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetEnabled(val interface{}) {
	if err := j.validateSetEnabledParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enabled",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetInternalValue(val *SqlDatabaseInstanceSettingsBackupConfiguration) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetLocation(val *string) {
	if err := j.validateSetLocationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"location",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetPointInTimeRecoveryEnabled(val interface{}) {
	if err := j.validateSetPointInTimeRecoveryEnabledParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"pointInTimeRecoveryEnabled",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetStartTime(val *string) {
	if err := j.validateSetStartTimeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"startTime",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference)SetTransactionLogRetentionDays(val *float64) {
	if err := j.validateSetTransactionLogRetentionDaysParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"transactionLogRetentionDays",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) PutBackupRetentionSettings(value *SqlDatabaseInstanceSettingsBackupConfigurationBackupRetentionSettings) {
	if err := s.validatePutBackupRetentionSettingsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		s,
		"putBackupRetentionSettings",
		[]interface{}{value},
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ResetBackupRetentionSettings() {
	_jsii_.InvokeVoid(
		s,
		"resetBackupRetentionSettings",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ResetBinaryLogEnabled() {
	_jsii_.InvokeVoid(
		s,
		"resetBinaryLogEnabled",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ResetEnabled() {
	_jsii_.InvokeVoid(
		s,
		"resetEnabled",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ResetLocation() {
	_jsii_.InvokeVoid(
		s,
		"resetLocation",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ResetPointInTimeRecoveryEnabled() {
	_jsii_.InvokeVoid(
		s,
		"resetPointInTimeRecoveryEnabled",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ResetStartTime() {
	_jsii_.InvokeVoid(
		s,
		"resetStartTime",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ResetTransactionLogRetentionDays() {
	_jsii_.InvokeVoid(
		s,
		"resetTransactionLogRetentionDays",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsBackupConfigurationOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

