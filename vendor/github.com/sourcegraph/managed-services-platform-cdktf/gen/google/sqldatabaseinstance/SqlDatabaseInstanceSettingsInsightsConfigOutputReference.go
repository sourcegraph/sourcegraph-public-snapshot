package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceSettingsInsightsConfigOutputReference interface {
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
	InternalValue() *SqlDatabaseInstanceSettingsInsightsConfig
	SetInternalValue(val *SqlDatabaseInstanceSettingsInsightsConfig)
	QueryInsightsEnabled() interface{}
	SetQueryInsightsEnabled(val interface{})
	QueryInsightsEnabledInput() interface{}
	QueryPlansPerMinute() *float64
	SetQueryPlansPerMinute(val *float64)
	QueryPlansPerMinuteInput() *float64
	QueryStringLength() *float64
	SetQueryStringLength(val *float64)
	QueryStringLengthInput() *float64
	RecordApplicationTags() interface{}
	SetRecordApplicationTags(val interface{})
	RecordApplicationTagsInput() interface{}
	RecordClientAddress() interface{}
	SetRecordClientAddress(val interface{})
	RecordClientAddressInput() interface{}
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
	ResetQueryInsightsEnabled()
	ResetQueryPlansPerMinute()
	ResetQueryStringLength()
	ResetRecordApplicationTags()
	ResetRecordClientAddress()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceSettingsInsightsConfigOutputReference
type jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) InternalValue() *SqlDatabaseInstanceSettingsInsightsConfig {
	var returns *SqlDatabaseInstanceSettingsInsightsConfig
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) QueryInsightsEnabled() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"queryInsightsEnabled",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) QueryInsightsEnabledInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"queryInsightsEnabledInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) QueryPlansPerMinute() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"queryPlansPerMinute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) QueryPlansPerMinuteInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"queryPlansPerMinuteInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) QueryStringLength() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"queryStringLength",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) QueryStringLengthInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"queryStringLengthInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) RecordApplicationTags() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"recordApplicationTags",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) RecordApplicationTagsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"recordApplicationTagsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) RecordClientAddress() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"recordClientAddress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) RecordClientAddressInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"recordClientAddressInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceSettingsInsightsConfigOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceSettingsInsightsConfigOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceSettingsInsightsConfigOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsInsightsConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceSettingsInsightsConfigOutputReference_Override(s SqlDatabaseInstanceSettingsInsightsConfigOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsInsightsConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetInternalValue(val *SqlDatabaseInstanceSettingsInsightsConfig) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetQueryInsightsEnabled(val interface{}) {
	if err := j.validateSetQueryInsightsEnabledParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"queryInsightsEnabled",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetQueryPlansPerMinute(val *float64) {
	if err := j.validateSetQueryPlansPerMinuteParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"queryPlansPerMinute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetQueryStringLength(val *float64) {
	if err := j.validateSetQueryStringLengthParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"queryStringLength",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetRecordApplicationTags(val interface{}) {
	if err := j.validateSetRecordApplicationTagsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"recordApplicationTags",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetRecordClientAddress(val interface{}) {
	if err := j.validateSetRecordClientAddressParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"recordClientAddress",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ResetQueryInsightsEnabled() {
	_jsii_.InvokeVoid(
		s,
		"resetQueryInsightsEnabled",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ResetQueryPlansPerMinute() {
	_jsii_.InvokeVoid(
		s,
		"resetQueryPlansPerMinute",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ResetQueryStringLength() {
	_jsii_.InvokeVoid(
		s,
		"resetQueryStringLength",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ResetRecordApplicationTags() {
	_jsii_.InvokeVoid(
		s,
		"resetRecordApplicationTags",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ResetRecordClientAddress() {
	_jsii_.InvokeVoid(
		s,
		"resetRecordClientAddress",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsInsightsConfigOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

