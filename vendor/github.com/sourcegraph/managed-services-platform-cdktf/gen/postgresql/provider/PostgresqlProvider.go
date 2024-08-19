package provider

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/provider/internal"
)

// Represents a {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs postgresql}.
type PostgresqlProvider interface {
	cdktf.TerraformProvider
	Alias() *string
	SetAlias(val *string)
	AliasInput() *string
	AwsRdsIamAuth() interface{}
	SetAwsRdsIamAuth(val interface{})
	AwsRdsIamAuthInput() interface{}
	AwsRdsIamProfile() *string
	SetAwsRdsIamProfile(val *string)
	AwsRdsIamProfileInput() *string
	AwsRdsIamRegion() *string
	SetAwsRdsIamRegion(val *string)
	AwsRdsIamRegionInput() *string
	AzureIdentityAuth() interface{}
	SetAzureIdentityAuth(val interface{})
	AzureIdentityAuthInput() interface{}
	AzureTenantId() *string
	SetAzureTenantId(val *string)
	AzureTenantIdInput() *string
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	Clientcert() *PostgresqlProviderClientcert
	SetClientcert(val *PostgresqlProviderClientcert)
	ClientcertInput() *PostgresqlProviderClientcert
	ConnectTimeout() *float64
	SetConnectTimeout(val *float64)
	ConnectTimeoutInput() *float64
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	Database() *string
	SetDatabase(val *string)
	DatabaseInput() *string
	DatabaseUsername() *string
	SetDatabaseUsername(val *string)
	DatabaseUsernameInput() *string
	ExpectedVersion() *string
	SetExpectedVersion(val *string)
	ExpectedVersionInput() *string
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	GcpIamImpersonateServiceAccount() *string
	SetGcpIamImpersonateServiceAccount(val *string)
	GcpIamImpersonateServiceAccountInput() *string
	Host() *string
	SetHost(val *string)
	HostInput() *string
	MaxConnections() *float64
	SetMaxConnections(val *float64)
	MaxConnectionsInput() *float64
	// Experimental.
	MetaAttributes() *map[string]interface{}
	// The tree node.
	Node() constructs.Node
	Password() *string
	SetPassword(val *string)
	PasswordInput() *string
	Port() *float64
	SetPort(val *float64)
	PortInput() *float64
	// Experimental.
	RawOverrides() interface{}
	Scheme() *string
	SetScheme(val *string)
	SchemeInput() *string
	Sslmode() *string
	SetSslmode(val *string)
	SslMode() *string
	SetSslMode(val *string)
	SslmodeInput() *string
	SslModeInput() *string
	Sslrootcert() *string
	SetSslrootcert(val *string)
	SslrootcertInput() *string
	Superuser() interface{}
	SetSuperuser(val interface{})
	SuperuserInput() interface{}
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformProviderSource() *string
	// Experimental.
	TerraformResourceType() *string
	Username() *string
	SetUsername(val *string)
	UsernameInput() *string
	// Experimental.
	AddOverride(path *string, value interface{})
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	ResetAlias()
	ResetAwsRdsIamAuth()
	ResetAwsRdsIamProfile()
	ResetAwsRdsIamRegion()
	ResetAzureIdentityAuth()
	ResetAzureTenantId()
	ResetClientcert()
	ResetConnectTimeout()
	ResetDatabase()
	ResetDatabaseUsername()
	ResetExpectedVersion()
	ResetGcpIamImpersonateServiceAccount()
	ResetHost()
	ResetMaxConnections()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPassword()
	ResetPort()
	ResetScheme()
	ResetSslmode()
	ResetSslMode()
	ResetSslrootcert()
	ResetSuperuser()
	ResetUsername()
	SynthesizeAttributes() *map[string]interface{}
	SynthesizeHclAttributes() *map[string]interface{}
	// Experimental.
	ToHclTerraform() interface{}
	// Experimental.
	ToMetadata() interface{}
	// Returns a string representation of this construct.
	ToString() *string
	// Adds this resource to the terraform JSON output.
	// Experimental.
	ToTerraform() interface{}
}

// The jsii proxy struct for PostgresqlProvider
type jsiiProxy_PostgresqlProvider struct {
	internal.Type__cdktfTerraformProvider
}

