package conversation

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"@cdktf/provider-slack.conversation.Conversation",
		reflect.TypeOf((*Conversation)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberProperty{JsiiProperty: "actionOnDestroy", GoGetter: "ActionOnDestroy"},
			_jsii_.MemberProperty{JsiiProperty: "actionOnDestroyInput", GoGetter: "ActionOnDestroyInput"},
			_jsii_.MemberProperty{JsiiProperty: "actionOnUpdatePermanentMembers", GoGetter: "ActionOnUpdatePermanentMembers"},
			_jsii_.MemberProperty{JsiiProperty: "actionOnUpdatePermanentMembersInput", GoGetter: "ActionOnUpdatePermanentMembersInput"},
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "adoptExistingChannel", GoGetter: "AdoptExistingChannel"},
			_jsii_.MemberProperty{JsiiProperty: "adoptExistingChannelInput", GoGetter: "AdoptExistingChannelInput"},
			_jsii_.MemberProperty{JsiiProperty: "cdktfStack", GoGetter: "CdktfStack"},
			_jsii_.MemberProperty{JsiiProperty: "connection", GoGetter: "Connection"},
			_jsii_.MemberProperty{JsiiProperty: "constructNodeMetadata", GoGetter: "ConstructNodeMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "count", GoGetter: "Count"},
			_jsii_.MemberProperty{JsiiProperty: "created", GoGetter: "Created"},
			_jsii_.MemberProperty{JsiiProperty: "creator", GoGetter: "Creator"},
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
			_jsii_.MemberProperty{JsiiProperty: "isArchived", GoGetter: "IsArchived"},
			_jsii_.MemberProperty{JsiiProperty: "isArchivedInput", GoGetter: "IsArchivedInput"},
			_jsii_.MemberProperty{JsiiProperty: "isExtShared", GoGetter: "IsExtShared"},
			_jsii_.MemberProperty{JsiiProperty: "isGeneral", GoGetter: "IsGeneral"},
			_jsii_.MemberProperty{JsiiProperty: "isOrgShared", GoGetter: "IsOrgShared"},
			_jsii_.MemberProperty{JsiiProperty: "isPrivate", GoGetter: "IsPrivate"},
			_jsii_.MemberProperty{JsiiProperty: "isPrivateInput", GoGetter: "IsPrivateInput"},
			_jsii_.MemberProperty{JsiiProperty: "isShared", GoGetter: "IsShared"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberProperty{JsiiProperty: "name", GoGetter: "Name"},
			_jsii_.MemberProperty{JsiiProperty: "nameInput", GoGetter: "NameInput"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "permanentMembers", GoGetter: "PermanentMembers"},
			_jsii_.MemberProperty{JsiiProperty: "permanentMembersInput", GoGetter: "PermanentMembersInput"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "provisioners", GoGetter: "Provisioners"},
			_jsii_.MemberProperty{JsiiProperty: "purpose", GoGetter: "Purpose"},
			_jsii_.MemberProperty{JsiiProperty: "purposeInput", GoGetter: "PurposeInput"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetActionOnDestroy", GoMethod: "ResetActionOnDestroy"},
			_jsii_.MemberMethod{JsiiMethod: "resetActionOnUpdatePermanentMembers", GoMethod: "ResetActionOnUpdatePermanentMembers"},
			_jsii_.MemberMethod{JsiiMethod: "resetAdoptExistingChannel", GoMethod: "ResetAdoptExistingChannel"},
			_jsii_.MemberMethod{JsiiMethod: "resetId", GoMethod: "ResetId"},
			_jsii_.MemberMethod{JsiiMethod: "resetIsArchived", GoMethod: "ResetIsArchived"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "resetPermanentMembers", GoMethod: "ResetPermanentMembers"},
			_jsii_.MemberMethod{JsiiMethod: "resetPurpose", GoMethod: "ResetPurpose"},
			_jsii_.MemberMethod{JsiiMethod: "resetTopic", GoMethod: "ResetTopic"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "topic", GoGetter: "Topic"},
			_jsii_.MemberProperty{JsiiProperty: "topicInput", GoGetter: "TopicInput"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
		},
		func() interface{} {
			j := jsiiProxy_Conversation{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfTerraformResource)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-slack.conversation.ConversationConfig",
		reflect.TypeOf((*ConversationConfig)(nil)).Elem(),
	)
}
