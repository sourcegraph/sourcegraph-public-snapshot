package computesubnetwork

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computesubnetwork/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork google_compute_subnetwork}.
type ComputeSubnetwork interface {
	cdktf.TerraformResource
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
	CreationTimestamp() *string
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Description() *string
	SetDescription(val *string)
	DescriptionInput() *string
	ExternalIpv6Prefix() *string
	SetExternalIpv6Prefix(val *string)
	ExternalIpv6PrefixInput() *string
	Fingerprint() *string
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	GatewayAddress() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	InternalIpv6Prefix() *string
	IpCidrRange() *string
	SetIpCidrRange(val *string)
	IpCidrRangeInput() *string
	Ipv6AccessType() *string
	SetIpv6AccessType(val *string)
	Ipv6AccessTypeInput() *string
	Ipv6CidrRange() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	LogConfig() ComputeSubnetworkLogConfigOutputReference
	LogConfigInput() *ComputeSubnetworkLogConfig
	Name() *string
	SetName(val *string)
	NameInput() *string
	Network() *string
	SetNetwork(val *string)
	NetworkInput() *string
	// The tree node.
	Node() constructs.Node
	PrivateIpGoogleAccess() interface{}
	SetPrivateIpGoogleAccess(val interface{})
	PrivateIpGoogleAccessInput() interface{}
	PrivateIpv6GoogleAccess() *string
	SetPrivateIpv6GoogleAccess(val *string)
	PrivateIpv6GoogleAccessInput() *string
	Project() *string
	SetProject(val *string)
	ProjectInput() *string
	// Experimental.
	Provider() cdktf.TerraformProvider
	// Experimental.
	SetProvider(val cdktf.TerraformProvider)
	// Experimental.
	Provisioners() *[]interface{}
	// Experimental.
	SetProvisioners(val *[]interface{})
	Purpose() *string
	SetPurpose(val *string)
	PurposeInput() *string
	// Experimental.
	RawOverrides() interface{}
	Region() *string
	SetRegion(val *string)
	RegionInput() *string
	ReservedInternalRange() *string
	SetReservedInternalRange(val *string)
	ReservedInternalRangeInput() *string
	Role() *string
	SetRole(val *string)
	RoleInput() *string
	SecondaryIpRange() ComputeSubnetworkSecondaryIpRangeList
	SecondaryIpRangeInput() interface{}
	SelfLink() *string
	StackType() *string
	SetStackType(val *string)
	StackTypeInput() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() ComputeSubnetworkTimeoutsOutputReference
	TimeoutsInput() interface{}
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
	PutLogConfig(value *ComputeSubnetworkLogConfig)
	PutSecondaryIpRange(value interface{})
	PutTimeouts(value *ComputeSubnetworkTimeouts)
	ResetDescription()
	ResetExternalIpv6Prefix()
	ResetId()
	ResetIpCidrRange()
	ResetIpv6AccessType()
	ResetLogConfig()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPrivateIpGoogleAccess()
	ResetPrivateIpv6GoogleAccess()
	ResetProject()
	ResetPurpose()
	ResetRegion()
	ResetReservedInternalRange()
	ResetRole()
	ResetSecondaryIpRange()
	ResetStackType()
	ResetTimeouts()
	SynthesizeAttributes() *map[string]interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for ComputeSubnetwork
type jsiiProxy_ComputeSubnetwork struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_ComputeSubnetwork) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) CreationTimestamp() *string {
	var returns *string
	_jsii_.Get(
		j,
		"creationTimestamp",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) ExternalIpv6Prefix() *string {
	var returns *string
	_jsii_.Get(
		j,
		"externalIpv6Prefix",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) ExternalIpv6PrefixInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"externalIpv6PrefixInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Fingerprint() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fingerprint",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) GatewayAddress() *string {
	var returns *string
	_jsii_.Get(
		j,
		"gatewayAddress",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) InternalIpv6Prefix() *string {
	var returns *string
	_jsii_.Get(
		j,
		"internalIpv6Prefix",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) IpCidrRange() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipCidrRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) IpCidrRangeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipCidrRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Ipv6AccessType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipv6AccessType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Ipv6AccessTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipv6AccessTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Ipv6CidrRange() *string {
	var returns *string
	_jsii_.Get(
		j,
		"ipv6CidrRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) LogConfig() ComputeSubnetworkLogConfigOutputReference {
	var returns ComputeSubnetworkLogConfigOutputReference
	_jsii_.Get(
		j,
		"logConfig",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) LogConfigInput() *ComputeSubnetworkLogConfig {
	var returns *ComputeSubnetworkLogConfig
	_jsii_.Get(
		j,
		"logConfigInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Network() *string {
	var returns *string
	_jsii_.Get(
		j,
		"network",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) NetworkInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"networkInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) PrivateIpGoogleAccess() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"privateIpGoogleAccess",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) PrivateIpGoogleAccessInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"privateIpGoogleAccessInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) PrivateIpv6GoogleAccess() *string {
	var returns *string
	_jsii_.Get(
		j,
		"privateIpv6GoogleAccess",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) PrivateIpv6GoogleAccessInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"privateIpv6GoogleAccessInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Purpose() *string {
	var returns *string
	_jsii_.Get(
		j,
		"purpose",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) PurposeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"purposeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Region() *string {
	var returns *string
	_jsii_.Get(
		j,
		"region",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) RegionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"regionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) ReservedInternalRange() *string {
	var returns *string
	_jsii_.Get(
		j,
		"reservedInternalRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) ReservedInternalRangeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"reservedInternalRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Role() *string {
	var returns *string
	_jsii_.Get(
		j,
		"role",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) RoleInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"roleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) SecondaryIpRange() ComputeSubnetworkSecondaryIpRangeList {
	var returns ComputeSubnetworkSecondaryIpRangeList
	_jsii_.Get(
		j,
		"secondaryIpRange",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) SecondaryIpRangeInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"secondaryIpRangeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) StackType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"stackType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) StackTypeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"stackTypeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) Timeouts() ComputeSubnetworkTimeoutsOutputReference {
	var returns ComputeSubnetworkTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeSubnetwork) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork google_compute_subnetwork} Resource.
func NewComputeSubnetwork(scope constructs.Construct, id *string, config *ComputeSubnetworkConfig) ComputeSubnetwork {
	_init_.Initialize()

	if err := validateNewComputeSubnetworkParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeSubnetwork{}

	_jsii_.Create(
		"@cdktf/provider-google.computeSubnetwork.ComputeSubnetwork",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_subnetwork google_compute_subnetwork} Resource.
func NewComputeSubnetwork_Override(c ComputeSubnetwork, scope constructs.Construct, id *string, config *ComputeSubnetworkConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeSubnetwork.ComputeSubnetwork",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetExternalIpv6Prefix(val *string) {
	if err := j.validateSetExternalIpv6PrefixParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"externalIpv6Prefix",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetIpCidrRange(val *string) {
	if err := j.validateSetIpCidrRangeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ipCidrRange",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetIpv6AccessType(val *string) {
	if err := j.validateSetIpv6AccessTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"ipv6AccessType",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetNetwork(val *string) {
	if err := j.validateSetNetworkParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"network",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetPrivateIpGoogleAccess(val interface{}) {
	if err := j.validateSetPrivateIpGoogleAccessParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"privateIpGoogleAccess",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetPrivateIpv6GoogleAccess(val *string) {
	if err := j.validateSetPrivateIpv6GoogleAccessParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"privateIpv6GoogleAccess",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetPurpose(val *string) {
	if err := j.validateSetPurposeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"purpose",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetRegion(val *string) {
	if err := j.validateSetRegionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"region",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetReservedInternalRange(val *string) {
	if err := j.validateSetReservedInternalRangeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"reservedInternalRange",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetRole(val *string) {
	if err := j.validateSetRoleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"role",
		val,
	)
}

func (j *jsiiProxy_ComputeSubnetwork)SetStackType(val *string) {
	if err := j.validateSetStackTypeParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"stackType",
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
func ComputeSubnetwork_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeSubnetwork_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeSubnetwork.ComputeSubnetwork",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeSubnetwork_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeSubnetwork_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeSubnetwork.ComputeSubnetwork",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeSubnetwork_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeSubnetwork_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeSubnetwork.ComputeSubnetwork",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func ComputeSubnetwork_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.computeSubnetwork.ComputeSubnetwork",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_ComputeSubnetwork) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := c.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := c.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := c.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		c,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := c.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		c,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := c.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		c,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := c.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		c,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := c.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		c,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetStringAttribute(terraformAttribute *string) *string {
	if err := c.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		c,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := c.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		c,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := c.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		c,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_ComputeSubnetwork) PutLogConfig(value *ComputeSubnetworkLogConfig) {
	if err := c.validatePutLogConfigParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putLogConfig",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSubnetwork) PutSecondaryIpRange(value interface{}) {
	if err := c.validatePutSecondaryIpRangeParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putSecondaryIpRange",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSubnetwork) PutTimeouts(value *ComputeSubnetworkTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetExternalIpv6Prefix() {
	_jsii_.InvokeVoid(
		c,
		"resetExternalIpv6Prefix",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetIpCidrRange() {
	_jsii_.InvokeVoid(
		c,
		"resetIpCidrRange",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetIpv6AccessType() {
	_jsii_.InvokeVoid(
		c,
		"resetIpv6AccessType",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetLogConfig() {
	_jsii_.InvokeVoid(
		c,
		"resetLogConfig",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetPrivateIpGoogleAccess() {
	_jsii_.InvokeVoid(
		c,
		"resetPrivateIpGoogleAccess",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetPrivateIpv6GoogleAccess() {
	_jsii_.InvokeVoid(
		c,
		"resetPrivateIpv6GoogleAccess",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetPurpose() {
	_jsii_.InvokeVoid(
		c,
		"resetPurpose",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetRegion() {
	_jsii_.InvokeVoid(
		c,
		"resetRegion",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetReservedInternalRange() {
	_jsii_.InvokeVoid(
		c,
		"resetReservedInternalRange",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetRole() {
	_jsii_.InvokeVoid(
		c,
		"resetRole",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetSecondaryIpRange() {
	_jsii_.InvokeVoid(
		c,
		"resetSecondaryIpRange",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetStackType() {
	_jsii_.InvokeVoid(
		c,
		"resetStackType",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeSubnetwork) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeSubnetwork) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

