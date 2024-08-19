package monitoringalertpolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy/internal"
)

type MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference interface {
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
	Count() *float64
	SetCount(val *float64)
	CountInput() *float64
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	// Experimental.
	Fqn() *string
	InternalValue() *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger
	SetInternalValue(val *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger)
	Percent() *float64
	SetPercent(val *float64)
	PercentInput() *float64
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
	ResetCount()
	ResetPercent()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference
type jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) Count() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) CountInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"countInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) InternalValue() *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger {
	var returns *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) Percent() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"percent",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) PercentInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"percentInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewMonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference {
	_init_.Initialize()

	if err := validateNewMonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewMonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference_Override(m MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		m,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference)SetCount(val *float64) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference)SetInternalValue(val *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTrigger) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference)SetPercent(val *float64) {
	if err := j.validateSetPercentParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"percent",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) ResetCount() {
	_jsii_.InvokeVoid(
		m,
		"resetCount",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) ResetPercent() {
	_jsii_.InvokeVoid(
		m,
		"resetPercent",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageTriggerOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

