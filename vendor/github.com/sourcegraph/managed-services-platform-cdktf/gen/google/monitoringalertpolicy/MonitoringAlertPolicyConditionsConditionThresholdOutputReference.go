package monitoringalertpolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy/internal"
)

type MonitoringAlertPolicyConditionsConditionThresholdOutputReference interface {
	cdktf.ComplexObject
	Aggregations() MonitoringAlertPolicyConditionsConditionThresholdAggregationsList
	AggregationsInput() interface{}
	Comparison() *string
	SetComparison(val *string)
	ComparisonInput() *string
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
	DenominatorAggregations() MonitoringAlertPolicyConditionsConditionThresholdDenominatorAggregationsList
	DenominatorAggregationsInput() interface{}
	DenominatorFilter() *string
	SetDenominatorFilter(val *string)
	DenominatorFilterInput() *string
	Duration() *string
	SetDuration(val *string)
	DurationInput() *string
	EvaluationMissingData() *string
	SetEvaluationMissingData(val *string)
	EvaluationMissingDataInput() *string
	Filter() *string
	SetFilter(val *string)
	FilterInput() *string
	ForecastOptions() MonitoringAlertPolicyConditionsConditionThresholdForecastOptionsOutputReference
	ForecastOptionsInput() *MonitoringAlertPolicyConditionsConditionThresholdForecastOptions
	// Experimental.
	Fqn() *string
	InternalValue() *MonitoringAlertPolicyConditionsConditionThreshold
	SetInternalValue(val *MonitoringAlertPolicyConditionsConditionThreshold)
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	ThresholdValue() *float64
	SetThresholdValue(val *float64)
	ThresholdValueInput() *float64
	Trigger() MonitoringAlertPolicyConditionsConditionThresholdTriggerOutputReference
	TriggerInput() *MonitoringAlertPolicyConditionsConditionThresholdTrigger
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
	PutAggregations(value interface{})
	PutDenominatorAggregations(value interface{})
	PutForecastOptions(value *MonitoringAlertPolicyConditionsConditionThresholdForecastOptions)
	PutTrigger(value *MonitoringAlertPolicyConditionsConditionThresholdTrigger)
	ResetAggregations()
	ResetDenominatorAggregations()
	ResetDenominatorFilter()
	ResetEvaluationMissingData()
	ResetFilter()
	ResetForecastOptions()
	ResetThresholdValue()
	ResetTrigger()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for MonitoringAlertPolicyConditionsConditionThresholdOutputReference
type jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) Aggregations() MonitoringAlertPolicyConditionsConditionThresholdAggregationsList {
	var returns MonitoringAlertPolicyConditionsConditionThresholdAggregationsList
	_jsii_.Get(
		j,
		"aggregations",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) AggregationsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"aggregationsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) Comparison() *string {
	var returns *string
	_jsii_.Get(
		j,
		"comparison",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ComparisonInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"comparisonInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) DenominatorAggregations() MonitoringAlertPolicyConditionsConditionThresholdDenominatorAggregationsList {
	var returns MonitoringAlertPolicyConditionsConditionThresholdDenominatorAggregationsList
	_jsii_.Get(
		j,
		"denominatorAggregations",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) DenominatorAggregationsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"denominatorAggregationsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) DenominatorFilter() *string {
	var returns *string
	_jsii_.Get(
		j,
		"denominatorFilter",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) DenominatorFilterInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"denominatorFilterInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) Duration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"duration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) DurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"durationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) EvaluationMissingData() *string {
	var returns *string
	_jsii_.Get(
		j,
		"evaluationMissingData",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) EvaluationMissingDataInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"evaluationMissingDataInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) Filter() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filter",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) FilterInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"filterInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ForecastOptions() MonitoringAlertPolicyConditionsConditionThresholdForecastOptionsOutputReference {
	var returns MonitoringAlertPolicyConditionsConditionThresholdForecastOptionsOutputReference
	_jsii_.Get(
		j,
		"forecastOptions",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ForecastOptionsInput() *MonitoringAlertPolicyConditionsConditionThresholdForecastOptions {
	var returns *MonitoringAlertPolicyConditionsConditionThresholdForecastOptions
	_jsii_.Get(
		j,
		"forecastOptionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) InternalValue() *MonitoringAlertPolicyConditionsConditionThreshold {
	var returns *MonitoringAlertPolicyConditionsConditionThreshold
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ThresholdValue() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"thresholdValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ThresholdValueInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"thresholdValueInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) Trigger() MonitoringAlertPolicyConditionsConditionThresholdTriggerOutputReference {
	var returns MonitoringAlertPolicyConditionsConditionThresholdTriggerOutputReference
	_jsii_.Get(
		j,
		"trigger",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) TriggerInput() *MonitoringAlertPolicyConditionsConditionThresholdTrigger {
	var returns *MonitoringAlertPolicyConditionsConditionThresholdTrigger
	_jsii_.Get(
		j,
		"triggerInput",
		&returns,
	)
	return returns
}


func NewMonitoringAlertPolicyConditionsConditionThresholdOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) MonitoringAlertPolicyConditionsConditionThresholdOutputReference {
	_init_.Initialize()

	if err := validateNewMonitoringAlertPolicyConditionsConditionThresholdOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionThresholdOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewMonitoringAlertPolicyConditionsConditionThresholdOutputReference_Override(m MonitoringAlertPolicyConditionsConditionThresholdOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionThresholdOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		m,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetComparison(val *string) {
	if err := j.validateSetComparisonParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"comparison",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetDenominatorFilter(val *string) {
	if err := j.validateSetDenominatorFilterParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"denominatorFilter",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetDuration(val *string) {
	if err := j.validateSetDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"duration",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetEvaluationMissingData(val *string) {
	if err := j.validateSetEvaluationMissingDataParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"evaluationMissingData",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetFilter(val *string) {
	if err := j.validateSetFilterParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"filter",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetInternalValue(val *MonitoringAlertPolicyConditionsConditionThreshold) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference)SetThresholdValue(val *float64) {
	if err := j.validateSetThresholdValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"thresholdValue",
		val,
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) PutAggregations(value interface{}) {
	if err := m.validatePutAggregationsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putAggregations",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) PutDenominatorAggregations(value interface{}) {
	if err := m.validatePutDenominatorAggregationsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putDenominatorAggregations",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) PutForecastOptions(value *MonitoringAlertPolicyConditionsConditionThresholdForecastOptions) {
	if err := m.validatePutForecastOptionsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putForecastOptions",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) PutTrigger(value *MonitoringAlertPolicyConditionsConditionThresholdTrigger) {
	if err := m.validatePutTriggerParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putTrigger",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetAggregations() {
	_jsii_.InvokeVoid(
		m,
		"resetAggregations",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetDenominatorAggregations() {
	_jsii_.InvokeVoid(
		m,
		"resetDenominatorAggregations",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetDenominatorFilter() {
	_jsii_.InvokeVoid(
		m,
		"resetDenominatorFilter",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetEvaluationMissingData() {
	_jsii_.InvokeVoid(
		m,
		"resetEvaluationMissingData",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetFilter() {
	_jsii_.InvokeVoid(
		m,
		"resetFilter",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetForecastOptions() {
	_jsii_.InvokeVoid(
		m,
		"resetForecastOptions",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetThresholdValue() {
	_jsii_.InvokeVoid(
		m,
		"resetThresholdValue",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ResetTrigger() {
	_jsii_.InvokeVoid(
		m,
		"resetTrigger",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionThresholdOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