func (j *jsiiProxy_PostgresqlProvider) Alias() *string {
	var returns *string
	_jsii_.Get(
		j,
		"alias",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AliasInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"aliasInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AwsRdsIamAuth() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"awsRdsIamAuth",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AwsRdsIamAuthInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"awsRdsIamAuthInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AwsRdsIamProfile() *string {
	var returns *string
	_jsii_.Get(
		j,
		"awsRdsIamProfile",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AwsRdsIamProfileInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"awsRdsIamProfileInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AwsRdsIamRegion() *string {
	var returns *string
	_jsii_.Get(
		j,
		"awsRdsIamRegion",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AwsRdsIamRegionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"awsRdsIamRegionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AzureIdentityAuth() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"azureIdentityAuth",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AzureIdentityAuthInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"azureIdentityAuthInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AzureTenantId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"azureTenantId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) AzureTenantIdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"azureTenantIdInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Clientcert() *PostgresqlProviderClientcert {
	var returns *PostgresqlProviderClientcert
	_jsii_.Get(
		j,
		"clientcert",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) ClientcertInput() *PostgresqlProviderClientcert {
	var returns *PostgresqlProviderClientcert
	_jsii_.Get(
		j,
		"clientcertInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) ConnectTimeout() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectTimeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) ConnectTimeoutInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectTimeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Database() *string {
	var returns *string
	_jsii_.Get(
		j,
		"database",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) DatabaseInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"databaseInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) DatabaseUsername() *string {
	var returns *string
	_jsii_.Get(
		j,
		"databaseUsername",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) DatabaseUsernameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"databaseUsernameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) ExpectedVersion() *string {
	var returns *string
	_jsii_.Get(
		j,
		"expectedVersion",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) ExpectedVersionInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"expectedVersionInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) GcpIamImpersonateServiceAccount() *string {
	var returns *string
	_jsii_.Get(
		j,
		"gcpIamImpersonateServiceAccount",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) GcpIamImpersonateServiceAccountInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"gcpIamImpersonateServiceAccountInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Host() *string {
	var returns *string
	_jsii_.Get(
		j,
		"host",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) HostInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"hostInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) MaxConnections() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnections",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) MaxConnectionsInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"maxConnectionsInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) MetaAttributes() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"metaAttributes",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Password() *string {
	var returns *string
	_jsii_.Get(
		j,
		"password",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) PasswordInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"passwordInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Port() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"port",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) PortInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"portInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Scheme() *string {
	var returns *string
	_jsii_.Get(
		j,
		"scheme",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) SchemeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"schemeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Sslmode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslmode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) SslMode() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslMode",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) SslmodeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslmodeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) SslModeInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslModeInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Sslrootcert() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslrootcert",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) SslrootcertInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"sslrootcertInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Superuser() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"superuser",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) SuperuserInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"superuserInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) TerraformProviderSource() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformProviderSource",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) Username() *string {
	var returns *string
	_jsii_.Get(
		j,
		"username",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_PostgresqlProvider) UsernameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"usernameInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs postgresql} Resource.
func NewPostgresqlProvider(scope constructs.Construct, id *string, config *PostgresqlProviderConfig) PostgresqlProvider {
	_init_.Initialize()

	if err := validateNewPostgresqlProviderParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_PostgresqlProvider{}

	_jsii_.Create(
		"@cdktf/provider-postgresql.provider.PostgresqlProvider",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs postgresql} Resource.
func NewPostgresqlProvider_Override(p PostgresqlProvider, scope constructs.Construct, id *string, config *PostgresqlProviderConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-postgresql.provider.PostgresqlProvider",
		[]interface{}{scope, id, config},
		p,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetAlias(val *string) {
	_jsii_.Set(
		j,
		"alias",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetAwsRdsIamAuth(val interface{}) {
	if err := j.validateSetAwsRdsIamAuthParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"awsRdsIamAuth",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetAwsRdsIamProfile(val *string) {
	_jsii_.Set(
		j,
		"awsRdsIamProfile",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetAwsRdsIamRegion(val *string) {
	_jsii_.Set(
		j,
		"awsRdsIamRegion",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetAzureIdentityAuth(val interface{}) {
	if err := j.validateSetAzureIdentityAuthParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"azureIdentityAuth",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetAzureTenantId(val *string) {
	_jsii_.Set(
		j,
		"azureTenantId",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetClientcert(val *PostgresqlProviderClientcert) {
	if err := j.validateSetClientcertParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"clientcert",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetConnectTimeout(val *float64) {
	_jsii_.Set(
		j,
		"connectTimeout",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetDatabase(val *string) {
	_jsii_.Set(
		j,
		"database",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetDatabaseUsername(val *string) {
	_jsii_.Set(
		j,
		"databaseUsername",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetExpectedVersion(val *string) {
	_jsii_.Set(
		j,
		"expectedVersion",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetGcpIamImpersonateServiceAccount(val *string) {
	_jsii_.Set(
		j,
		"gcpIamImpersonateServiceAccount",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetHost(val *string) {
	_jsii_.Set(
		j,
		"host",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetMaxConnections(val *float64) {
	_jsii_.Set(
		j,
		"maxConnections",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetPassword(val *string) {
	_jsii_.Set(
		j,
		"password",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetPort(val *float64) {
	_jsii_.Set(
		j,
		"port",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetScheme(val *string) {
	_jsii_.Set(
		j,
		"scheme",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetSslmode(val *string) {
	_jsii_.Set(
		j,
		"sslmode",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetSslMode(val *string) {
	_jsii_.Set(
		j,
		"sslMode",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetSslrootcert(val *string) {
	_jsii_.Set(
		j,
		"sslrootcert",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetSuperuser(val interface{}) {
	if err := j.validateSetSuperuserParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"superuser",
		val,
	)
}

func (j *jsiiProxy_PostgresqlProvider)SetUsername(val *string) {
	_jsii_.Set(
		j,
		"username",
		val,
	)
}

// Generates CDKTF code for importing a PostgresqlProvider resource upon running "cdktf plan <stack-name>".
func PostgresqlProvider_GenerateConfigForImport(scope constructs.Construct, importToId *string, importFromId *string, provider cdktf.TerraformProvider) cdktf.ImportableResource {
	_init_.Initialize()

	if err := validatePostgresqlProvider_GenerateConfigForImportParameters(scope, importToId, importFromId); err != nil {
		panic(err)
	}
	var returns cdktf.ImportableResource

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.provider.PostgresqlProvider",
		"generateConfigForImport",
		[]interface{}{scope, importToId, importFromId, provider},
		&returns,
	)

	return returns
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
func PostgresqlProvider_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePostgresqlProvider_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.provider.PostgresqlProvider",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func PostgresqlProvider_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePostgresqlProvider_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.provider.PostgresqlProvider",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func PostgresqlProvider_IsTerraformProvider(x interface{}) *bool {
	_init_.Initialize()

	if err := validatePostgresqlProvider_IsTerraformProviderParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.provider.PostgresqlProvider",
		"isTerraformProvider",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func PostgresqlProvider_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-postgresql.provider.PostgresqlProvider",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (p *jsiiProxy_PostgresqlProvider) AddOverride(path *string, value interface{}) {
	if err := p.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (p *jsiiProxy_PostgresqlProvider) OverrideLogicalId(newLogicalId *string) {
	if err := p.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		p,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetAlias() {
	_jsii_.InvokeVoid(
		p,
		"resetAlias",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetAwsRdsIamAuth() {
	_jsii_.InvokeVoid(
		p,
		"resetAwsRdsIamAuth",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetAwsRdsIamProfile() {
	_jsii_.InvokeVoid(
		p,
		"resetAwsRdsIamProfile",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetAwsRdsIamRegion() {
	_jsii_.InvokeVoid(
		p,
		"resetAwsRdsIamRegion",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetAzureIdentityAuth() {
	_jsii_.InvokeVoid(
		p,
		"resetAzureIdentityAuth",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetAzureTenantId() {
	_jsii_.InvokeVoid(
		p,
		"resetAzureTenantId",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetClientcert() {
	_jsii_.InvokeVoid(
		p,
		"resetClientcert",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetConnectTimeout() {
	_jsii_.InvokeVoid(
		p,
		"resetConnectTimeout",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetDatabase() {
	_jsii_.InvokeVoid(
		p,
		"resetDatabase",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetDatabaseUsername() {
	_jsii_.InvokeVoid(
		p,
		"resetDatabaseUsername",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetExpectedVersion() {
	_jsii_.InvokeVoid(
		p,
		"resetExpectedVersion",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetGcpIamImpersonateServiceAccount() {
	_jsii_.InvokeVoid(
		p,
		"resetGcpIamImpersonateServiceAccount",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetHost() {
	_jsii_.InvokeVoid(
		p,
		"resetHost",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetMaxConnections() {
	_jsii_.InvokeVoid(
		p,
		"resetMaxConnections",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		p,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetPassword() {
	_jsii_.InvokeVoid(
		p,
		"resetPassword",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetPort() {
	_jsii_.InvokeVoid(
		p,
		"resetPort",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetScheme() {
	_jsii_.InvokeVoid(
		p,
		"resetScheme",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetSslmode() {
	_jsii_.InvokeVoid(
		p,
		"resetSslmode",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetSslMode() {
	_jsii_.InvokeVoid(
		p,
		"resetSslMode",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetSslrootcert() {
	_jsii_.InvokeVoid(
		p,
		"resetSslrootcert",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetSuperuser() {
	_jsii_.InvokeVoid(
		p,
		"resetSuperuser",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) ResetUsername() {
	_jsii_.InvokeVoid(
		p,
		"resetUsername",
		nil, // no parameters
	)
}

func (p *jsiiProxy_PostgresqlProvider) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PostgresqlProvider) SynthesizeHclAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		p,
		"synthesizeHclAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PostgresqlProvider) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		p,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PostgresqlProvider) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		p,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PostgresqlProvider) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		p,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (p *jsiiProxy_PostgresqlProvider) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		p,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

