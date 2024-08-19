package sqluser

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqluser/internal"
)

type SqlUserPasswordPolicyOutputReference interface {
	cdktf.ComplexObject
	AllowedFailedAttempts() *float64
	SetAllowedFailedAttempts(val *float64)
	AllowedFailedAttemptsInput() *float64
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
	EnableFailedAttemptsCheck() interface{}
	SetEnableFailedAttemptsCheck(val interface{})
	EnableFailedAttemptsCheckInput() interface{}
	EnablePasswordVerification() interface{}
	SetEnablePasswordVerification(val interface{})
	EnablePasswordVerificationInput() interface{}
	// Experimental.
	Fqn() *string
	InternalValue() *SqlUserPasswordPolicy
	SetInternalValue(val *SqlUserPasswordPolicy)
	PasswordExpirationDuration() *string
	SetPasswordExpirationDuration(val *string)
	PasswordExpirationDurationInput() *string
	Status() SqlUserPasswordPolicyStatusList
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
	ResetAllowedFailedAttempts()
	ResetEnableFailedAttemptsCheck()
	ResetEnablePasswordVerification()
	ResetPasswordExpirationDuration()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlUserPasswordPolicyOutputReference
type jsiiProxy_SqlUserPasswordPolicyOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) AllowedFailedAttempts() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"allowedFailedAttempts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) AllowedFailedAttemptsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"allowedFailedAttemptsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) EnableFailedAttemptsCheck() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableFailedAttemptsCheck",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) EnableFailedAttemptsCheckInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enableFailedAttemptsCheckInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) EnablePasswordVerification() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enablePasswordVerification",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) EnablePasswordVerificationInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"enablePasswordVerificationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) InternalValue() *SqlUserPasswordPolicy {
	var returns *SqlUserPasswordPolicy
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) PasswordExpirationDuration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"passwordExpirationDuration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) PasswordExpirationDurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"passwordExpirationDurationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) Status() SqlUserPasswordPolicyStatusList {
	var returns SqlUserPasswordPolicyStatusList
	_jsii_.Get(
		j,
		"status",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewSqlUserPasswordPolicyOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlUserPasswordPolicyOutputReference {
	_init_.Initialize()

	if err := validateNewSqlUserPasswordPolicyOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlUserPasswordPolicyOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlUser.SqlUserPasswordPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlUserPasswordPolicyOutputReference_Override(s SqlUserPasswordPolicyOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlUser.SqlUserPasswordPolicyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetAllowedFailedAttempts(val *float64) {
	if err := j.validateSetAllowedFailedAttemptsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"allowedFailedAttempts",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetEnableFailedAttemptsCheck(val interface{}) {
	if err := j.validateSetEnableFailedAttemptsCheckParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enableFailedAttemptsCheck",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetEnablePasswordVerification(val interface{}) {
	if err := j.validateSetEnablePasswordVerificationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"enablePasswordVerification",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetInternalValue(val *SqlUserPasswordPolicy) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetPasswordExpirationDuration(val *string) {
	if err := j.validateSetPasswordExpirationDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"passwordExpirationDuration",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlUserPasswordPolicyOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) ResetAllowedFailedAttempts() {
	_jsii_.InvokeVoid(
		s,
		"resetAllowedFailedAttempts",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) ResetEnableFailedAttemptsCheck() {
	_jsii_.InvokeVoid(
		s,
		"resetEnableFailedAttemptsCheck",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) ResetEnablePasswordVerification() {
	_jsii_.InvokeVoid(
		s,
		"resetEnablePasswordVerification",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) ResetPasswordExpirationDuration() {
	_jsii_.InvokeVoid(
		s,
		"resetPasswordExpirationDuration",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_SqlUserPasswordPolicyOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

