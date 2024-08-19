package password

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/random/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/random/password/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/random/3.5.1/docs/resources/password random_password}.
type Password interface {
	cdktf.TerraformResource
	BcryptHash() *string
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	Id() *string
	Keepers() *map[string]*string
	SetKeepers(val *map[string]*string)
	KeepersInput() *map[string]*string
	Length() *float64
	SetLength(val *float64)
	LengthInput() *float64
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Lower() interface{}
	SetLower(val interface{})
	LowerInput() interface{}
	MinLower() *float64
	SetMinLower(val *float64)
	MinLowerInput() *float64
	MinNumeric() *float64
	SetMinNumeric(val *float64)
	MinNumericInput() *float64
	MinSpecial() *float64
	SetMinSpecial(val *float64)
	MinSpecialInput() *float64
	MinUpper() *float64
	SetMinUpper(val *float64)
	MinUpperInput() *float64
	// The tree node.
	Node() constructs.Node
	Number() interface{}
	SetNumber(val interface{})
	NumberInput() interface{}
	Numeric() interface{}
	SetNumeric(val interface{})
	NumericInput() interface{}
	OverrideSpecial() *string
	SetOverrideSpecial(val *string)
	OverrideSpecialInput() *string
	// Experimental.
	Provider() cdktf.TerraformProvider
	// Experimental.
	SetProvider(val cdktf.TerraformProvider)
	// Experimental.
	Provisioners() *[]interface{}
	// Experimental.
	SetProvisioners(val *[]interface{})
	// Experimental.
	RawOverrides() interface{}
	Result() *string
	Special() interface{}
	SetSpecial(val interface{})
	SpecialInput() interface{}
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Upper() interface{}
	SetUpper(val interface{})
	UpperInput() interface{}
	// Experimental.
	AddOverride(path *string, value interface{})
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
	InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	ResetKeepers()
	ResetLower()
	ResetMinLower()
	ResetMinNumeric()
	ResetMinSpecial()
	ResetMinUpper()
	ResetNumber()
	ResetNumeric()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetOverrideSpecial()
	ResetSpecial()
	ResetUpper()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for Password
type jsiiProxy_Password struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_Password) BcryptHash() *string {
	var returns *string
	_jsii_.Get(
		j,
		"bcryptHash",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Keepers() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"keepers",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) KeepersInput() *map[string]*string {
	var returns *map[string]*string
	_jsii_.Get(
		j,
		"keepersInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Length() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"length",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) LengthInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"lengthInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Lower() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"lower",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) LowerInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"lowerInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinLower() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minLower",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinLowerInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minLowerInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinNumeric() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minNumeric",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinNumericInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minNumericInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinSpecial() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minSpecial",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinSpecialInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minSpecialInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinUpper() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minUpper",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) MinUpperInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"minUpperInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Number() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"number",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) NumberInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"numberInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Numeric() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"numeric",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) NumericInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"numericInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) OverrideSpecial() *string {
	var returns *string
	_jsii_.Get(
		j,
		"overrideSpecial",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) OverrideSpecialInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"overrideSpecialInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Result() *string {
	var returns *string
	_jsii_.Get(
		j,
		"result",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Special() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"special",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) SpecialInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"specialInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) Upper() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"upper",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Password) UpperInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"upperInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/random/3.5.1/docs/resources/password random_password} Resource.
func NewPassword(scope constructs.Construct, id *string, config *PasswordConfig) Password {
	_init_.Initialize()

	if err := validateNewPasswordParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_Password{}

	_jsii_.Create(
		"@cdktf/provider-random.password.Password",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/random/3.5.1/docs/resources/password random_password} Resource.
func NewPassword_Override(p Password, scope constructs.Construct, id *string, config *PasswordConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-random.password.Password",
		[]interface{}{scope, id, config},
		p,
	)
}

func (j *jsiiProxy_Password)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_Password)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_Password)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_Password)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_Password)SetKeepers(val *map[string]*string) {
	if err := j.validateSetKeepersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"keepers",
		val,
	)
}

func (j *jsiiProxy_Password)SetLength(val *float64) {
	if err := j.validateSetLengthParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"length",
		val,
	)
}

