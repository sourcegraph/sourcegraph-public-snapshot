package provider

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/nobl9/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/nobl9/provider/internal"
)

// Represents a {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs nobl9}.
type Nobl9Provider interface {
	cdktf.TerraformProvider
	Alias() *string
	SetAlias(val *string)
	AliasInput() *string
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	ClientId() *string
	SetClientId(val *string)
	ClientIdInput() *string
	ClientSecret() *string
	SetClientSecret(val *string)
	ClientSecretInput() *string
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	IngestUrl() *string
	SetIngestUrl(val *string)
	IngestUrlInput() *string
	// Experimental.
	MetaAttributes() *map[string]interface{}
	// The tree node.
	Node() constructs.Node
	OktaAuthServer() *string
	SetOktaAuthServer(val *string)
	OktaAuthServerInput() *string
	OktaOrgUrl() *string
	SetOktaOrgUrl(val *string)
	OktaOrgUrlInput() *string
	Organization() *string
	SetOrganization(val *string)
	OrganizationInput() *string
	Project() *string
	SetProject(val *string)
	ProjectInput() *string
	// Experimental.
	RawOverrides() interface{}
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformProviderSource() *string
	// Experimental.
	TerraformResourceType() *string
	// Experimental.
	AddOverride(path *string, value interface{})
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	ResetAlias()
	ResetIngestUrl()
	ResetOktaAuthServer()
	ResetOktaOrgUrl()
	ResetOrganization()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for Nobl9Provider
type jsiiProxy_Nobl9Provider struct {
	internal.Type__cdktfTerraformProvider
}

func (j *jsiiProxy_Nobl9Provider) Alias() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alias",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) AliasInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"aliasInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) ClientId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) ClientIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) ClientSecret() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientSecret",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) ClientSecretInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"clientSecretInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) IngestUrl() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ingestUrl",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) IngestUrlInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ingestUrlInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) MetaAttributes() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"metaAttributes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) OktaAuthServer() *string {
	var returns *string
	_jsii_.Get(
		j,
		"oktaAuthServer",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) OktaAuthServerInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"oktaAuthServerInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) OktaOrgUrl() *string {
	var returns *string
	_jsii_.Get(
		j,
		"oktaOrgUrl",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) OktaOrgUrlInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"oktaOrgUrlInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) Organization() *string {
	var returns *string
	_jsii_.Get(
		j,
		"organization",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) OrganizationInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"organizationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) TerraformProviderSource() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformProviderSource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Nobl9Provider) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs nobl9} Resource.
func NewNobl9Provider(scope constructs.Construct, id *string, config *Nobl9ProviderConfig) Nobl9Provider {
	_init_.Initialize()

	if err := validateNewNobl9ProviderParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_Nobl9Provider{}

	_jsii_.Create(
		"@cdktf/provider-nobl9.provider.Nobl9Provider",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/nobl9/nobl9/0.22.0/docs nobl9} Resource.
func NewNobl9Provider_Override(n Nobl9Provider, scope constructs.Construct, id *string, config *Nobl9ProviderConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-nobl9.provider.Nobl9Provider",
		[]interface{}{scope, id, config},
		n,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetAlias(val *string) {
	_jsii_.Set(
		j,
		"alias",
		val,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetClientId(val *string) {
	_jsii_.Set(
		j,
		"clientId",
		val,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetClientSecret(val *string) {
	_jsii_.Set(
		j,
		"clientSecret",
		val,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetIngestUrl(val *string) {
	_jsii_.Set(
		j,
		"ingestUrl",
		val,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetOktaAuthServer(val *string) {
	_jsii_.Set(
		j,
		"oktaAuthServer",
		val,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetOktaOrgUrl(val *string) {
	_jsii_.Set(
		j,
		"oktaOrgUrl",
		val,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetOrganization(val *string) {
	_jsii_.Set(
		j,
		"organization",
		val,
	)
}

func (j *jsiiProxy_Nobl9Provider)SetProject(val *string) {
	_jsii_.Set(
		j,
		"project",
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
func Nobl9Provider_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateNobl9Provider_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-nobl9.provider.Nobl9Provider",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Nobl9Provider_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateNobl9Provider_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-nobl9.provider.Nobl9Provider",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Nobl9Provider_IsTerraformProvider(x interface{}) *bool {
	_init_.Initialize()

	if err := validateNobl9Provider_IsTerraformProviderParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-nobl9.provider.Nobl9Provider",
		"isTerraformProvider",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func Nobl9Provider_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-nobl9.provider.Nobl9Provider",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (n *jsiiProxy_Nobl9Provider) AddOverride(path *string, value interface{}) {
	if err := n.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		n,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (n *jsiiProxy_Nobl9Provider) OverrideLogicalId(newLogicalId *string) {
	if err := n.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		n,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (n *jsiiProxy_Nobl9Provider) ResetAlias() {
	_jsii_.InvokeVoid(
		n,
		"resetAlias",
		nil, // no parameters
	)
}

func (n *jsiiProxy_Nobl9Provider) ResetIngestUrl() {
	_jsii_.InvokeVoid(
		n,
		"resetIngestUrl",
		nil, // no parameters
	)
}

func (n *jsiiProxy_Nobl9Provider) ResetOktaAuthServer() {
	_jsii_.InvokeVoid(
		n,
		"resetOktaAuthServer",
		nil, // no parameters
	)
}

func (n *jsiiProxy_Nobl9Provider) ResetOktaOrgUrl() {
	_jsii_.InvokeVoid(
		n,
		"resetOktaOrgUrl",
		nil, // no parameters
	)
}

func (n *jsiiProxy_Nobl9Provider) ResetOrganization() {
	_jsii_.InvokeVoid(
		n,
		"resetOrganization",
		nil, // no parameters
	)
}

func (n *jsiiProxy_Nobl9Provider) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		n,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (n *jsiiProxy_Nobl9Provider) ResetProject() {
	_jsii_.InvokeVoid(
		n,
		"resetProject",
		nil, // no parameters
	)
}

func (n *jsiiProxy_Nobl9Provider) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		n,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (n *jsiiProxy_Nobl9Provider) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		n,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (n *jsiiProxy_Nobl9Provider) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		n,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (n *jsiiProxy_Nobl9Provider) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		n,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

