package monitoringalertpolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy/internal"
)

type MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference interface {
	cdktf.ComplexObject
	AlertRule() *string
	SetAlertRule(val *string)
	AlertRuleInput() *string
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
	Duration() *string
	SetDuration(val *string)
	DurationInput() *string
	EvaluationInterval() *string
	SetEvaluationInterval(val *string)
	EvaluationIntervalInput() *string
	// Experimental.
	Fqn() *string
	InternalValue() *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage
	SetInternalValue(val *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage)
	Labels() *map[string]*string
	SetLabels(val *map[string]*string)
	LabelsInput() *map[string]*string
	Query() *string
	SetQuery(val *string)
	QueryInput() *string
	RuleGroup() *string
	SetRuleGroup(val *string)
	RuleGroupInput() *string
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
	ResetAlertRule()
	ResetDuration()
	ResetEvaluationInterval()
	ResetLabels()
	ResetRuleGroup()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference
type jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) AlertRule() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alertRule",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) AlertRuleInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alertRuleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) Duration() *string {
	var returns *string
	_jsii_.Get(
		j,
		"duration",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) DurationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"durationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) EvaluationInterval() *string {
	var returns *string
	_jsii_.Get(
		j,
		"evaluationInterval",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) EvaluationIntervalInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"evaluationIntervalInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) InternalValue() *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage {
	var returns *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) Labels() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labels",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) LabelsInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"labelsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) Query() *string {
	var returns *string
	_jsii_.Get(
		j,
		"query",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) QueryInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"queryInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) RuleGroup() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ruleGroup",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) RuleGroupInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ruleGroupInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewMonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference {
	_init_.Initialize()

	if err := validateNewMonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewMonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference_Override(m MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		m,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetAlertRule(val *string) {
	if err := j.validateSetAlertRuleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"alertRule",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetDuration(val *string) {
	if err := j.validateSetDurationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"duration",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetEvaluationInterval(val *string) {
	if err := j.validateSetEvaluationIntervalParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"evaluationInterval",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetInternalValue(val *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetLabels(val *map[string]*string) {
	if err := j.validateSetLabelsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"labels",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetQuery(val *string) {
	if err := j.validateSetQueryParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"query",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetRuleGroup(val *string) {
	if err := j.validateSetRuleGroupParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ruleGroup",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ResetAlertRule() {
	_jsii_.InvokeVoid(
		m,
		"resetAlertRule",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ResetDuration() {
	_jsii_.InvokeVoid(
		m,
		"resetDuration",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ResetEvaluationInterval() {
	_jsii_.InvokeVoid(
		m,
		"resetEvaluationInterval",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ResetLabels() {
	_jsii_.InvokeVoid(
		m,
		"resetLabels",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ResetRuleGroup() {
	_jsii_.InvokeVoid(
		m,
		"resetRuleGroup",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