func (j *jsiiProxy_Password)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_Password)SetLower(val interface{}) {
	if err := j.validateSetLowerParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lower",
		val,
	)
}

func (j *jsiiProxy_Password)SetMinLower(val *float64) {
	if err := j.validateSetMinLowerParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"minLower",
		val,
	)
}

func (j *jsiiProxy_Password)SetMinNumeric(val *float64) {
	if err := j.validateSetMinNumericParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"minNumeric",
		val,
	)
}

func (j *jsiiProxy_Password)SetMinSpecial(val *float64) {
	if err := j.validateSetMinSpecialParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"minSpecial",
		val,
	)
}

func (j *jsiiProxy_Password)SetMinUpper(val *float64) {
	if err := j.validateSetMinUpperParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"minUpper",
		val,
	)
}

func (j *jsiiProxy_Password)SetNumber(val interface{}) {
	if err := j.validateSetNumberParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"number",
		val,
	)
}

func (j *jsiiProxy_Password)SetNumeric(val interface{}) {
	if err := j.validateSetNumericParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"numeric",
		val,
	)
}

func (j *jsiiProxy_Password)SetOverrideSpecial(val *string) {
	if err := j.validateSetOverrideSpecialParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"overrideSpecial",
		val,
	)
}

func (j *jsiiProxy_Password)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_Password)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_Password)SetSpecial(val interface{}) {
	if err := j.validateSetSpecialParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"special",
		val,
	)
}

func (j *jsiiProxy_Password)SetUpper(val interface{}) {
	if err := j.validateSetUpperParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"upper",
		val,
	)
}

// Checks if `x` is a construct.
//
// Use this method instead of `instanceof` to properly detect `Construct`
// instances, even when the construct library is symlinked.
//
// Explanation: in JavaScript, multiple copies of the `constructs` library on
// disk are seen as independent, completely different libraries. As a
// consequence, the class `Construct` in each copy of the `constructs` library
// is seen as a different class, and an instance of one class will not test as
// `instanceof` the other class. `npm install` will not create installations
// like this, but users may manually symlink construct libraries together or
// use a monorepo tool: in those cases, multiple copies of the `constructs`
// library can be accidentally installed, and `instanceof` will behave
// unpredictably. It is safest to avoid using `instanceof`, and using
// this type-testing method instead.
//
// Returns: true if `x` is an object created from a class which extends `Construct`.
func Password_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePassword_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-random.password.Password",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Password_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePassword_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-random.password.Password",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Password_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePassword_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-random.password.Password",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func Password_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-random.password.Password",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (p *jsiiProxy_Password) AddOverride(path *string, value interface{}) {
	if err := p.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (p *jsiiProxy_Password) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := p.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := p.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := p.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		p,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := p.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		p,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := p.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		p,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := p.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		p,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := p.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		p,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetStringAttribute(terraformAttribute *string) *string {
	if err := p.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		p,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := p.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		p,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := p.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		p,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) OverrideLogicalId(newLogicalId *string) {
	if err := p.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (p *jsiiProxy_Password) ResetKeepers() {
	_jsii_.InvokeVoid(
		p,
		"resetKeepers",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetLower() {
	_jsii_.InvokeVoid(
		p,
		"resetLower",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetMinLower() {
	_jsii_.InvokeVoid(
		p,
		"resetMinLower",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetMinNumeric() {
	_jsii_.InvokeVoid(
		p,
		"resetMinNumeric",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetMinSpecial() {
	_jsii_.InvokeVoid(
		p,
		"resetMinSpecial",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetMinUpper() {
	_jsii_.InvokeVoid(
		p,
		"resetMinUpper",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetNumber() {
	_jsii_.InvokeVoid(
		p,
		"resetNumber",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetNumeric() {
	_jsii_.InvokeVoid(
		p,
		"resetNumeric",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		p,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetOverrideSpecial() {
	_jsii_.InvokeVoid(
		p,
		"resetOverrideSpecial",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetSpecial() {
	_jsii_.InvokeVoid(
		p,
		"resetSpecial",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) ResetUpper() {
	_jsii_.InvokeVoid(
		p,
		"resetUpper",
		nil, // no parameters
	)
}

func (p *jsiiProxy_Password) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		p,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		p,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_Password) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		p,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

