package monitoringalertpolicy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/monitoringalertpolicy/internal"
)

type MonitoringAlertPolicyConditionsOutputReference interface {
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
	ConditionAbsent() MonitoringAlertPolicyConditionsConditionAbsentOutputReference
	ConditionAbsentInput() *MonitoringAlertPolicyConditionsConditionAbsent
	ConditionMatchedLog() MonitoringAlertPolicyConditionsConditionMatchedLogOutputReference
	ConditionMatchedLogInput() *MonitoringAlertPolicyConditionsConditionMatchedLog
	ConditionMonitoringQueryLanguage() MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageOutputReference
	ConditionMonitoringQueryLanguageInput() *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage
	ConditionPrometheusQueryLanguage() MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference
	ConditionPrometheusQueryLanguageInput() *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage
	ConditionThreshold() MonitoringAlertPolicyConditionsConditionThresholdOutputReference
	ConditionThresholdInput() *MonitoringAlertPolicyConditionsConditionThreshold
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	DisplayName() *string
	SetDisplayName(val *string)
	DisplayNameInput() *string
	// Experimental.
	Fqn() *string
	InternalValue() interface{}
	SetInternalValue(val interface{})
	Name() *string
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
	PutConditionAbsent(value *MonitoringAlertPolicyConditionsConditionAbsent)
	PutConditionMatchedLog(value *MonitoringAlertPolicyConditionsConditionMatchedLog)
	PutConditionMonitoringQueryLanguage(value *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage)
	PutConditionPrometheusQueryLanguage(value *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage)
	PutConditionThreshold(value *MonitoringAlertPolicyConditionsConditionThreshold)
	ResetConditionAbsent()
	ResetConditionMatchedLog()
	ResetConditionMonitoringQueryLanguage()
	ResetConditionPrometheusQueryLanguage()
	ResetConditionThreshold()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for MonitoringAlertPolicyConditionsOutputReference
type jsiiProxy_MonitoringAlertPolicyConditionsOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionAbsent() MonitoringAlertPolicyConditionsConditionAbsentOutputReference {
	var returns MonitoringAlertPolicyConditionsConditionAbsentOutputReference
	_jsii_.Get(
		j,
		"conditionAbsent",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionAbsentInput() *MonitoringAlertPolicyConditionsConditionAbsent {
	var returns *MonitoringAlertPolicyConditionsConditionAbsent
	_jsii_.Get(
		j,
		"conditionAbsentInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionMatchedLog() MonitoringAlertPolicyConditionsConditionMatchedLogOutputReference {
	var returns MonitoringAlertPolicyConditionsConditionMatchedLogOutputReference
	_jsii_.Get(
		j,
		"conditionMatchedLog",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionMatchedLogInput() *MonitoringAlertPolicyConditionsConditionMatchedLog {
	var returns *MonitoringAlertPolicyConditionsConditionMatchedLog
	_jsii_.Get(
		j,
		"conditionMatchedLogInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionMonitoringQueryLanguage() MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageOutputReference {
	var returns MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguageOutputReference
	_jsii_.Get(
		j,
		"conditionMonitoringQueryLanguage",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionMonitoringQueryLanguageInput() *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage {
	var returns *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage
	_jsii_.Get(
		j,
		"conditionMonitoringQueryLanguageInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionPrometheusQueryLanguage() MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference {
	var returns MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguageOutputReference
	_jsii_.Get(
		j,
		"conditionPrometheusQueryLanguage",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionPrometheusQueryLanguageInput() *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage {
	var returns *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage
	_jsii_.Get(
		j,
		"conditionPrometheusQueryLanguageInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionThreshold() MonitoringAlertPolicyConditionsConditionThresholdOutputReference {
	var returns MonitoringAlertPolicyConditionsConditionThresholdOutputReference
	_jsii_.Get(
		j,
		"conditionThreshold",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ConditionThresholdInput() *MonitoringAlertPolicyConditionsConditionThreshold {
	var returns *MonitoringAlertPolicyConditionsConditionThreshold
	_jsii_.Get(
		j,
		"conditionThresholdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) DisplayName() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayName",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) DisplayNameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"displayNameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) InternalValue() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}


func NewMonitoringAlertPolicyConditionsOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) MonitoringAlertPolicyConditionsOutputReference {
	_init_.Initialize()

	if err := validateNewMonitoringAlertPolicyConditionsOutputReferenceParameters(terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet); err != nil {
		panic(err)
	}
	j := jsiiProxy_MonitoringAlertPolicyConditionsOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		&j,
	)

	return &j
}

func NewMonitoringAlertPolicyConditionsOutputReference_Override(m MonitoringAlertPolicyConditionsOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string, complexObjectIndex *float64, complexObjectIsFromSet *bool) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.monitoringAlertPolicy.MonitoringAlertPolicyConditionsOutputReference",
		[]interface{}{terraformResource, terraformAttribute, complexObjectIndex, complexObjectIsFromSet},
		m,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference)SetDisplayName(val *string) {
	if err := j.validateSetDisplayNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"displayName",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference)SetInternalValue(val interface{}) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		m,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) PutConditionAbsent(value *MonitoringAlertPolicyConditionsConditionAbsent) {
	if err := m.validatePutConditionAbsentParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putConditionAbsent",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) PutConditionMatchedLog(value *MonitoringAlertPolicyConditionsConditionMatchedLog) {
	if err := m.validatePutConditionMatchedLogParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putConditionMatchedLog",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) PutConditionMonitoringQueryLanguage(value *MonitoringAlertPolicyConditionsConditionMonitoringQueryLanguage) {
	if err := m.validatePutConditionMonitoringQueryLanguageParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putConditionMonitoringQueryLanguage",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) PutConditionPrometheusQueryLanguage(value *MonitoringAlertPolicyConditionsConditionPrometheusQueryLanguage) {
	if err := m.validatePutConditionPrometheusQueryLanguageParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putConditionPrometheusQueryLanguage",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) PutConditionThreshold(value *MonitoringAlertPolicyConditionsConditionThreshold) {
	if err := m.validatePutConditionThresholdParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		m,
		"putConditionThreshold",
		[]interface{}{value},
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ResetConditionAbsent() {
	_jsii_.InvokeVoid(
		m,
		"resetConditionAbsent",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ResetConditionMatchedLog() {
	_jsii_.InvokeVoid(
		m,
		"resetConditionMatchedLog",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ResetConditionMonitoringQueryLanguage() {
	_jsii_.InvokeVoid(
		m,
		"resetConditionMonitoringQueryLanguage",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ResetConditionPrometheusQueryLanguage() {
	_jsii_.InvokeVoid(
		m,
		"resetConditionPrometheusQueryLanguage",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ResetConditionThreshold() {
	_jsii_.InvokeVoid(
		m,
		"resetConditionThreshold",
		nil, // no parameters
	)
}

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (m *jsiiProxy_MonitoringAlertPolicyConditionsOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		m,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

