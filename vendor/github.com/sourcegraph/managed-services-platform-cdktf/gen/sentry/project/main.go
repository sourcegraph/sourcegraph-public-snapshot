package project

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"@cdktf/provider-sentry.project.Project",
		reflect.TypeOf((*Project)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "color", GoGetter: "Color"},
			_jsii_.MemberProperty{JsiiProperty: "connection", GoGetter: "Connection"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "count", GoGetter: "Count"},
			_jsii_.MemberProperty{JsiiProperty: "defaultKey", GoGetter: "DefaultKey"},
			_jsii_.MemberProperty{JsiiProperty: "defaultKeyInput", GoGetter: "DefaultKeyInput"},
			_jsii_.MemberProperty{JsiiProperty: "defaultRules", GoGetter: "DefaultRules"},
			_jsii_.MemberProperty{JsiiProperty: "defaultRulesInput", GoGetter: "DefaultRulesInput"},
			_jsii_.MemberProperty{JsiiProperty: "dependsOn", GoGetter: "DependsOn"},
			_jsii_.MemberProperty{JsiiProperty: "digestsMaxDelay", GoGetter: "DigestsMaxDelay"},
			_jsii_.MemberProperty{JsiiProperty: "digestsMaxDelayInput", GoGetter: "DigestsMaxDelayInput"},
			_jsii_.MemberProperty{JsiiProperty: "digestsMinDelay", GoGetter: "DigestsMinDelay"},
			_jsii_.MemberProperty{JsiiProperty: "digestsMinDelayInput", GoGetter: "DigestsMinDelayInput"},
			_jsii_.MemberProperty{JsiiProperty: "features", GoGetter: "Features"},
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
			_jsii_.MemberProperty{JsiiProperty: "internalId", GoGetter: "InternalId"},
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "isBookmarked", GoGetter: "IsBookmarked"},
			_jsii_.MemberProperty{JsiiProperty: "isPublic", GoGetter: "IsPublic"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "nameInput", GoGetter: "NameInput"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "organization", GoGetter: "Organization"},
			_jsii_.MemberProperty{JsiiProperty: "organizationInput", GoGetter: "OrganizationInput"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "platform", GoGetter: "Platform"},
			_jsii_.MemberProperty{JsiiProperty: "platformInput", GoGetter: "PlatformInput"},
			_jsii_.MemberProperty{JsiiProperty: "projectId", GoGetter: "ProjectId"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "provisioners", GoGetter: "Provisioners"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetDefaultKey", GoMethod: "ResetDefaultKey"},
			_jsii_.MemberMethod{JsiiMethod: "resetDefaultRules", GoMethod: "ResetDefaultRules"},
			_jsii_.MemberMethod{JsiiMethod: "resetDigestsMaxDelay", GoMethod: "ResetDigestsMaxDelay"},
			_jsii_.MemberMethod{JsiiMethod: "resetDigestsMinDelay", GoMethod: "ResetDigestsMinDelay"},
			_jsii_.MemberMethod{JsiiMethod: "resetId", GoMethod: "ResetId"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "resetPlatform", GoMethod: "ResetPlatform"},
			_jsii_.MemberMethod{JsiiMethod: "resetResolveAge", GoMethod: "ResetResolveAge"},
			_jsii_.MemberMethod{JsiiMethod: "resetSlug", GoMethod: "ResetSlug"},
			_jsii_.MemberMethod{JsiiMethod: "resetTeam", GoMethod: "ResetTeam"},
			_jsii_.MemberMethod{JsiiMethod: "resetTeams", GoMethod: "ResetTeams"},
			_jsii_.MemberProperty{JsiiProperty: "resolveAge", GoGetter: "ResolveAge"},
			_jsii_.MemberProperty{JsiiProperty: "resolveAgeInput", GoGetter: "ResolveAgeInput"},
			_jsii_.MemberProperty{JsiiProperty: "slug", GoGetter: "Slug"},
			_jsii_.MemberProperty{JsiiProperty: "slugInput", GoGetter: "SlugInput"},
			_jsii_.MemberProperty{JsiiProperty: "status", GoGetter: "Status"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "team", GoGetter: "Team"},
			_jsii_.MemberProperty{JsiiProperty: "teamInput", GoGetter: "TeamInput"},
			_jsii_.MemberProperty{JsiiProperty: "teams", GoGetter: "Teams"},
			_jsii_.MemberProperty{JsiiProperty: "teamsInput", GoGetter: "TeamsInput"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_Project{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfTerraformResource)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-sentry.project.ProjectConfig",
		reflect.TypeOf((*ProjectConfig)(nil)).Elem(),
	)
}
