package workspacerun

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/jsii"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/tfe/workspacerun/internal"
)

type WorkspaceRunDestroyOutputReference interface {
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
	InternalValue() *WorkspaceRunDestroy
	SetInternalValue(val *WorkspaceRunDestroy)
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

// The jsii proxy struct for WorkspaceRunDestroyOutputReference
type jsiiProxy_WorkspaceRunDestroyOutputReference struct {
	internal.Type__cdktfComplexObject
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) ComplexObjectIndex() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"complexObjectIndex",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) ComplexObjectIsFromSet() *bool {
	var returns *bool
	_jsii_.Get(
		j,
		"complexObjectIsFromSet",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) CreationStack() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"creationStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) InternalValue() *WorkspaceRunDestroy {
	var returns *WorkspaceRunDestroy
	_jsii_.Get(
		j,
		"internalValue",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) ManualConfirm() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"manualConfirm",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) ManualConfirmInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"manualConfirmInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) Retry() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"retry",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) RetryAttempts() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryAttempts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) RetryAttemptsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryAttemptsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) RetryBackoffMax() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMax",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) RetryBackoffMaxInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMaxInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) RetryBackoffMin() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMin",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) RetryBackoffMinInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"retryBackoffMinInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) RetryInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"retryInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) TerraformAttribute() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformAttribute",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) TerraformResource() cdktf.IInterpolatingParent {
	var returns cdktf.IInterpolatingParent
	_jsii_.Get(
		j,
		"terraformResource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) WaitForRun() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"waitForRun",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference) WaitForRunInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"waitForRunInput",
		&returns,
	)
	return returns
}


func NewWorkspaceRunDestroyOutputReference(terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) WorkspaceRunDestroyOutputReference {
	_init_.Initialize()

	if err := validateNewWorkspaceRunDestroyOutputReferenceParameters(terraformResource, terraformAttribute); err != nil {
		panic(err)
	}
	j := jsiiProxy_WorkspaceRunDestroyOutputReference{}

	_jsii_.Create(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunDestroyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		&j,
	)

	return &j
}

func NewWorkspaceRunDestroyOutputReference_Override(w WorkspaceRunDestroyOutputReference, terraformResource cdktf.IInterpolatingParent, terraformAttribute *string) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunDestroyOutputReference",
		[]interface{}{terraformResource, terraformAttribute},
		w,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetComplexObjectIndex(val interface{}) {
	if err := j.validateSetComplexObjectIndexParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIndex",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetComplexObjectIsFromSet(val *bool) {
	if err := j.validateSetComplexObjectIsFromSetParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"complexObjectIsFromSet",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetInternalValue(val *WorkspaceRunDestroy) {
	if err := j.validateSetInternalValueParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"internalValue",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetManualConfirm(val interface{}) {
	if err := j.validateSetManualConfirmParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"manualConfirm",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetRetry(val interface{}) {
	if err := j.validateSetRetryParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retry",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetRetryAttempts(val *float64) {
	if err := j.validateSetRetryAttemptsParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryAttempts",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetRetryBackoffMax(val *float64) {
	if err := j.validateSetRetryBackoffMaxParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryBackoffMax",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetRetryBackoffMin(val *float64) {
	if err := j.validateSetRetryBackoffMinParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"retryBackoffMin",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetTerraformAttribute(val *string) {
	if err := j.validateSetTerraformAttributeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformAttribute",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetTerraformResource(val cdktf.IInterpolatingParent) {
	if err := j.validateSetTerraformResourceParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"terraformResource",
		val,
	)
}

func (j *jsiiProxy_WorkspaceRunDestroyOutputReference)SetWaitForRun(val interface{}) {
	if err := j.validateSetWaitForRunParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"waitForRun",
		val,
	)
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) ComputeFqn() *string {
	var returns *string

	_jsii_.Invoke(
		w,
		"computeFqn",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetStringAttribute(terraformAttribute *string) *string {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) InterpolationAsList() cdktf.IResolvable {
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		w,
		"interpolationAsList",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) InterpolationForAttribute(property *string) cdktf.IResolvable {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) ResetRetry() {
	_jsii_.InvokeVoid(
		w,
		"resetRetry",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) ResetRetryAttempts() {
	_jsii_.InvokeVoid(
		w,
		"resetRetryAttempts",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) ResetRetryBackoffMax() {
	_jsii_.InvokeVoid(
		w,
		"resetRetryBackoffMax",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) ResetRetryBackoffMin() {
	_jsii_.InvokeVoid(
		w,
		"resetRetryBackoffMin",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) ResetWaitForRun() {
	_jsii_.InvokeVoid(
		w,
		"resetWaitForRun",
		nil, // no parameters
	)
}

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) Resolve(_context cdktf.IResolveContext) interface{} {
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

func (w *jsiiProxy_WorkspaceRunDestroyOutputReference) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		w,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

