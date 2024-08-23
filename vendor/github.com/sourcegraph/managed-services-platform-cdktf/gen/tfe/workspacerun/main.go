package workspacerun

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRun",
		reflect.TypeOf((*WorkspaceRun)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "apply", GoGetter: "Apply"},
			_jsii_.MemberProperty{JsiiProperty: "applyInput", GoGetter: "ApplyInput"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "connection", GoGetter: "Connection"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "count", GoGetter: "Count"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
			_jsii_.MemberProperty{JsiiProperty: "destroy", GoGetter: "Destroy"},
			_jsii_.MemberProperty{JsiiProperty: "destroyInput", GoGetter: "DestroyInput"},
			_jsii_.MemberProperty{JsiiProperty: "forEach", GoGetter: "ForEach"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMapAttribute", GoMethod: "GetAnyMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanAttribute", GoMethod: "GetBooleanAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMapAttribute", GoMethod: "GetBooleanMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getListAttribute", GoMethod: "GetListAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberAttribute", GoMethod: "GetNumberAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberListAttribute", GoMethod: "GetNumberListAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMapAttribute", GoMethod: "GetNumberMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getStringAttribute", GoMethod: "GetStringAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMapAttribute", GoMethod: "GetStringMapAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "id", GoGetter: "Id"},
			_jsii_.MemberProperty{JsiiProperty: "idInput", GoGetter: "IdInput"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "provisioners", GoGetter: "Provisioners"},
			_jsii_.MemberMethod{JsiiMethod: "putApply", GoMethod: "PutApply"},
			_jsii_.MemberMethod{JsiiMethod: "putDestroy", GoMethod: "PutDestroy"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetApply", GoMethod: "ResetApply"},
			_jsii_.MemberMethod{JsiiMethod: "resetDestroy", GoMethod: "ResetDestroy"},
			_jsii_.MemberMethod{JsiiMethod: "resetId", GoMethod: "ResetId"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "workspaceId", GoGetter: "WorkspaceId"},
			_jsii_.MemberProperty{JsiiProperty: "workspaceIdInput", GoGetter: "WorkspaceIdInput"},
		},
		func() interface{} {
			j := jsiiProxy_WorkspaceRun{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfTerraformResource)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunApply",
		reflect.TypeOf((*WorkspaceRunApply)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunApplyOutputReference",
		reflect.TypeOf((*WorkspaceRunApplyOutputReference)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "complexObjectIndex", GoGetter: "ComplexObjectIndex"},
			_jsii_.MemberProperty{JsiiProperty: "complexObjectIsFromSet", GoGetter: "ComplexObjectIsFromSet"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMapAttribute", GoMethod: "GetAnyMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanAttribute", GoMethod: "GetBooleanAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMapAttribute", GoMethod: "GetBooleanMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getListAttribute", GoMethod: "GetListAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberAttribute", GoMethod: "GetNumberAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberListAttribute", GoMethod: "GetNumberListAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMapAttribute", GoMethod: "GetNumberMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getStringAttribute", GoMethod: "GetStringAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMapAttribute", GoMethod: "GetStringMapAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "internalValue", GoGetter: "InternalValue"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationAsList", GoMethod: "InterpolationAsList"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "manualConfirm", GoGetter: "ManualConfirm"},
			_jsii_.MemberProperty{JsiiProperty: "manualConfirmInput", GoGetter: "ManualConfirmInput"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetry", GoMethod: "ResetRetry"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetryAttempts", GoMethod: "ResetRetryAttempts"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetryBackoffMax", GoMethod: "ResetRetryBackoffMax"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetryBackoffMin", GoMethod: "ResetRetryBackoffMin"},
			_jsii_.MemberMethod{JsiiMethod: "resetWaitForRun", GoMethod: "ResetWaitForRun"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "retry", GoGetter: "Retry"},
			_jsii_.MemberProperty{JsiiProperty: "retryAttempts", GoGetter: "RetryAttempts"},
			_jsii_.MemberProperty{JsiiProperty: "retryAttemptsInput", GoGetter: "RetryAttemptsInput"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMax", GoGetter: "RetryBackoffMax"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMaxInput", GoGetter: "RetryBackoffMaxInput"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMin", GoGetter: "RetryBackoffMin"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMinInput", GoGetter: "RetryBackoffMinInput"},
			_jsii_.MemberProperty{JsiiProperty: "retryInput", GoGetter: "RetryInput"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "waitForRun", GoGetter: "WaitForRun"},
			_jsii_.MemberProperty{JsiiProperty: "waitForRunInput", GoGetter: "WaitForRunInput"},
		},
		func() interface{} {
			j := jsiiProxy_WorkspaceRunApplyOutputReference{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfComplexObject)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunConfig",
		reflect.TypeOf((*WorkspaceRunConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunDestroy",
		reflect.TypeOf((*WorkspaceRunDestroy)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"@cdktf/provider-tfe.workspaceRun.WorkspaceRunDestroyOutputReference",
		reflect.TypeOf((*WorkspaceRunDestroyOutputReference)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "complexObjectIndex", GoGetter: "ComplexObjectIndex"},
			_jsii_.MemberProperty{JsiiProperty: "complexObjectIsFromSet", GoGetter: "ComplexObjectIsFromSet"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMapAttribute", GoMethod: "GetAnyMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanAttribute", GoMethod: "GetBooleanAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMapAttribute", GoMethod: "GetBooleanMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getListAttribute", GoMethod: "GetListAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberAttribute", GoMethod: "GetNumberAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberListAttribute", GoMethod: "GetNumberListAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMapAttribute", GoMethod: "GetNumberMapAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getStringAttribute", GoMethod: "GetStringAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMapAttribute", GoMethod: "GetStringMapAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "internalValue", GoGetter: "InternalValue"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationAsList", GoMethod: "InterpolationAsList"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "manualConfirm", GoGetter: "ManualConfirm"},
			_jsii_.MemberProperty{JsiiProperty: "manualConfirmInput", GoGetter: "ManualConfirmInput"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetry", GoMethod: "ResetRetry"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetryAttempts", GoMethod: "ResetRetryAttempts"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetryBackoffMax", GoMethod: "ResetRetryBackoffMax"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetryBackoffMin", GoMethod: "ResetRetryBackoffMin"},
			_jsii_.MemberMethod{JsiiMethod: "resetWaitForRun", GoMethod: "ResetWaitForRun"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "retry", GoGetter: "Retry"},
			_jsii_.MemberProperty{JsiiProperty: "retryAttempts", GoGetter: "RetryAttempts"},
			_jsii_.MemberProperty{JsiiProperty: "retryAttemptsInput", GoGetter: "RetryAttemptsInput"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMax", GoGetter: "RetryBackoffMax"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMaxInput", GoGetter: "RetryBackoffMaxInput"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMin", GoGetter: "RetryBackoffMin"},
			_jsii_.MemberProperty{JsiiProperty: "retryBackoffMinInput", GoGetter: "RetryBackoffMinInput"},
			_jsii_.MemberProperty{JsiiProperty: "retryInput", GoGetter: "RetryInput"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "waitForRun", GoGetter: "WaitForRun"},
			_jsii_.MemberProperty{JsiiProperty: "waitForRunInput", GoGetter: "WaitForRunInput"},
		},
		func() interface{} {
			j := jsiiProxy_WorkspaceRunDestroyOutputReference{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfComplexObject)
			return &j
		},
	)
}
