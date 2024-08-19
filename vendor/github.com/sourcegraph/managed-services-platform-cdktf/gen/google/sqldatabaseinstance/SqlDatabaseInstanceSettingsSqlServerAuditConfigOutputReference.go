package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference interface {
	cdktf.ComplexObject
	Bucket() *string
	SetBucket(val *string)
	BucketInput() *string
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
	InternalValue() *SqlDatabaseInstanceSettingsSqlServerAuditConfig
	SetInternalValue(val *SqlDatabaseInstanceSettingsSqlServerAuditConfig)
	RetentionInterval() *string
	SetRetentionInterval(val *string)
	RetentionIntervalInput() *string
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	UploadInterval() *string
	SetUploadInterval(val *string)
	UploadIntervalInput() *string
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
	ResetBucket()
	ResetRetentionInterval()
	ResetUploadInterval()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference
type jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) Bucket() *string {
	var returns *string
	_jsii_.Get(
		j,
		"bucket",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) BucketInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"bucketInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) InternalValue() *SqlDatabaseInstanceSettingsSqlServerAuditConfig {
	var returns *SqlDatabaseInstanceSettingsSqlServerAuditConfig
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) RetentionInterval() *string {
	var returns *string
	_jsii_.Get(
		j,
		"retentionInterval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) RetentionIntervalInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"retentionIntervalInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) UploadInterval() *string {
	var returns *string
	_jsii_.Get(
		j,
		"uploadInterval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) UploadIntervalInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"uploadIntervalInput",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference_Override(s SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetBucket(val *string) {
	if err := j.validateSetBucketParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"bucket",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetInternalValue(val *SqlDatabaseInstanceSettingsSqlServerAuditConfig) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetRetentionInterval(val *string) {
	if err := j.validateSetRetentionIntervalParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retentionInterval",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference)SetUploadInterval(val *string) {
	if err := j.validateSetUploadIntervalParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"uploadInterval",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) ResetBucket() {
	_jsii_.InvokeVoid(
		s,
		"resetBucket",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) ResetRetentionInterval() {
	_jsii_.InvokeVoid(
		s,
		"resetRetentionInterval",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) ResetUploadInterval() {
	_jsii_.InvokeVoid(
		s,
		"resetUploadInterval",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsSqlServerAuditConfigOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

