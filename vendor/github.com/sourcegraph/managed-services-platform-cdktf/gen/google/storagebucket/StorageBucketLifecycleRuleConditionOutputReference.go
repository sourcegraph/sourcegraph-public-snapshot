package storagebucket

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/storagebucket/internal"
)

type StorageBucketLifecycleRuleConditionOutputReference interface {
	cdktf.ComplexObject
	Age() *float64
	SetAge(val *float64)
	AgeInput() *float64
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
	CreatedBefore() *string
	SetCreatedBefore(val *string)
	CreatedBeforeInput() *string
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	CustomTimeBefore() *string
	SetCustomTimeBefore(val *string)
	CustomTimeBeforeInput() *string
	DaysSinceCustomTime() *float64
	SetDaysSinceCustomTime(val *float64)
	DaysSinceCustomTimeInput() *float64
	DaysSinceNoncurrentTime() *float64
	SetDaysSinceNoncurrentTime(val *float64)
	DaysSinceNoncurrentTimeInput() *float64
	// Experimental.
	Fqn() *string
	InternalValue() *StorageBucketLifecycleRuleCondition
	SetInternalValue(val *StorageBucketLifecycleRuleCondition)
	MatchesPrefix() *[]*string
	SetMatchesPrefix(val *[]*string)
	MatchesPrefixInput() *[]*string
	MatchesStorageClass() *[]*string
	SetMatchesStorageClass(val *[]*string)
	MatchesStorageClassInput() *[]*string
	MatchesSuffix() *[]*string
	SetMatchesSuffix(val *[]*string)
	MatchesSuffixInput() *[]*string
	NoAge() interface{}
	SetNoAge(val interface{})
	NoAgeInput() interface{}
	NoncurrentTimeBefore() *string
	SetNoncurrentTimeBefore(val *string)
	NoncurrentTimeBeforeInput() *string
	NumNewerVersions() *float64
	SetNumNewerVersions(val *float64)
	NumNewerVersionsInput() *float64
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	WithState() *string
	SetWithState(val *string)
	WithStateInput() *string
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
	ResetAge()
	ResetCreatedBefore()
	ResetCustomTimeBefore()
	ResetDaysSinceCustomTime()
	ResetDaysSinceNoncurrentTime()
	ResetMatchesPrefix()
	ResetMatchesStorageClass()
	ResetMatchesSuffix()
	ResetNoAge()
	ResetNoncurrentTimeBefore()
	ResetNumNewerVersions()
	ResetWithState()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for StorageBucketLifecycleRuleConditionOutputReference
type jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) Age() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"age",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) AgeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"ageInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) CreatedBefore() *string {
	var returns *string
	_jsii_.Get(
		j,
		"createdBefore",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) CreatedBeforeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"createdBeforeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) CustomTimeBefore() *string {
	var returns *string
	_jsii_.Get(
		j,
		"customTimeBefore",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) CustomTimeBeforeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"customTimeBeforeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) DaysSinceCustomTime() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"daysSinceCustomTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) DaysSinceCustomTimeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"daysSinceCustomTimeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) DaysSinceNoncurrentTime() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"daysSinceNoncurrentTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) DaysSinceNoncurrentTimeInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"daysSinceNoncurrentTimeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) InternalValue() *StorageBucketLifecycleRuleCondition {
	var returns *StorageBucketLifecycleRuleCondition
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) MatchesPrefix() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"matchesPrefix",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) MatchesPrefixInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"matchesPrefixInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) MatchesStorageClass() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"matchesStorageClass",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) MatchesStorageClassInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"matchesStorageClassInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) MatchesSuffix() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"matchesSuffix",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) MatchesSuffixInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"matchesSuffixInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) NoAge() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"noAge",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) NoAgeInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"noAgeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) NoncurrentTimeBefore() *string {
	var returns *string
	_jsii_.Get(
		j,
		"noncurrentTimeBefore",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) NoncurrentTimeBeforeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"noncurrentTimeBeforeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) NumNewerVersions() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"numNewerVersions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) NumNewerVersionsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"numNewerVersionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) WithState() *string {
	var returns *string
	_jsii_.Get(
		j,
		"withState",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) WithStateInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"withStateInput",
		&returns,
	)
	return returns
}


func NewStorageBucketLifecycleRuleConditionOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) StorageBucketLifecycleRuleConditionOutputReference {
	_init_.Initialize()

	if err := validateNewStorageBucketLifecycleRuleConditionOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.storageBucket.StorageBucketLifecycleRuleConditionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewStorageBucketLifecycleRuleConditionOutputReference_Override(s StorageBucketLifecycleRuleConditionOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.storageBucket.StorageBucketLifecycleRuleConditionOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetAge(val *float64) {
	if err := j.validateSetAgeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"age",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetCreatedBefore(val *string) {
	if err := j.validateSetCreatedBeforeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"createdBefore",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetCustomTimeBefore(val *string) {
	if err := j.validateSetCustomTimeBeforeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"customTimeBefore",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetDaysSinceCustomTime(val *float64) {
	if err := j.validateSetDaysSinceCustomTimeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"daysSinceCustomTime",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetDaysSinceNoncurrentTime(val *float64) {
	if err := j.validateSetDaysSinceNoncurrentTimeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"daysSinceNoncurrentTime",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetInternalValue(val *StorageBucketLifecycleRuleCondition) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetMatchesPrefix(val *[]*string) {
	if err := j.validateSetMatchesPrefixParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"matchesPrefix",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetMatchesStorageClass(val *[]*string) {
	if err := j.validateSetMatchesStorageClassParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"matchesStorageClass",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetMatchesSuffix(val *[]*string) {
	if err := j.validateSetMatchesSuffixParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"matchesSuffix",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetNoAge(val interface{}) {
	if err := j.validateSetNoAgeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"noAge",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetNoncurrentTimeBefore(val *string) {
	if err := j.validateSetNoncurrentTimeBeforeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"noncurrentTimeBefore",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetNumNewerVersions(val *float64) {
	if err := j.validateSetNumNewerVersionsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"numNewerVersions",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference)SetWithState(val *string) {
	if err := j.validateSetWithStateParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"withState",
		val,
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetAge() {
	_jsii_.InvokeVoid(
		s,
		"resetAge",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetCreatedBefore() {
	_jsii_.InvokeVoid(
		s,
		"resetCreatedBefore",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetCustomTimeBefore() {
	_jsii_.InvokeVoid(
		s,
		"resetCustomTimeBefore",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetDaysSinceCustomTime() {
	_jsii_.InvokeVoid(
		s,
		"resetDaysSinceCustomTime",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetDaysSinceNoncurrentTime() {
	_jsii_.InvokeVoid(
		s,
		"resetDaysSinceNoncurrentTime",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetMatchesPrefix() {
	_jsii_.InvokeVoid(
		s,
		"resetMatchesPrefix",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetMatchesStorageClass() {
	_jsii_.InvokeVoid(
		s,
		"resetMatchesStorageClass",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetMatchesSuffix() {
	_jsii_.InvokeVoid(
		s,
		"resetMatchesSuffix",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetNoAge() {
	_jsii_.InvokeVoid(
		s,
		"resetNoAge",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetNoncurrentTimeBefore() {
	_jsii_.InvokeVoid(
		s,
		"resetNoncurrentTimeBefore",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetNumNewerVersions() {
	_jsii_.InvokeVoid(
		s,
		"resetNumNewerVersions",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ResetWithState() {
	_jsii_.InvokeVoid(
		s,
		"resetWithState",
		nil, // no parameters
	)
}

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_StorageBucketLifecycleRuleConditionOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

