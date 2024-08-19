package sqldatabaseinstance

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqldatabaseinstance/internal"
)

type SqlDatabaseInstanceCloneOutputReference interface {
	cdktf.ComplexObject
	AllocatedIpRange() *string
	SetAllocatedIpRange(val *string)
	AllocatedIpRangeInput() *string
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
	DatabaseNames() *[]*string
	SetDatabaseNames(val *[]*string)
	DatabaseNamesInput() *[]*string
	// Experimental.
	Fqn() *string
	InternalValue() *SqlDatabaseInstanceClone
	SetInternalValue(val *SqlDatabaseInstanceClone)
	PointInTime() *string
	SetPointInTime(val *string)
	PointInTimeInput() *string
	PreferredZone() *string
	SetPreferredZone(val *string)
	PreferredZoneInput() *string
	SourceInstanceName() *string
	SetSourceInstanceName(val *string)
	SourceInstanceNameInput() *string
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
	ResetAllocatedIpRange()
	ResetDatabaseNames()
	ResetPointInTime()
	ResetPreferredZone()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for SqlDatabaseInstanceCloneOutputReference
type jsiiProxy_SqlDatabaseInstanceCloneOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) AllocatedIpRange() *string {
	var returns *string
	_jsii_.Get(
		j,
		"allocatedIpRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) AllocatedIpRangeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"allocatedIpRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) DatabaseNames() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"databaseNames",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) DatabaseNamesInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"databaseNamesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) InternalValue() *SqlDatabaseInstanceClone {
	var returns *SqlDatabaseInstanceClone
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) PointInTime() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pointInTime",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) PointInTimeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"pointInTimeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) PreferredZone() *string {
	var returns *string
	_jsii_.Get(
		j,
		"preferredZone",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) PreferredZoneInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"preferredZoneInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) SourceInstanceName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceInstanceName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) SourceInstanceNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sourceInstanceNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewSqlDatabaseInstanceCloneOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) SqlDatabaseInstanceCloneOutputReference {
	_init_.Initialize()

	if err := validateNewSqlDatabaseInstanceCloneOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_SqlDatabaseInstanceCloneOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceCloneOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewSqlDatabaseInstanceCloneOutputReference_Override(s SqlDatabaseInstanceCloneOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.sqlDatabaseInstance.SqlDatabaseInstanceCloneOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		s,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetAllocatedIpRange(val *string) {
	if err := j.validateSetAllocatedIpRangeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"allocatedIpRange",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetDatabaseNames(val *[]*string) {
	if err := j.validateSetDatabaseNamesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"databaseNames",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetInternalValue(val *SqlDatabaseInstanceClone) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetPointInTime(val *string) {
	if err := j.validateSetPointInTimeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"pointInTime",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetPreferredZone(val *string) {
	if err := j.validateSetPreferredZoneParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"preferredZone",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetSourceInstanceName(val *string) {
	if err := j.validateSetSourceInstanceNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sourceInstanceName",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_SqlDatabaseInstanceCloneOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		s,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ResetAllocatedIpRange() {
	_jsii_.InvokeVoid(
		s,
		"resetAllocatedIpRange",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ResetDatabaseNames() {
	_jsii_.InvokeVoid(
		s,
		"resetDatabaseNames",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ResetPointInTime() {
	_jsii_.InvokeVoid(
		s,
		"resetPointInTime",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ResetPreferredZone() {
	_jsii_.InvokeVoid(
		s,
		"resetPreferredZone",
		nil, // no parameters
	)
}

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (s *jsiiProxy_SqlDatabaseInstanceCloneOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		s,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

