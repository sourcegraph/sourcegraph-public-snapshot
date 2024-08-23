// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Cloud Development Kit for Terraform
package cdktf

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterEnum(
		"cdktf.AnnotationMetadataEntryType",
		reflect.TypeOf((*AnnotationMetadataEntryType)(nil)).Elem(),
		map[string]interface{}{
			"INFO": AnnotationMetadataEntryType_INFO,
			"WARN": AnnotationMetadataEntryType_WARN,
			"ERROR": AnnotationMetadataEntryType_ERROR,
		},
	)
	_jsii_.RegisterClass(
		"cdktf.Annotations",
		reflect.TypeOf((*Annotations)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addError", GoMethod: "AddError"},
			_jsii_.MemberMethod{JsiiMethod: "addInfo", GoMethod: "AddInfo"},
			_jsii_.MemberMethod{JsiiMethod: "addWarning", GoMethod: "AddWarning"},
		},
		func() interface{} {
			return &jsiiProxy_Annotations{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.AnyListList",
		reflect.TypeOf((*AnyListList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "allWithMapKey", GoMethod: "AllWithMapKey"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_AnyListList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.AnyListMap",
		reflect.TypeOf((*AnyListMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_AnyListMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexMap)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.AnyMap",
		reflect.TypeOf((*AnyMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "lookup", GoMethod: "Lookup"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_AnyMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.AnyMapList",
		reflect.TypeOf((*AnyMapList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_AnyMapList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_MapList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.App",
		reflect.TypeOf((*App)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "crossStackReference", GoMethod: "CrossStackReference"},
			_jsii_.MemberProperty{JsiiProperty: "hclOutput", GoGetter: "HclOutput"},
			_jsii_.MemberProperty{JsiiProperty: "manifest", GoGetter: "Manifest"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "outdir", GoGetter: "Outdir"},
			_jsii_.MemberProperty{JsiiProperty: "skipBackendValidation", GoGetter: "SkipBackendValidation"},
			_jsii_.MemberProperty{JsiiProperty: "skipValidation", GoGetter: "SkipValidation"},
			_jsii_.MemberMethod{JsiiMethod: "synth", GoMethod: "Synth"},
			_jsii_.MemberProperty{JsiiProperty: "targetStackId", GoGetter: "TargetStackId"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_App{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.AppConfig",
		reflect.TypeOf((*AppConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.Aspects",
		reflect.TypeOf((*Aspects)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "add", GoMethod: "Add"},
			_jsii_.MemberProperty{JsiiProperty: "all", GoGetter: "All"},
		},
		func() interface{} {
			return &jsiiProxy_Aspects{}
		},
	)
	_jsii_.RegisterEnum(
		"cdktf.AssetType",
		reflect.TypeOf((*AssetType)(nil)).Elem(),
		map[string]interface{}{
			"FILE": AssetType_FILE,
			"DIRECTORY": AssetType_DIRECTORY,
			"ARCHIVE": AssetType_ARCHIVE,
		},
	)
	_jsii_.RegisterClass(
		"cdktf.AzurermBackend",
		reflect.TypeOf((*AzurermBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_AzurermBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.AzurermBackendConfig",
		reflect.TypeOf((*AzurermBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.BooleanList",
		reflect.TypeOf((*BooleanList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "allWithMapKey", GoMethod: "AllWithMapKey"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_BooleanList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.BooleanListList",
		reflect.TypeOf((*BooleanListList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "allWithMapKey", GoMethod: "AllWithMapKey"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_BooleanListList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.BooleanListMap",
		reflect.TypeOf((*BooleanListMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_BooleanListMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexMap)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.BooleanMap",
		reflect.TypeOf((*BooleanMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "lookup", GoMethod: "Lookup"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_BooleanMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.BooleanMapList",
		reflect.TypeOf((*BooleanMapList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_BooleanMapList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_MapList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.CloudBackend",
		reflect.TypeOf((*CloudBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_CloudBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.CloudBackendConfig",
		reflect.TypeOf((*CloudBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.CloudWorkspace",
		reflect.TypeOf((*CloudWorkspace)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			return &jsiiProxy_CloudWorkspace{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.ComplexComputedList",
		reflect.TypeOf((*ComplexComputedList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "complexComputedListIndex", GoGetter: "ComplexComputedListIndex"},
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
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_ComplexComputedList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IInterpolatingParent)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.ComplexList",
		reflect.TypeOf((*ComplexList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "allWithMapKey", GoMethod: "AllWithMapKey"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_ComplexList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.ComplexMap",
		reflect.TypeOf((*ComplexMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_ComplexMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.ComplexObject",
		reflect.TypeOf((*ComplexObject)(nil)).Elem(),
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
			_jsii_.MemberMethod{JsiiMethod: "interpolationAsList", GoMethod: "InterpolationAsList"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_ComplexObject{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IInterpolatingParent)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.ConsulBackend",
		reflect.TypeOf((*ConsulBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_ConsulBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.ConsulBackendConfig",
		reflect.TypeOf((*ConsulBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.CosBackend",
		reflect.TypeOf((*CosBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_CosBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.CosBackendAssumeRole",
		reflect.TypeOf((*CosBackendAssumeRole)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.CosBackendConfig",
		reflect.TypeOf((*CosBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.DataConfig",
		reflect.TypeOf((*DataConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataResource",
		reflect.TypeOf((*DataResource)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addMoveTarget", GoMethod: "AddMoveTarget"},
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "connection", GoGetter: "Connection"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "count", GoGetter: "Count"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
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
			_jsii_.MemberMethod{JsiiMethod: "hasResourceMove", GoMethod: "HasResourceMove"},
			_jsii_.MemberProperty{JsiiProperty: "id", GoGetter: "Id"},
			_jsii_.MemberMethod{JsiiMethod: "importFrom", GoMethod: "ImportFrom"},
			_jsii_.MemberProperty{JsiiProperty: "input", GoGetter: "Input"},
			_jsii_.MemberProperty{JsiiProperty: "inputInput", GoGetter: "InputInput"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberMethod{JsiiMethod: "moveFromId", GoMethod: "MoveFromId"},
			_jsii_.MemberMethod{JsiiMethod: "moveTo", GoMethod: "MoveTo"},
			_jsii_.MemberMethod{JsiiMethod: "moveToId", GoMethod: "MoveToId"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "output", GoGetter: "Output"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "provisioners", GoGetter: "Provisioners"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetInput", GoMethod: "ResetInput"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "resetTriggersReplace", GoMethod: "ResetTriggersReplace"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "triggersReplace", GoGetter: "TriggersReplace"},
			_jsii_.MemberProperty{JsiiProperty: "triggersReplaceInput", GoGetter: "TriggersReplaceInput"},
		},
		func() interface{} {
			j := jsiiProxy_DataResource{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformResource)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteState",
		reflect.TypeOf((*DataTerraformRemoteState)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteState{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateAzurerm",
		reflect.TypeOf((*DataTerraformRemoteStateAzurerm)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateAzurerm{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateAzurermConfig",
		reflect.TypeOf((*DataTerraformRemoteStateAzurermConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateConfig",
		reflect.TypeOf((*DataTerraformRemoteStateConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateConsul",
		reflect.TypeOf((*DataTerraformRemoteStateConsul)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateConsul{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateConsulConfig",
		reflect.TypeOf((*DataTerraformRemoteStateConsulConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateCos",
		reflect.TypeOf((*DataTerraformRemoteStateCos)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateCos{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateCosConfig",
		reflect.TypeOf((*DataTerraformRemoteStateCosConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateGcs",
		reflect.TypeOf((*DataTerraformRemoteStateGcs)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateGcs{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateGcsConfig",
		reflect.TypeOf((*DataTerraformRemoteStateGcsConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateHttp",
		reflect.TypeOf((*DataTerraformRemoteStateHttp)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateHttp{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateHttpConfig",
		reflect.TypeOf((*DataTerraformRemoteStateHttpConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateLocal",
		reflect.TypeOf((*DataTerraformRemoteStateLocal)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateLocal{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateLocalConfig",
		reflect.TypeOf((*DataTerraformRemoteStateLocalConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateOss",
		reflect.TypeOf((*DataTerraformRemoteStateOss)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateOss{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateOssConfig",
		reflect.TypeOf((*DataTerraformRemoteStateOssConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStatePg",
		reflect.TypeOf((*DataTerraformRemoteStatePg)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStatePg{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStatePgConfig",
		reflect.TypeOf((*DataTerraformRemoteStatePgConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateRemoteConfig",
		reflect.TypeOf((*DataTerraformRemoteStateRemoteConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateS3",
		reflect.TypeOf((*DataTerraformRemoteStateS3)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateS3{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateS3Config",
		reflect.TypeOf((*DataTerraformRemoteStateS3Config)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DataTerraformRemoteStateSwift",
		reflect.TypeOf((*DataTerraformRemoteStateSwift)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_DataTerraformRemoteStateSwift{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformRemoteState)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.DataTerraformRemoteStateSwiftConfig",
		reflect.TypeOf((*DataTerraformRemoteStateSwiftConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.DefaultTokenResolver",
		reflect.TypeOf((*DefaultTokenResolver)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "resolveList", GoMethod: "ResolveList"},
			_jsii_.MemberMethod{JsiiMethod: "resolveMap", GoMethod: "ResolveMap"},
			_jsii_.MemberMethod{JsiiMethod: "resolveNumberList", GoMethod: "ResolveNumberList"},
			_jsii_.MemberMethod{JsiiMethod: "resolveString", GoMethod: "ResolveString"},
			_jsii_.MemberMethod{JsiiMethod: "resolveToken", GoMethod: "ResolveToken"},
		},
		func() interface{} {
			j := jsiiProxy_DefaultTokenResolver{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITokenResolver)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.DynamicListTerraformIterator",
		reflect.TypeOf((*DynamicListTerraformIterator)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "dynamic", GoMethod: "Dynamic"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForList", GoMethod: "ForExpressionForList"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForMap", GoMethod: "ForExpressionForMap"},
			_jsii_.MemberMethod{JsiiMethod: "getAny", GoMethod: "GetAny"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMap", GoMethod: "GetAnyMap"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMap", GoMethod: "GetBooleanMap"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getMap", GoMethod: "GetMap"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberList", GoMethod: "GetNumberList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMap", GoMethod: "GetNumberMap"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMap", GoMethod: "GetStringMap"},
			_jsii_.MemberProperty{JsiiProperty: "key", GoGetter: "Key"},
			_jsii_.MemberMethod{JsiiMethod: "keys", GoMethod: "Keys"},
			_jsii_.MemberMethod{JsiiMethod: "pluckProperty", GoMethod: "PluckProperty"},
			_jsii_.MemberProperty{JsiiProperty: "value", GoGetter: "Value"},
			_jsii_.MemberMethod{JsiiMethod: "values", GoMethod: "Values"},
		},
		func() interface{} {
			j := jsiiProxy_DynamicListTerraformIterator{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_MapTerraformIterator)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.EncodingOptions",
		reflect.TypeOf((*EncodingOptions)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.FileProvisioner",
		reflect.TypeOf((*FileProvisioner)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.Fn",
		reflect.TypeOf((*Fn)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			j := jsiiProxy_Fn{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_FnGenerated)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.FnGenerated",
		reflect.TypeOf((*FnGenerated)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_FnGenerated{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.GcsBackend",
		reflect.TypeOf((*GcsBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_GcsBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.GcsBackendConfig",
		reflect.TypeOf((*GcsBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.HttpBackend",
		reflect.TypeOf((*HttpBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_HttpBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.HttpBackendConfig",
		reflect.TypeOf((*HttpBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterInterface(
		"cdktf.IAnyProducer",
		reflect.TypeOf((*IAnyProducer)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "produce", GoMethod: "Produce"},
		},
		func() interface{} {
			return &jsiiProxy_IAnyProducer{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IAspect",
		reflect.TypeOf((*IAspect)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "visit", GoMethod: "Visit"},
		},
		func() interface{} {
			return &jsiiProxy_IAspect{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IFragmentConcatenator",
		reflect.TypeOf((*IFragmentConcatenator)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "join", GoMethod: "Join"},
		},
		func() interface{} {
			return &jsiiProxy_IFragmentConcatenator{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IImportableConfig",
		reflect.TypeOf((*IImportableConfig)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "importId", GoGetter: "ImportId"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
		},
		func() interface{} {
			return &jsiiProxy_IImportableConfig{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IInterpolatingParent",
		reflect.TypeOf((*IInterpolatingParent)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
		},
		func() interface{} {
			return &jsiiProxy_IInterpolatingParent{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IListProducer",
		reflect.TypeOf((*IListProducer)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "produce", GoMethod: "Produce"},
		},
		func() interface{} {
			return &jsiiProxy_IListProducer{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IManifest",
		reflect.TypeOf((*IManifest)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "stacks", GoGetter: "Stacks"},
			_jsii_.MemberProperty{JsiiProperty: "version", GoGetter: "Version"},
		},
		func() interface{} {
			return &jsiiProxy_IManifest{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.INumberProducer",
		reflect.TypeOf((*INumberProducer)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "produce", GoMethod: "Produce"},
		},
		func() interface{} {
			return &jsiiProxy_INumberProducer{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IPostProcessor",
		reflect.TypeOf((*IPostProcessor)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "postProcess", GoMethod: "PostProcess"},
		},
		func() interface{} {
			return &jsiiProxy_IPostProcessor{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IRemoteWorkspace",
		reflect.TypeOf((*IRemoteWorkspace)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_IRemoteWorkspace{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IResolvable",
		reflect.TypeOf((*IResolvable)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			return &jsiiProxy_IResolvable{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IResolveContext",
		reflect.TypeOf((*IResolveContext)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "ignoreEscapes", GoGetter: "IgnoreEscapes"},
			_jsii_.MemberProperty{JsiiProperty: "iteratorContext", GoGetter: "IteratorContext"},
			_jsii_.MemberProperty{JsiiProperty: "preparing", GoGetter: "Preparing"},
			_jsii_.MemberMethod{JsiiMethod: "registerPostProcessor", GoMethod: "RegisterPostProcessor"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "scope", GoGetter: "Scope"},
			_jsii_.MemberProperty{JsiiProperty: "suppressBraces", GoGetter: "SuppressBraces"},
			_jsii_.MemberProperty{JsiiProperty: "warnEscapes", GoGetter: "WarnEscapes"},
		},
		func() interface{} {
			return &jsiiProxy_IResolveContext{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IResource",
		reflect.TypeOf((*IResource)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "stack", GoGetter: "Stack"},
		},
		func() interface{} {
			j := jsiiProxy_IResource{}
			_jsii_.InitJsiiProxy(&j.Type__constructsIConstruct)
			return &j
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IResourceConstructor",
		reflect.TypeOf((*IResourceConstructor)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_IResourceConstructor{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IScopeCallback",
		reflect.TypeOf((*IScopeCallback)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_IScopeCallback{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IStackSynthesizer",
		reflect.TypeOf((*IStackSynthesizer)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "synthesize", GoMethod: "Synthesize"},
		},
		func() interface{} {
			return &jsiiProxy_IStackSynthesizer{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.IStringProducer",
		reflect.TypeOf((*IStringProducer)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "produce", GoMethod: "Produce"},
		},
		func() interface{} {
			return &jsiiProxy_IStringProducer{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.ISynthesisSession",
		reflect.TypeOf((*ISynthesisSession)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "manifest", GoGetter: "Manifest"},
			_jsii_.MemberProperty{JsiiProperty: "outdir", GoGetter: "Outdir"},
			_jsii_.MemberProperty{JsiiProperty: "skipValidation", GoGetter: "SkipValidation"},
		},
		func() interface{} {
			return &jsiiProxy_ISynthesisSession{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.ITerraformAddressable",
		reflect.TypeOf((*ITerraformAddressable)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
		},
		func() interface{} {
			return &jsiiProxy_ITerraformAddressable{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.ITerraformDependable",
		reflect.TypeOf((*ITerraformDependable)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
		},
		func() interface{} {
			j := jsiiProxy_ITerraformDependable{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.ITerraformIterator",
		reflect.TypeOf((*ITerraformIterator)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_ITerraformIterator{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.ITerraformResource",
		reflect.TypeOf((*ITerraformResource)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "count", GoGetter: "Count"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
			_jsii_.MemberProperty{JsiiProperty: "forEach", GoGetter: "ForEach"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
		},
		func() interface{} {
			return &jsiiProxy_ITerraformResource{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.ITokenMapper",
		reflect.TypeOf((*ITokenMapper)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "mapToken", GoMethod: "MapToken"},
		},
		func() interface{} {
			return &jsiiProxy_ITokenMapper{}
		},
	)
	_jsii_.RegisterInterface(
		"cdktf.ITokenResolver",
		reflect.TypeOf((*ITokenResolver)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "resolveList", GoMethod: "ResolveList"},
			_jsii_.MemberMethod{JsiiMethod: "resolveMap", GoMethod: "ResolveMap"},
			_jsii_.MemberMethod{JsiiMethod: "resolveNumberList", GoMethod: "ResolveNumberList"},
			_jsii_.MemberMethod{JsiiMethod: "resolveString", GoMethod: "ResolveString"},
			_jsii_.MemberMethod{JsiiMethod: "resolveToken", GoMethod: "ResolveToken"},
		},
		func() interface{} {
			return &jsiiProxy_ITokenResolver{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.ImportableResource",
		reflect.TypeOf((*ImportableResource)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_ImportableResource{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.Lazy",
		reflect.TypeOf((*Lazy)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_Lazy{}
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.LazyAnyValueOptions",
		reflect.TypeOf((*LazyAnyValueOptions)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.LazyBase",
		reflect.TypeOf((*LazyBase)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addPostProcessor", GoMethod: "AddPostProcessor"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberMethod{JsiiMethod: "resolveLazy", GoMethod: "ResolveLazy"},
			_jsii_.MemberMethod{JsiiMethod: "toJSON", GoMethod: "ToJSON"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_LazyBase{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.LazyListValueOptions",
		reflect.TypeOf((*LazyListValueOptions)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.LazyStringValueOptions",
		reflect.TypeOf((*LazyStringValueOptions)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.ListTerraformIterator",
		reflect.TypeOf((*ListTerraformIterator)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "dynamic", GoMethod: "Dynamic"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForList", GoMethod: "ForExpressionForList"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForMap", GoMethod: "ForExpressionForMap"},
			_jsii_.MemberMethod{JsiiMethod: "getAny", GoMethod: "GetAny"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMap", GoMethod: "GetAnyMap"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMap", GoMethod: "GetBooleanMap"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getMap", GoMethod: "GetMap"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberList", GoMethod: "GetNumberList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMap", GoMethod: "GetNumberMap"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMap", GoMethod: "GetStringMap"},
			_jsii_.MemberProperty{JsiiProperty: "key", GoGetter: "Key"},
			_jsii_.MemberMethod{JsiiMethod: "keys", GoMethod: "Keys"},
			_jsii_.MemberMethod{JsiiMethod: "pluckProperty", GoMethod: "PluckProperty"},
			_jsii_.MemberProperty{JsiiProperty: "value", GoGetter: "Value"},
			_jsii_.MemberMethod{JsiiMethod: "values", GoMethod: "Values"},
		},
		func() interface{} {
			j := jsiiProxy_ListTerraformIterator{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformIterator)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.LocalBackend",
		reflect.TypeOf((*LocalBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_LocalBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.LocalBackendConfig",
		reflect.TypeOf((*LocalBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.LocalExecProvisioner",
		reflect.TypeOf((*LocalExecProvisioner)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.Manifest",
		reflect.TypeOf((*Manifest)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "buildManifest", GoMethod: "BuildManifest"},
			_jsii_.MemberMethod{JsiiMethod: "forStack", GoMethod: "ForStack"},
			_jsii_.MemberProperty{JsiiProperty: "hclOutput", GoGetter: "HclOutput"},
			_jsii_.MemberProperty{JsiiProperty: "outdir", GoGetter: "Outdir"},
			_jsii_.MemberProperty{JsiiProperty: "stackFileName", GoGetter: "StackFileName"},
			_jsii_.MemberProperty{JsiiProperty: "stacks", GoGetter: "Stacks"},
			_jsii_.MemberProperty{JsiiProperty: "version", GoGetter: "Version"},
			_jsii_.MemberMethod{JsiiMethod: "writeToFile", GoMethod: "WriteToFile"},
		},
		func() interface{} {
			j := jsiiProxy_Manifest{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IManifest)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.MapList",
		reflect.TypeOf((*MapList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_MapList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IInterpolatingParent)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.MapTerraformIterator",
		reflect.TypeOf((*MapTerraformIterator)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "dynamic", GoMethod: "Dynamic"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForList", GoMethod: "ForExpressionForList"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForMap", GoMethod: "ForExpressionForMap"},
			_jsii_.MemberMethod{JsiiMethod: "getAny", GoMethod: "GetAny"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMap", GoMethod: "GetAnyMap"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMap", GoMethod: "GetBooleanMap"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getMap", GoMethod: "GetMap"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberList", GoMethod: "GetNumberList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMap", GoMethod: "GetNumberMap"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMap", GoMethod: "GetStringMap"},
			_jsii_.MemberProperty{JsiiProperty: "key", GoGetter: "Key"},
			_jsii_.MemberMethod{JsiiMethod: "keys", GoMethod: "Keys"},
			_jsii_.MemberMethod{JsiiMethod: "pluckProperty", GoMethod: "PluckProperty"},
			_jsii_.MemberProperty{JsiiProperty: "value", GoGetter: "Value"},
			_jsii_.MemberMethod{JsiiMethod: "values", GoMethod: "Values"},
		},
		func() interface{} {
			j := jsiiProxy_MapTerraformIterator{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformIterator)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.MigrateIds",
		reflect.TypeOf((*MigrateIds)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "visit", GoMethod: "Visit"},
		},
		func() interface{} {
			j := jsiiProxy_MigrateIds{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IAspect)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.NamedCloudWorkspace",
		reflect.TypeOf((*NamedCloudWorkspace)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "project", GoGetter: "Project"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_NamedCloudWorkspace{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_CloudWorkspace)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.NamedRemoteWorkspace",
		reflect.TypeOf((*NamedRemoteWorkspace)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
		},
		func() interface{} {
			j := jsiiProxy_NamedRemoteWorkspace{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IRemoteWorkspace)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.NumberListList",
		reflect.TypeOf((*NumberListList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "allWithMapKey", GoMethod: "AllWithMapKey"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_NumberListList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.NumberListMap",
		reflect.TypeOf((*NumberListMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_NumberListMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexMap)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.NumberMap",
		reflect.TypeOf((*NumberMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "lookup", GoMethod: "Lookup"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_NumberMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.NumberMapList",
		reflect.TypeOf((*NumberMapList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_NumberMapList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_MapList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.Op",
		reflect.TypeOf((*Op)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_Op{}
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.OssAssumeRole",
		reflect.TypeOf((*OssAssumeRole)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.OssBackend",
		reflect.TypeOf((*OssBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_OssBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.OssBackendConfig",
		reflect.TypeOf((*OssBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.PgBackend",
		reflect.TypeOf((*PgBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_PgBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.PgBackendConfig",
		reflect.TypeOf((*PgBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.Postcondition",
		reflect.TypeOf((*Postcondition)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.Precondition",
		reflect.TypeOf((*Precondition)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.PrefixedRemoteWorkspaces",
		reflect.TypeOf((*PrefixedRemoteWorkspaces)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "prefix", GoGetter: "Prefix"},
		},
		func() interface{} {
			j := jsiiProxy_PrefixedRemoteWorkspaces{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IRemoteWorkspace)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.RemoteBackend",
		reflect.TypeOf((*RemoteBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_RemoteBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.RemoteBackendConfig",
		reflect.TypeOf((*RemoteBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.RemoteExecProvisioner",
		reflect.TypeOf((*RemoteExecProvisioner)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.ResolveOptions",
		reflect.TypeOf((*ResolveOptions)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.Resource",
		reflect.TypeOf((*Resource)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "stack", GoGetter: "Stack"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_Resource{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResource)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.ResourceTerraformIterator",
		reflect.TypeOf((*ResourceTerraformIterator)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "dynamic", GoMethod: "Dynamic"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForList", GoMethod: "ForExpressionForList"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForMap", GoMethod: "ForExpressionForMap"},
			_jsii_.MemberMethod{JsiiMethod: "getAny", GoMethod: "GetAny"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMap", GoMethod: "GetAnyMap"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMap", GoMethod: "GetBooleanMap"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getMap", GoMethod: "GetMap"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberList", GoMethod: "GetNumberList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMap", GoMethod: "GetNumberMap"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMap", GoMethod: "GetStringMap"},
			_jsii_.MemberProperty{JsiiProperty: "key", GoGetter: "Key"},
			_jsii_.MemberMethod{JsiiMethod: "keys", GoMethod: "Keys"},
			_jsii_.MemberMethod{JsiiMethod: "pluckProperty", GoMethod: "PluckProperty"},
			_jsii_.MemberProperty{JsiiProperty: "value", GoGetter: "Value"},
			_jsii_.MemberMethod{JsiiMethod: "values", GoMethod: "Values"},
		},
		func() interface{} {
			j := jsiiProxy_ResourceTerraformIterator{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformIterator)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.S3Backend",
		reflect.TypeOf((*S3Backend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_S3Backend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.S3BackendAssumeRoleConfig",
		reflect.TypeOf((*S3BackendAssumeRoleConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.S3BackendAssumeRoleWithWebIdentityConfig",
		reflect.TypeOf((*S3BackendAssumeRoleWithWebIdentityConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.S3BackendConfig",
		reflect.TypeOf((*S3BackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.S3BackendEndpointConfig",
		reflect.TypeOf((*S3BackendEndpointConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.SSHProvisionerConnection",
		reflect.TypeOf((*SSHProvisionerConnection)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.StackAnnotation",
		reflect.TypeOf((*StackAnnotation)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.StackManifest",
		reflect.TypeOf((*StackManifest)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.StringConcat",
		reflect.TypeOf((*StringConcat)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "join", GoMethod: "Join"},
		},
		func() interface{} {
			j := jsiiProxy_StringConcat{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IFragmentConcatenator)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.StringListList",
		reflect.TypeOf((*StringListList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "allWithMapKey", GoMethod: "AllWithMapKey"},
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_StringListList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.StringListMap",
		reflect.TypeOf((*StringListMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_StringListMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ComplexMap)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.StringMap",
		reflect.TypeOf((*StringMap)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "lookup", GoMethod: "Lookup"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
		},
		func() interface{} {
			j := jsiiProxy_StringMap{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IResolvable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.StringMapList",
		reflect.TypeOf((*StringMapList)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "computeFqn", GoMethod: "ComputeFqn"},
			_jsii_.MemberProperty{JsiiProperty: "creationStack", GoGetter: "CreationStack"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberMethod{JsiiMethod: "resolve", GoMethod: "Resolve"},
			_jsii_.MemberProperty{JsiiProperty: "terraformAttribute", GoGetter: "TerraformAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResource", GoGetter: "TerraformResource"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "wrapsSet", GoGetter: "WrapsSet"},
		},
		func() interface{} {
			j := jsiiProxy_StringMapList{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_MapList)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.SwiftBackend",
		reflect.TypeOf((*SwiftBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_SwiftBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformBackend)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.SwiftBackendConfig",
		reflect.TypeOf((*SwiftBackendConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TaggedCloudWorkspaces",
		reflect.TypeOf((*TaggedCloudWorkspaces)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "project", GoGetter: "Project"},
			_jsii_.MemberProperty{JsiiProperty: "tags", GoGetter: "Tags"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TaggedCloudWorkspaces{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_CloudWorkspace)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformAsset",
		reflect.TypeOf((*TerraformAsset)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "assetHash", GoGetter: "AssetHash"},
			_jsii_.MemberProperty{JsiiProperty: "fileName", GoGetter: "FileName"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "path", GoGetter: "Path"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberProperty{JsiiProperty: "type", GoGetter: "Type"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformAsset{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformAssetConfig",
		reflect.TypeOf((*TerraformAssetConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformBackend",
		reflect.TypeOf((*TerraformBackend)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getRemoteStateDataSource", GoMethod: "GetRemoteStateDataSource"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformBackend{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformCondition",
		reflect.TypeOf((*TerraformCondition)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformCount",
		reflect.TypeOf((*TerraformCount)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "index", GoGetter: "Index"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			return &jsiiProxy_TerraformCount{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformDataSource",
		reflect.TypeOf((*TerraformDataSource)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "count", GoGetter: "Count"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
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
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformDataSource{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IInterpolatingParent)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformDependable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformResource)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformElement",
		reflect.TypeOf((*TerraformElement)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformElement{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformElementMetadata",
		reflect.TypeOf((*TerraformElementMetadata)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformHclModule",
		reflect.TypeOf((*TerraformHclModule)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberMethod{JsiiMethod: "addProvider", GoMethod: "AddProvider"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
			_jsii_.MemberProperty{JsiiProperty: "forEach", GoGetter: "ForEach"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForOutput", GoMethod: "InterpolationForOutput"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "providers", GoGetter: "Providers"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "set", GoMethod: "Set"},
			_jsii_.MemberProperty{JsiiProperty: "skipAssetCreationFromLocalModules", GoGetter: "SkipAssetCreationFromLocalModules"},
			_jsii_.MemberProperty{JsiiProperty: "source", GoGetter: "Source"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "variables", GoGetter: "Variables"},
			_jsii_.MemberProperty{JsiiProperty: "version", GoGetter: "Version"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformHclModule{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformModule)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformHclModuleConfig",
		reflect.TypeOf((*TerraformHclModuleConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformIterator",
		reflect.TypeOf((*TerraformIterator)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "dynamic", GoMethod: "Dynamic"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForList", GoMethod: "ForExpressionForList"},
			_jsii_.MemberMethod{JsiiMethod: "forExpressionForMap", GoMethod: "ForExpressionForMap"},
			_jsii_.MemberMethod{JsiiMethod: "getAny", GoMethod: "GetAny"},
			_jsii_.MemberMethod{JsiiMethod: "getAnyMap", GoMethod: "GetAnyMap"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getBooleanMap", GoMethod: "GetBooleanMap"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getMap", GoMethod: "GetMap"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberList", GoMethod: "GetNumberList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumberMap", GoMethod: "GetNumberMap"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberMethod{JsiiMethod: "getStringMap", GoMethod: "GetStringMap"},
			_jsii_.MemberMethod{JsiiMethod: "keys", GoMethod: "Keys"},
			_jsii_.MemberMethod{JsiiMethod: "pluckProperty", GoMethod: "PluckProperty"},
			_jsii_.MemberMethod{JsiiMethod: "values", GoMethod: "Values"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformIterator{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformIterator)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformLocal",
		reflect.TypeOf((*TerraformLocal)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "asAnyMap", GoGetter: "AsAnyMap"},
			_jsii_.MemberProperty{JsiiProperty: "asBoolean", GoGetter: "AsBoolean"},
			_jsii_.MemberProperty{JsiiProperty: "asBooleanMap", GoGetter: "AsBooleanMap"},
			_jsii_.MemberProperty{JsiiProperty: "asList", GoGetter: "AsList"},
			_jsii_.MemberProperty{JsiiProperty: "asNumber", GoGetter: "AsNumber"},
			_jsii_.MemberProperty{JsiiProperty: "asNumberMap", GoGetter: "AsNumberMap"},
			_jsii_.MemberProperty{JsiiProperty: "asString", GoGetter: "AsString"},
			_jsii_.MemberProperty{JsiiProperty: "asStringMap", GoGetter: "AsStringMap"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "expression", GoGetter: "Expression"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformLocal{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformMetaArguments",
		reflect.TypeOf((*TerraformMetaArguments)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformModule",
		reflect.TypeOf((*TerraformModule)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberMethod{JsiiMethod: "addProvider", GoMethod: "AddProvider"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
			_jsii_.MemberProperty{JsiiProperty: "forEach", GoGetter: "ForEach"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForOutput", GoMethod: "InterpolationForOutput"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "providers", GoGetter: "Providers"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "skipAssetCreationFromLocalModules", GoGetter: "SkipAssetCreationFromLocalModules"},
			_jsii_.MemberProperty{JsiiProperty: "source", GoGetter: "Source"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "version", GoGetter: "Version"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformModule{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformDependable)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformModuleConfig",
		reflect.TypeOf((*TerraformModuleConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformModuleProvider",
		reflect.TypeOf((*TerraformModuleProvider)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformModuleUserConfig",
		reflect.TypeOf((*TerraformModuleUserConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformOutput",
		reflect.TypeOf((*TerraformOutput)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
			_jsii_.MemberProperty{JsiiProperty: "description", GoGetter: "Description"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "precondition", GoGetter: "Precondition"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "sensitive", GoGetter: "Sensitive"},
			_jsii_.MemberProperty{JsiiProperty: "staticId", GoGetter: "StaticId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "value", GoGetter: "Value"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformOutput{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformOutputConfig",
		reflect.TypeOf((*TerraformOutputConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformProvider",
		reflect.TypeOf((*TerraformProvider)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "alias", GoGetter: "Alias"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberProperty{JsiiProperty: "metaAttributes", GoGetter: "MetaAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformProviderSource", GoGetter: "TerraformProviderSource"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformProvider{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformProviderConfig",
		reflect.TypeOf((*TerraformProviderConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformProviderGeneratorMetadata",
		reflect.TypeOf((*TerraformProviderGeneratorMetadata)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformRemoteState",
		reflect.TypeOf((*TerraformRemoteState)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberMethod{JsiiMethod: "get", GoMethod: "Get"},
			_jsii_.MemberMethod{JsiiMethod: "getBoolean", GoMethod: "GetBoolean"},
			_jsii_.MemberMethod{JsiiMethod: "getList", GoMethod: "GetList"},
			_jsii_.MemberMethod{JsiiMethod: "getNumber", GoMethod: "GetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "getString", GoMethod: "GetString"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformRemoteState{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformResource",
		reflect.TypeOf((*TerraformResource)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addMoveTarget", GoMethod: "AddMoveTarget"},
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "connection", GoGetter: "Connection"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "count", GoGetter: "Count"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
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
			_jsii_.MemberMethod{JsiiMethod: "hasResourceMove", GoMethod: "HasResourceMove"},
			_jsii_.MemberMethod{JsiiMethod: "importFrom", GoMethod: "ImportFrom"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberMethod{JsiiMethod: "moveFromId", GoMethod: "MoveFromId"},
			_jsii_.MemberMethod{JsiiMethod: "moveTo", GoMethod: "MoveTo"},
			_jsii_.MemberMethod{JsiiMethod: "moveToId", GoMethod: "MoveToId"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "provisioners", GoGetter: "Provisioners"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformResource{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_IInterpolatingParent)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformDependable)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformResource)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformResourceConfig",
		reflect.TypeOf((*TerraformResourceConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformResourceImport",
		reflect.TypeOf((*TerraformResourceImport)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformResourceLifecycle",
		reflect.TypeOf((*TerraformResourceLifecycle)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformResourceMoveById",
		reflect.TypeOf((*TerraformResourceMoveById)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformResourceMoveByTarget",
		reflect.TypeOf((*TerraformResourceMoveByTarget)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformResourceTargets",
		reflect.TypeOf((*TerraformResourceTargets)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addResourceTarget", GoMethod: "AddResourceTarget"},
			_jsii_.MemberMethod{JsiiMethod: "getResourceByTarget", GoMethod: "GetResourceByTarget"},
		},
		func() interface{} {
			return &jsiiProxy_TerraformResourceTargets{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformSelf",
		reflect.TypeOf((*TerraformSelf)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_TerraformSelf{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformStack",
		reflect.TypeOf((*TerraformStack)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addDependency", GoMethod: "AddDependency"},
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberMethod{JsiiMethod: "allocateLogicalId", GoMethod: "AllocateLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "allProviders", GoMethod: "AllProviders"},
			_jsii_.MemberProperty{JsiiProperty: "dependencies", GoGetter: "Dependencies"},
			_jsii_.MemberMethod{JsiiMethod: "dependsOn", GoMethod: "DependsOn"},
			_jsii_.MemberMethod{JsiiMethod: "ensureBackendExists", GoMethod: "EnsureBackendExists"},
			_jsii_.MemberMethod{JsiiMethod: "getLogicalId", GoMethod: "GetLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "hasResourceMove", GoMethod: "HasResourceMove"},
			_jsii_.MemberProperty{JsiiProperty: "moveTargets", GoGetter: "MoveTargets"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "prepareStack", GoMethod: "PrepareStack"},
			_jsii_.MemberMethod{JsiiMethod: "registerIncomingCrossStackReference", GoMethod: "RegisterIncomingCrossStackReference"},
			_jsii_.MemberMethod{JsiiMethod: "registerOutgoingCrossStackReference", GoMethod: "RegisterOutgoingCrossStackReference"},
			_jsii_.MemberMethod{JsiiMethod: "runAllValidations", GoMethod: "RunAllValidations"},
			_jsii_.MemberProperty{JsiiProperty: "synthesizer", GoGetter: "Synthesizer"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformStack{}
			_jsii_.InitJsiiProxy(&j.Type__constructsConstruct)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformStackMetadata",
		reflect.TypeOf((*TerraformStackMetadata)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.TerraformVariable",
		reflect.TypeOf((*TerraformVariable)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberMethod{JsiiMethod: "addValidation", GoMethod: "AddValidation"},
			_jsii_.MemberProperty{JsiiProperty: "booleanValue", GoGetter: "BooleanValue"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "default", GoGetter: "Default"},
			_jsii_.MemberProperty{JsiiProperty: "description", GoGetter: "Description"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberProperty{JsiiProperty: "listValue", GoGetter: "ListValue"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "nullable", GoGetter: "Nullable"},
			_jsii_.MemberProperty{JsiiProperty: "numberValue", GoGetter: "NumberValue"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "sensitive", GoGetter: "Sensitive"},
			_jsii_.MemberProperty{JsiiProperty: "stringValue", GoGetter: "StringValue"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeHclAttributes", GoMethod: "SynthesizeHclAttributes"},
			_jsii_.MemberMethod{JsiiMethod: "toHclTerraform", GoMethod: "ToHclTerraform"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "type", GoGetter: "Type"},
			_jsii_.MemberProperty{JsiiProperty: "validation", GoGetter: "Validation"},
			_jsii_.MemberProperty{JsiiProperty: "value", GoGetter: "Value"},
		},
		func() interface{} {
			j := jsiiProxy_TerraformVariable{}
			_jsii_.InitJsiiProxy(&j.jsiiProxy_TerraformElement)
			_jsii_.InitJsiiProxy(&j.jsiiProxy_ITerraformAddressable)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformVariableConfig",
		reflect.TypeOf((*TerraformVariableConfig)(nil)).Elem(),
	)
	_jsii_.RegisterStruct(
		"cdktf.TerraformVariableValidationConfig",
		reflect.TypeOf((*TerraformVariableValidationConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.Testing",
		reflect.TypeOf((*Testing)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_Testing{}
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.TestingAppConfig",
		reflect.TypeOf((*TestingAppConfig)(nil)).Elem(),
	)
	_jsii_.RegisterClass(
		"cdktf.Token",
		reflect.TypeOf((*Token)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_Token{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.Tokenization",
		reflect.TypeOf((*Tokenization)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_Tokenization{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.TokenizedStringFragments",
		reflect.TypeOf((*TokenizedStringFragments)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addEscape", GoMethod: "AddEscape"},
			_jsii_.MemberMethod{JsiiMethod: "addIntrinsic", GoMethod: "AddIntrinsic"},
			_jsii_.MemberMethod{JsiiMethod: "addLiteral", GoMethod: "AddLiteral"},
			_jsii_.MemberMethod{JsiiMethod: "addToken", GoMethod: "AddToken"},
			_jsii_.MemberMethod{JsiiMethod: "concat", GoMethod: "Concat"},
			_jsii_.MemberProperty{JsiiProperty: "escapes", GoGetter: "Escapes"},
			_jsii_.MemberProperty{JsiiProperty: "firstToken", GoGetter: "FirstToken"},
			_jsii_.MemberProperty{JsiiProperty: "firstValue", GoGetter: "FirstValue"},
			_jsii_.MemberProperty{JsiiProperty: "intrinsic", GoGetter: "Intrinsic"},
			_jsii_.MemberMethod{JsiiMethod: "join", GoMethod: "Join"},
			_jsii_.MemberProperty{JsiiProperty: "length", GoGetter: "Length"},
			_jsii_.MemberProperty{JsiiProperty: "literals", GoGetter: "Literals"},
			_jsii_.MemberMethod{JsiiMethod: "mapTokens", GoMethod: "MapTokens"},
			_jsii_.MemberProperty{JsiiProperty: "tokens", GoGetter: "Tokens"},
		},
		func() interface{} {
			return &jsiiProxy_TokenizedStringFragments{}
		},
	)
	_jsii_.RegisterClass(
		"cdktf.VariableType",
		reflect.TypeOf((*VariableType)(nil)).Elem(),
		nil, // no members
		func() interface{} {
			return &jsiiProxy_VariableType{}
		},
	)
	_jsii_.RegisterStruct(
		"cdktf.WinrmProvisionerConnection",
		reflect.TypeOf((*WinrmProvisionerConnection)(nil)).Elem(),
	)
}
