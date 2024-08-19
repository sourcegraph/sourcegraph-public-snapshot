package serviceaccountkey

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"@cdktf/provider-google.serviceAccountKey.ServiceAccountKey",
		reflect.TypeOf((*ServiceAccountKey)(nil)).Elem(),
		[]_jsii_.Member{
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
			_jsii_.MemberProperty{JsiiProperty: "id", GoGetter: "Id"},
			_jsii_.MemberProperty{JsiiProperty: "idInput", GoGetter: "IdInput"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "keepers", GoGetter: "Keepers"},
			_jsii_.MemberProperty{JsiiProperty: "keepersInput", GoGetter: "KeepersInput"},
			_jsii_.MemberProperty{JsiiProperty: "keyAlgorithm", GoGetter: "KeyAlgorithm"},
			_jsii_.MemberProperty{JsiiProperty: "keyAlgorithmInput", GoGetter: "KeyAlgorithmInput"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "privateKey", GoGetter: "PrivateKey"},
			_jsii_.MemberProperty{JsiiProperty: "privateKeyType", GoGetter: "PrivateKeyType"},
			_jsii_.MemberProperty{JsiiProperty: "privateKeyTypeInput", GoGetter: "PrivateKeyTypeInput"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "provisioners", GoGetter: "Provisioners"},
			_jsii_.MemberProperty{JsiiProperty: "publicKey", GoGetter: "PublicKey"},
			_jsii_.MemberProperty{JsiiProperty: "publicKeyData", GoGetter: "PublicKeyData"},
			_jsii_.MemberProperty{JsiiProperty: "publicKeyDataInput", GoGetter: "PublicKeyDataInput"},
			_jsii_.MemberProperty{JsiiProperty: "publicKeyType", GoGetter: "PublicKeyType"},
			_jsii_.MemberProperty{JsiiProperty: "publicKeyTypeInput", GoGetter: "PublicKeyTypeInput"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetId", GoMethod: "ResetId"},
			_jsii_.MemberMethod{JsiiMethod: "resetKeepers", GoMethod: "ResetKeepers"},
			_jsii_.MemberMethod{JsiiMethod: "resetKeyAlgorithm", GoMethod: "ResetKeyAlgorithm"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "resetPrivateKeyType", GoMethod: "ResetPrivateKeyType"},
			_jsii_.MemberMethod{JsiiMethod: "resetPublicKeyData", GoMethod: "ResetPublicKeyData"},
			_jsii_.MemberMethod{JsiiMethod: "resetPublicKeyType", GoMethod: "ResetPublicKeyType"},
			_jsii_.MemberProperty{JsiiProperty: "serviceAccountId", GoGetter: "ServiceAccountId"},
			_jsii_.MemberProperty{JsiiProperty: "serviceAccountIdInput", GoGetter: "ServiceAccountIdInput"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "validAfter", GoGetter: "ValidAfter"},
			_jsii_.MemberProperty{JsiiProperty: "validBefore", GoGetter: "ValidBefore"},
		},
		func() interface{} {
			j := jsiiProxy_ServiceAccountKey{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfTerraformResource)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-google.serviceAccountKey.ServiceAccountKeyConfig",
		reflect.TypeOf((*ServiceAccountKeyConfig)(nil)).Elem(),
	)
}
