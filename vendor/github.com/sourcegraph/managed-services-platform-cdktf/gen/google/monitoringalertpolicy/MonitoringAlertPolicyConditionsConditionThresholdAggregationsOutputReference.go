package monitoringalertpolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy/internal"
)

type MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference interface {
	cdktf.ComplexObject
	AlignmentPeriod() *string
	SetAlignmentPeriod(val *string)
	AlignmentPeriodInput() *string
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
	CrossSeriesReducer() *string
	SetCrossSeriesReducer(val *string)
	CrossSeriesReducerInput() *string
	// Experimental.
	Fqn() *string
	GroupByFields() *[]*string
	SetGroupByFields(val *[]*string)
	GroupByFieldsInput() *[]*string
	InternalValue() interface{}
	SetInternalValue(val interface{})
	PerSeriesAligner() *string
	SetPerSeriesAligner(val *string)
	PerSeriesAlignerInput() *string
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
	ResetAlignmentPeriod()
	ResetCrossSeriesReducer()
	ResetGroupByFields()
	ResetPerSeriesAligner()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference
type jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) AlignmentPeriod() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alignmentPeriod",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) AlignmentPeriodInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alignmentPeriodInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) CrossSeriesReducer() *string {
	var returns *string
	_jsii_.Get(
		j,
		"crossSeriesReducer",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) CrossSeriesReducerInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"crossSeriesReducerInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GroupByFields() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"groupByFields",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GroupByFieldsInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"groupByFieldsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) PerSeriesAligner() *string {
	var returns *string
	_jsii_.Get(
		j,
		"perSeriesAligner",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) PerSeriesAlignerInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"perSeriesAlignerInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewMonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference {
	_init_.Initialize()

	if err := validateNewMonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewMonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference_Override(m MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		m,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetAlignmentPeriod(val *string) {
	if err := j.validateSetAlignmentPeriodParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"alignmentPeriod",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetCrossSeriesReducer(val *string) {
	if err := j.validateSetCrossSeriesReducerParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"crossSeriesReducer",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetGroupByFields(val *[]*string) {
	if err := j.validateSetGroupByFieldsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"groupByFields",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetPerSeriesAligner(val *string) {
	if err := j.validateSetPerSeriesAlignerParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"perSeriesAligner",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ResetAlignmentPeriod() {
	_jsii_.InvokeVoid(
		m,
		"resetAlignmentPeriod",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ResetCrossSeriesReducer() {
	_jsii_.InvokeVoid(
		m,
		"resetCrossSeriesReducer",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ResetGroupByFields() {
	_jsii_.InvokeVoid(
		m,
		"resetGroupByFields",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ResetPerSeriesAligner() {
	_jsii_.InvokeVoid(
		m,
		"resetPerSeriesAligner",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdAggregationsOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

