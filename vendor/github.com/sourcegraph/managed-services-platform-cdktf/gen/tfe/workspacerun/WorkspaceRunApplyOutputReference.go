package workspacerun

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/workspacerun/internal"
)

type WorkspaceRunApplyOutputReference interface {
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
	// The creation stack of this resolvable which will be appended to errors thrown during resolution.
	//
	// If this returns an empty array the stack will not be attached.
	// Experimental.
	CreationStack() *[]*string
	// Experimental.
	Fqn() *string
	InternalValue() *WorkspaceRunApply
	SetInternalValue(val *WorkspaceRunApply)
	ManualConfirm() interface{}
	SetManualConfirm(val interface{})
	ManualConfirmInput() interface{}
	Retry() interface{}
	SetRetry(val interface{})
	RetryAttempts() *float64
	SetRetryAttempts(val *float64)
	RetryAttemptsInput() *float64
	RetryBackoffMax() *float64
	SetRetryBackoffMax(val *float64)
	RetryBackoffMaxInput() *float64
	RetryBackoffMin() *float64
	SetRetryBackoffMin(val *float64)
	RetryBackoffMinInput() *float64
	RetryInput() interface{}
	// Experimental.
	TerraformAttribute() *string
	// Experimental.
	SetTerraformAttribute(val *string)
	// Experimental.
	TerraformResource() cdktf.IInterpolatingParent
	// Experimental.
	SetTerraformResource(val cdktf.IInterpolatingParent)
	WaitForRun() interface{}
	SetWaitForRun(val interface{})
	WaitForRunInput() interface{}
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
	ResetRetry()
	ResetRetryAttempts()
	ResetRetryBackoffMax()
	ResetRetryBackoffMin()
	ResetWaitForRun()
	// Produce the Token's value at resolution time.
	// Experimental.
	Resolve(_context cdktf.IResolveContext) interface{}
	// Return a string representation of this resolvable object.
	//
	// Returns a reversible string representation.
	// Experimental.
	ToString() *string
}

// The jsii proxy struct for WorkspaceRunApplyOutputReference
type jsiiProxy_WorkspaceRunApplyOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) InternalValue() *WorkspaceRunApply {
	var returns *WorkspaceRunApply
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) ManualConfirm() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"manualConfirm",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) ManualConfirmInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"manualConfirmInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) Retry() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"retry",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) RetryAttempts() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryAttempts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) RetryAttemptsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryAttemptsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) RetryBackoffMax() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMax",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) RetryBackoffMaxInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMaxInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) RetryBackoffMin() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMin",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) RetryBackoffMinInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMinInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) RetryInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"retryInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) WaitForRun() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"waitForRun",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference) WaitForRunInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"waitForRunInput",
		&returns,
	)
	return returns
}


func NewWorkspaceRunApplyOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) WorkspaceRunApplyOutputReference {
	_init_.Initialize()

	if err := validateNewWorkspaceRunApplyOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_WorkspaceRunApplyOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunApplyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewWorkspaceRunApplyOutputReference_Override(w WorkspaceRunApplyOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunApplyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		w,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetInternalValue(val *WorkspaceRunApply) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetManualConfirm(val interface{}) {
	if err := j.validateSetManualConfirmParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"manualConfirm",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetRetry(val interface{}) {
	if err := j.validateSetRetryParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retry",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetRetryAttempts(val *float64) {
	if err := j.validateSetRetryAttemptsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryAttempts",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetRetryBackoffMax(val *float64) {
	if err := j.validateSetRetryBackoffMaxParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryBackoffMax",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetRetryBackoffMin(val *float64) {
	if err := j.validateSetRetryBackoffMinParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryBackoffMin",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunApplyOutputReference)SetWaitForRun(val interface{}) {
	if err := j.validateSetWaitForRunParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"waitForRun",
		val,
	)
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		w,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := w.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		w,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := w.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		w,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := w.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		w,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := w.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		w,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := w.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		w,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := w.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		w,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := w.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		w,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetStringAttribute(terraformAttribute *string) *string {
	if err := w.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		w,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := w.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		w,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		w,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
	if err := w.validateInterpolationForAttributeParameters(property); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		w,
		"interpolationForAttribute",
		[]interface{}{property},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) ResetRetry() {
	_jsii_.InvokeVoid(
		w,
		"resetRetry",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) ResetRetryAttempts() {
	_jsii_.InvokeVoid(
		w,
		"resetRetryAttempts",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) ResetRetryBackoffMax() {
	_jsii_.InvokeVoid(
		w,
		"resetRetryBackoffMax",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) ResetRetryBackoffMin() {
	_jsii_.InvokeVoid(
		w,
		"resetRetryBackoffMin",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) ResetWaitForRun() {
	_jsii_.InvokeVoid(
		w,
		"resetWaitForRun",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
	if err := w.validateResolveParameters(_context); err != nil {
		panic(err)
	}
	var returns interface{}

	_jsii_.Invoke(
		w,
		"resolve",
		[]interface{}{_context},
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunApplyOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		w,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

