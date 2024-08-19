package provider

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"@cdktf/provider-cloudflare.provider.CloudflareProvider",
		reflect.TypeOf((*CloudflareProvider)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "alias", GoGetter: "Alias"},
			_jsii_.MemberProperty{JsiiProperty: "aliasInput", GoGetter: "AliasInput"},
			_jsii_.MemberProperty{JsiiProperty: "apiBasePath", GoGetter: "ApiBasePath"},
			_jsii_.MemberProperty{JsiiProperty: "apiBasePathInput", GoGetter: "ApiBasePathInput"},
			_jsii_.MemberProperty{JsiiProperty: "apiClientLogging", GoGetter: "ApiClientLogging"},
			_jsii_.MemberProperty{JsiiProperty: "apiClientLoggingInput", GoGetter: "ApiClientLoggingInput"},
			_jsii_.MemberProperty{JsiiProperty: "apiHostname", GoGetter: "ApiHostname"},
			_jsii_.MemberProperty{JsiiProperty: "apiHostnameInput", GoGetter: "ApiHostnameInput"},
			_jsii_.MemberProperty{JsiiProperty: "apiKey", GoGetter: "ApiKey"},
			_jsii_.MemberProperty{JsiiProperty: "apiKeyInput", GoGetter: "ApiKeyInput"},
			_jsii_.MemberProperty{JsiiProperty: "apiToken", GoGetter: "ApiToken"},
			_jsii_.MemberProperty{JsiiProperty: "apiTokenInput", GoGetter: "ApiTokenInput"},
			_jsii_.MemberProperty{JsiiProperty: "apiUserServiceKey", GoGetter: "ApiUserServiceKey"},
			_jsii_.MemberProperty{JsiiProperty: "apiUserServiceKeyInput", GoGetter: "ApiUserServiceKeyInput"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "email", GoGetter: "Email"},
			_jsii_.MemberProperty{JsiiProperty: "emailInput", GoGetter: "EmailInput"},
			_jsii_.MemberProperty{JsiiProperty: "fqn", GoGetter: "Fqn"},
			_jsii_.MemberProperty{JsiiProperty: "friendlyUniqueId", GoGetter: "FriendlyUniqueId"},
			_jsii_.MemberProperty{JsiiProperty: "maxBackoff", GoGetter: "MaxBackoff"},
			_jsii_.MemberProperty{JsiiProperty: "maxBackoffInput", GoGetter: "MaxBackoffInput"},
			_jsii_.MemberProperty{JsiiProperty: "metaAttributes", GoGetter: "MetaAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "minBackoff", GoGetter: "MinBackoff"},
			_jsii_.MemberProperty{JsiiProperty: "minBackoffInput", GoGetter: "MinBackoffInput"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetAlias", GoMethod: "ResetAlias"},
			_jsii_.MemberMethod{JsiiMethod: "resetApiBasePath", GoMethod: "ResetApiBasePath"},
			_jsii_.MemberMethod{JsiiMethod: "resetApiClientLogging", GoMethod: "ResetApiClientLogging"},
			_jsii_.MemberMethod{JsiiMethod: "resetApiHostname", GoMethod: "ResetApiHostname"},
			_jsii_.MemberMethod{JsiiMethod: "resetApiKey", GoMethod: "ResetApiKey"},
			_jsii_.MemberMethod{JsiiMethod: "resetApiToken", GoMethod: "ResetApiToken"},
			_jsii_.MemberMethod{JsiiMethod: "resetApiUserServiceKey", GoMethod: "ResetApiUserServiceKey"},
			_jsii_.MemberMethod{JsiiMethod: "resetEmail", GoMethod: "ResetEmail"},
			_jsii_.MemberMethod{JsiiMethod: "resetMaxBackoff", GoMethod: "ResetMaxBackoff"},
			_jsii_.MemberMethod{JsiiMethod: "resetMinBackoff", GoMethod: "ResetMinBackoff"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "resetRetries", GoMethod: "ResetRetries"},
			_jsii_.MemberMethod{JsiiMethod: "resetRps", GoMethod: "ResetRps"},
			_jsii_.MemberProperty{JsiiProperty: "retries", GoGetter: "Retries"},
			_jsii_.MemberProperty{JsiiProperty: "retriesInput", GoGetter: "RetriesInput"},
			_jsii_.MemberProperty{JsiiProperty: "rps", GoGetter: "Rps"},
			_jsii_.MemberProperty{JsiiProperty: "rpsInput", GoGetter: "RpsInput"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformProviderSource", GoGetter: "TerraformProviderSource"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_CloudflareProvider{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfTerraformProvider)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-cloudflare.provider.CloudflareProviderConfig",
		reflect.TypeOf((*CloudflareProviderConfig)(nil)).Elem(),
	)
}
