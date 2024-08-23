package password

import (
	"reflect"

	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
)

func init() {
	_jsii_.RegisterClass(
		"@cdktf/provider-random.password.Password",
		reflect.TypeOf((*Password)(nil)).Elem(),
		[]_jsii_.Member{
			_jsii_.MemberMethod{JsiiMethod: "addOverride", GoMethod: "AddOverride"},
			_jsii_.MemberProperty{JsiiProperty: "bcryptHash", GoGetter: "BcryptHash"},
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
			_jsii_.MemberMethod{JsiiMethod: "interpolationForAttribute", GoMethod: "InterpolationForAttribute"},
			_jsii_.MemberProperty{JsiiProperty: "keepers", GoGetter: "Keepers"},
			_jsii_.MemberProperty{JsiiProperty: "keepersInput", GoGetter: "KeepersInput"},
			_jsii_.MemberProperty{JsiiProperty: "length", GoGetter: "Length"},
			_jsii_.MemberProperty{JsiiProperty: "lengthInput", GoGetter: "LengthInput"},
			_jsii_.MemberProperty{JsiiProperty: "lifecycle", GoGetter: "Lifecycle"},
			_jsii_.MemberProperty{JsiiProperty: "lower", GoGetter: "Lower"},
			_jsii_.MemberProperty{JsiiProperty: "lowerInput", GoGetter: "LowerInput"},
			_jsii_.MemberProperty{JsiiProperty: "minLower", GoGetter: "MinLower"},
			_jsii_.MemberProperty{JsiiProperty: "minLowerInput", GoGetter: "MinLowerInput"},
			_jsii_.MemberProperty{JsiiProperty: "minNumeric", GoGetter: "MinNumeric"},
			_jsii_.MemberProperty{JsiiProperty: "minNumericInput", GoGetter: "MinNumericInput"},
			_jsii_.MemberProperty{JsiiProperty: "minSpecial", GoGetter: "MinSpecial"},
			_jsii_.MemberProperty{JsiiProperty: "minSpecialInput", GoGetter: "MinSpecialInput"},
			_jsii_.MemberProperty{JsiiProperty: "minUpper", GoGetter: "MinUpper"},
			_jsii_.MemberProperty{JsiiProperty: "minUpperInput", GoGetter: "MinUpperInput"},
			_jsii_.MemberProperty{JsiiProperty: "node", GoGetter: "Node"},
			_jsii_.MemberProperty{JsiiProperty: "number", GoGetter: "Number"},
			_jsii_.MemberProperty{JsiiProperty: "numberInput", GoGetter: "NumberInput"},
			_jsii_.MemberProperty{JsiiProperty: "numeric", GoGetter: "Numeric"},
			_jsii_.MemberProperty{JsiiProperty: "numericInput", GoGetter: "NumericInput"},
			_jsii_.MemberMethod{JsiiMethod: "overrideLogicalId", GoMethod: "OverrideLogicalId"},
			_jsii_.MemberProperty{JsiiProperty: "overrideSpecial", GoGetter: "OverrideSpecial"},
			_jsii_.MemberProperty{JsiiProperty: "overrideSpecialInput", GoGetter: "OverrideSpecialInput"},
			_jsii_.MemberProperty{JsiiProperty: "provider", GoGetter: "Provider"},
			_jsii_.MemberProperty{JsiiProperty: "provisioners", GoGetter: "Provisioners"},
			_jsii_.MemberProperty{JsiiProperty: "rawOverrides", GoGetter: "RawOverrides"},
			_jsii_.MemberMethod{JsiiMethod: "resetKeepers", GoMethod: "ResetKeepers"},
			_jsii_.MemberMethod{JsiiMethod: "resetLower", GoMethod: "ResetLower"},
			_jsii_.MemberMethod{JsiiMethod: "resetMinLower", GoMethod: "ResetMinLower"},
			_jsii_.MemberMethod{JsiiMethod: "resetMinNumeric", GoMethod: "ResetMinNumeric"},
			_jsii_.MemberMethod{JsiiMethod: "resetMinSpecial", GoMethod: "ResetMinSpecial"},
			_jsii_.MemberMethod{JsiiMethod: "resetMinUpper", GoMethod: "ResetMinUpper"},
			_jsii_.MemberMethod{JsiiMethod: "resetNumber", GoMethod: "ResetNumber"},
			_jsii_.MemberMethod{JsiiMethod: "resetNumeric", GoMethod: "ResetNumeric"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideLogicalId", GoMethod: "ResetOverrideLogicalId"},
			_jsii_.MemberMethod{JsiiMethod: "resetOverrideSpecial", GoMethod: "ResetOverrideSpecial"},
			_jsii_.MemberMethod{JsiiMethod: "resetSpecial", GoMethod: "ResetSpecial"},
			_jsii_.MemberMethod{JsiiMethod: "resetUpper", GoMethod: "ResetUpper"},
			_jsii_.MemberProperty{JsiiProperty: "result", GoGetter: "Result"},
			_jsii_.MemberProperty{JsiiProperty: "special", GoGetter: "Special"},
			_jsii_.MemberProperty{JsiiProperty: "specialInput", GoGetter: "SpecialInput"},
			_jsii_.MemberMethod{JsiiMethod: "synthesizeAttributes", GoMethod: "SynthesizeAttributes"},
			_jsii_.MemberProperty{JsiiProperty: "terraformGeneratorMetadata", GoGetter: "TerraformGeneratorMetadata"},
			_jsii_.MemberProperty{JsiiProperty: "terraformMetaArguments", GoGetter: "TerraformMetaArguments"},
			_jsii_.MemberProperty{JsiiProperty: "terraformResourceType", GoGetter: "TerraformResourceType"},
			_jsii_.MemberMethod{JsiiMethod: "toMetadata", GoMethod: "ToMetadata"},
			_jsii_.MemberMethod{JsiiMethod: "toString", GoMethod: "ToString"},
			_jsii_.MemberMethod{JsiiMethod: "toTerraform", GoMethod: "ToTerraform"},
			_jsii_.MemberProperty{JsiiProperty: "upper", GoGetter: "Upper"},
			_jsii_.MemberProperty{JsiiProperty: "upperInput", GoGetter: "UpperInput"},
		},
		func() interface{} {
			j := jsiiProxy_Password{}
			_jsii_.InitJsiiProxy(&j.Type__cdktfTerraformResource)
			return &j
		},
	)
	_jsii_.RegisterStruct(
		"@cdktf/provider-random.password.PasswordConfig",
		reflect.TypeOf((*PasswordConfig)(nil)).Elem(),
	)
}
