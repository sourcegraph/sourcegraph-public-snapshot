package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference interface {
	cdktf.ComplexObject
	Complexity() *string
	SetComplexity(val *string)
	ComplexityInput() *string
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
	DisallowUsernameSubstring() interface{}
	SetDisallowUsernameSubstring(val interface{})
	DisallowUsernameSubstringInput() interface{}
	EnablePasswordPolicy() interface{}
	SetEnablePasswordPolicy(val interface{})
	EnablePasswordPolicyInput() interface{}
	// Experimental.
	Fqn() *string
	InternalValue() *SqlDatabaseInstanceSettingsPasswordValidationPolicy
	SetInternalValue(val *SqlDatabaseInstanceSettingsPasswordValidationPolicy)
	MinLength() *float64
	SetMinLength(val *float64)
	MinLengthInput() *float64
	PasswordChangeInterval() *string
	SetPasswordChangeInterval(val *string)
	PasswordChangeIntervalInput() *string
	ReuseInterval() *float64
	SetReuseInterval(val *float64)
	ReuseIntervalInput() *float64
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
	ResetComplexity()
	ResetDisallowUsernameSubstring()
	ResetMinLength()
	ResetPasswordChangeInterval()
	ResetReuseInterval()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference
type jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) Complexity() *string {
	var returns *string
	_jsii_.Get(
		j,
		"complexity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ComplexityInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"complexityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) DisallowUsernameSubstring() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"disallowUsernameSubstring",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) DisallowUsernameSubstringInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"disallowUsernameSubstringInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) EnablePasswordPolicy() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enablePasswordPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) EnablePasswordPolicyInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enablePasswordPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) InternalValue() *SqlDatabaseInstanceSettingsPasswordValidationPolicy {
	var returns *SqlDatabaseInstanceSettingsPasswordValidationPolicy
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) MinLength() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minLength",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) MinLengthInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minLengthInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) PasswordChangeInterval() *string {
	var returns *string
	_jsii_.Get(
		j,
		"passwordChangeInterval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) PasswordChangeIntervalInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"passwordChangeIntervalInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ReuseInterval() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"reuseInterval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ReuseIntervalInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"reuseIntervalInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference_Override(s SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetComplexity(val *string) {
	if err := j.validateSetComplexityParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexity",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetDisallowUsernameSubstring(val interface{}) {
	if err := j.validateSetDisallowUsernameSubstringParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"disallowUsernameSubstring",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetEnablePasswordPolicy(val interface{}) {
	if err := j.validateSetEnablePasswordPolicyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enablePasswordPolicy",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetInternalValue(val *SqlDatabaseInstanceSettingsPasswordValidationPolicy) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetMinLength(val *float64) {
	if err := j.validateSetMinLengthParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"minLength",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetPasswordChangeInterval(val *string) {
	if err := j.validateSetPasswordChangeIntervalParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"passwordChangeInterval",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetReuseInterval(val *float64) {
	if err := j.validateSetReuseIntervalParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"reuseInterval",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ResetComplexity() {
	_jsii_.InvokeVoid(
		s,
		"resetComplexity",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ResetDisallowUsernameSubstring() {
	_jsii_.InvokeVoid(
		s,
		"resetDisallowUsernameSubstring",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ResetMinLength() {
	_jsii_.InvokeVoid(
		s,
		"resetMinLength",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ResetPasswordChangeInterval() {
	_jsii_.InvokeVoid(
		s,
		"resetPasswordChangeInterval",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ResetReuseInterval() {
	_jsii_.InvokeVoid(
		s,
		"resetReuseInterval",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceSettingsPasswordValidationPolicyOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

