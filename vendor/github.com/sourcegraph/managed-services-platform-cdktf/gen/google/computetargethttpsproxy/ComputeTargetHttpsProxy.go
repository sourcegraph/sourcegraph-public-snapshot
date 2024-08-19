package computetargethttpsproxy

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/google/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/computetargethttpsproxy/internal"
)

// Represents a {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy google_compute_target_https_proxy}.
type ComputeTargetHttpsProxy interface {
	cdktf.TerraformResource
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	CertificateManagerCertificates() *[]*string
	SetCertificateManagerCertificates(val *[]*string)
	CertificateManagerCertificatesInput() *[]*string
	CertificateMap() *string
	SetCertificateMap(val *string)
	CertificateMapInput() *string
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
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	HttpKeepAliveTimeoutSec() *float64
	SetHttpKeepAliveTimeoutSec(val *float64)
	HttpKeepAliveTimeoutSecInput() *float64
	Id() *string
	SetId(val *string)
	IdInput() *string
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
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
	ProxyBind() interface{}
	SetProxyBind(val interface{})
	ProxyBindInput() interface{}
	ProxyId() *float64
	QuicOverride() *string
	SetQuicOverride(val *string)
	QuicOverrideInput() *string
	// Experimental.
	RawOverrides() interface{}
	SelfLink() *string
	ServerTlsPolicy() *string
	SetServerTlsPolicy(val *string)
	ServerTlsPolicyInput() *string
	SslCertificates() *[]*string
	SetSslCertificates(val *[]*string)
	SslCertificatesInput() *[]*string
	SslPolicy() *string
	SetSslPolicy(val *string)
	SslPolicyInput() *string
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	Timeouts() ComputeTargetHttpsProxyTimeoutsOutputReference
	TimeoutsInput() interface{}
	UrlMap() *string
	SetUrlMap(val *string)
	UrlMapInput() *string
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
	PutTimeouts(value *ComputeTargetHttpsProxyTimeouts)
	ResetCertificateManagerCertificates()
	ResetCertificateMap()
	ResetDescription()
	ResetHttpKeepAliveTimeoutSec()
	ResetId()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetProject()
	ResetProxyBind()
	ResetQuicOverride()
	ResetServerTlsPolicy()
	ResetSslCertificates()
	ResetSslPolicy()
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

// The jsii proxy struct for ComputeTargetHttpsProxy
type jsiiProxy_ComputeTargetHttpsProxy struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) CertificateManagerCertificates() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"certificateManagerCertificates",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) CertificateManagerCertificatesInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"certificateManagerCertificatesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) CertificateMap() *string {
	var returns *string
	_jsii_.Get(
		j,
		"certificateMap",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) CertificateMapInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"certificateMapInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) CreationTimestamp() *string {
	var returns *string
	_jsii_.Get(
		j,
		"creationTimestamp",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Description() *string {
	var returns *string
	_jsii_.Get(
		j,
		"description",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) DescriptionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"descriptionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) HttpKeepAliveTimeoutSec() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"httpKeepAliveTimeoutSec",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) HttpKeepAliveTimeoutSecInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"httpKeepAliveTimeoutSecInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Project() *string {
	var returns *string
	_jsii_.Get(
		j,
		"project",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ProjectInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"projectInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ProxyBind() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"proxyBind",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ProxyBindInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"proxyBindInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ProxyId() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"proxyId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) QuicOverride() *string {
	var returns *string
	_jsii_.Get(
		j,
		"quicOverride",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) QuicOverrideInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"quicOverrideInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) SelfLink() *string {
	var returns *string
	_jsii_.Get(
		j,
		"selfLink",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ServerTlsPolicy() *string {
	var returns *string
	_jsii_.Get(
		j,
		"serverTlsPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) ServerTlsPolicyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"serverTlsPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) SslCertificates() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"sslCertificates",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) SslCertificatesInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"sslCertificatesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) SslPolicy() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslPolicy",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) SslPolicyInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslPolicyInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) Timeouts() ComputeTargetHttpsProxyTimeoutsOutputReference {
	var returns ComputeTargetHttpsProxyTimeoutsOutputReference
	_jsii_.Get(
		j,
		"timeouts",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) TimeoutsInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"timeoutsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) UrlMap() *string {
	var returns *string
	_jsii_.Get(
		j,
		"urlMap",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_ComputeTargetHttpsProxy) UrlMapInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"urlMapInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy google_compute_target_https_proxy} Resource.
func NewComputeTargetHttpsProxy(scope constructs.Construct, id *string, config *ComputeTargetHttpsProxyConfig) ComputeTargetHttpsProxy {
	_init_.Initialize()

	if err := validateNewComputeTargetHttpsProxyParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_ComputeTargetHttpsProxy{}

	_jsii_.Create(
		"@cdktf/provider-google.computeTargetHttpsProxy.ComputeTargetHttpsProxy",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/compute_target_https_proxy google_compute_target_https_proxy} Resource.
func NewComputeTargetHttpsProxy_Override(c ComputeTargetHttpsProxy, scope constructs.Construct, id *string, config *ComputeTargetHttpsProxyConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-google.computeTargetHttpsProxy.ComputeTargetHttpsProxy",
		[]interface{}{scope, id, config},
		c,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetCertificateManagerCertificates(val *[]*string) {
	if err := j.validateSetCertificateManagerCertificatesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"certificateManagerCertificates",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetCertificateMap(val *string) {
	if err := j.validateSetCertificateMapParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"certificateMap",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetDescription(val *string) {
	if err := j.validateSetDescriptionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"description",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetHttpKeepAliveTimeoutSec(val *float64) {
	if err := j.validateSetHttpKeepAliveTimeoutSecParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"httpKeepAliveTimeoutSec",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetProject(val *string) {
	if err := j.validateSetProjectParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"project",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetProxyBind(val interface{}) {
	if err := j.validateSetProxyBindParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"proxyBind",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetQuicOverride(val *string) {
	if err := j.validateSetQuicOverrideParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"quicOverride",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetServerTlsPolicy(val *string) {
	if err := j.validateSetServerTlsPolicyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"serverTlsPolicy",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetSslCertificates(val *[]*string) {
	if err := j.validateSetSslCertificatesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sslCertificates",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetSslPolicy(val *string) {
	if err := j.validateSetSslPolicyParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"sslPolicy",
		val,
	)
}

func (j *jsiiProxy_ComputeTargetHttpsProxy)SetUrlMap(val *string) {
	if err := j.validateSetUrlMapParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"urlMap",
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
func ComputeTargetHttpsProxy_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeTargetHttpsProxy_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeTargetHttpsProxy.ComputeTargetHttpsProxy",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeTargetHttpsProxy_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeTargetHttpsProxy_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeTargetHttpsProxy.ComputeTargetHttpsProxy",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func ComputeTargetHttpsProxy_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateComputeTargetHttpsProxy_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-google.computeTargetHttpsProxy.ComputeTargetHttpsProxy",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func ComputeTargetHttpsProxy_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-google.computeTargetHttpsProxy.ComputeTargetHttpsProxy",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) AddOverride(path *string, value interface{}) {
	if err := c.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetListAttribute(terraformAttribute *string) *[]*string {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetNumberAttribute(terraformAttribute *string) *float64 {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetStringAttribute(terraformAttribute *string) *string {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
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

func (c *jsiiProxy_ComputeTargetHttpsProxy) OverrideLogicalId(newLogicalId *string) {
	if err := c.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) PutTimeouts(value *ComputeTargetHttpsProxyTimeouts) {
	if err := c.validatePutTimeoutsParameters(value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		c,
		"putTimeouts",
		[]interface{}{value},
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetCertificateManagerCertificates() {
	_jsii_.InvokeVoid(
		c,
		"resetCertificateManagerCertificates",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetCertificateMap() {
	_jsii_.InvokeVoid(
		c,
		"resetCertificateMap",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetDescription() {
	_jsii_.InvokeVoid(
		c,
		"resetDescription",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetHttpKeepAliveTimeoutSec() {
	_jsii_.InvokeVoid(
		c,
		"resetHttpKeepAliveTimeoutSec",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetId() {
	_jsii_.InvokeVoid(
		c,
		"resetId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		c,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetProject() {
	_jsii_.InvokeVoid(
		c,
		"resetProject",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetProxyBind() {
	_jsii_.InvokeVoid(
		c,
		"resetProxyBind",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetQuicOverride() {
	_jsii_.InvokeVoid(
		c,
		"resetQuicOverride",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetServerTlsPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetServerTlsPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetSslCertificates() {
	_jsii_.InvokeVoid(
		c,
		"resetSslCertificates",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetSslPolicy() {
	_jsii_.InvokeVoid(
		c,
		"resetSslPolicy",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ResetTimeouts() {
	_jsii_.InvokeVoid(
		c,
		"resetTimeouts",
		nil, // no parameters
	)
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		c,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		c,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (c *jsiiProxy_ComputeTargetHttpsProxy) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		c,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

