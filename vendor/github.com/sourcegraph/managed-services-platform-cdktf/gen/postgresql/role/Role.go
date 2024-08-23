package role

import (
	_jsii_ "github.com/aws/jsii-runtime-go/runtime"
	_init_ "github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/jsii"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/role/internal"
)

// Represents a {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role postgresql_role}.
type Role interface {
	cdktf.TerraformResource
	AssumeRole() *string
	SetAssumeRole(val *string)
	AssumeRoleInput() *string
	BypassRowLevelSecurity() interface{}
	SetBypassRowLevelSecurity(val interface{})
	BypassRowLevelSecurityInput() interface{}
	// Experimental.
	CdktfStack() cdktf.TerraformStack
	// Experimental.
	Connection() interface{}
	// Experimental.
	SetConnection(val interface{})
	ConnectionLimit() *float64
	SetConnectionLimit(val *float64)
	ConnectionLimitInput() *float64
	// Experimental.
	ConstructNodeMetadata() *map[string]interface{}
	// Experimental.
	Count() interface{}
	// Experimental.
	SetCount(val interface{})
	CreateDatabase() interface{}
	SetCreateDatabase(val interface{})
	CreateDatabaseInput() interface{}
	CreateRole() interface{}
	SetCreateRole(val interface{})
	CreateRoleInput() interface{}
	// Experimental.
	DependsOn() *[]*string
	// Experimental.
	SetDependsOn(val *[]*string)
	Encrypted() *string
	SetEncrypted(val *string)
	EncryptedInput() *string
	EncryptedPassword() interface{}
	SetEncryptedPassword(val interface{})
	EncryptedPasswordInput() interface{}
	// Experimental.
	ForEach() cdktf.ITerraformIterator
	// Experimental.
	SetForEach(val cdktf.ITerraformIterator)
	// Experimental.
	Fqn() *string
	// Experimental.
	FriendlyUniqueId() *string
	Id() *string
	SetId(val *string)
	IdInput() *string
	IdleInTransactionSessionTimeout() *float64
	SetIdleInTransactionSessionTimeout(val *float64)
	IdleInTransactionSessionTimeoutInput() *float64
	Inherit() interface{}
	SetInherit(val interface{})
	InheritInput() interface{}
	// Experimental.
	Lifecycle() *cdktf.TerraformResourceLifecycle
	// Experimental.
	SetLifecycle(val *cdktf.TerraformResourceLifecycle)
	Login() interface{}
	SetLogin(val interface{})
	LoginInput() interface{}
	Name() *string
	SetName(val *string)
	NameInput() *string
	// The tree node.
	Node() constructs.Node
	Password() *string
	SetPassword(val *string)
	PasswordInput() *string
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
	Replication() interface{}
	SetReplication(val interface{})
	ReplicationInput() interface{}
	Roles() *[]*string
	SetRoles(val *[]*string)
	RolesInput() *[]*string
	SearchPath() *[]*string
	SetSearchPath(val *[]*string)
	SearchPathInput() *[]*string
	SkipDropRole() interface{}
	SetSkipDropRole(val interface{})
	SkipDropRoleInput() interface{}
	SkipReassignOwned() interface{}
	SetSkipReassignOwned(val interface{})
	SkipReassignOwnedInput() interface{}
	StatementTimeout() *float64
	SetStatementTimeout(val *float64)
	StatementTimeoutInput() *float64
	Superuser() interface{}
	SetSuperuser(val interface{})
	SuperuserInput() interface{}
	// Experimental.
	TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata
	// Experimental.
	TerraformMetaArguments() *map[string]interface{}
	// Experimental.
	TerraformResourceType() *string
	ValidUntil() *string
	SetValidUntil(val *string)
	ValidUntilInput() *string
	// Adds a user defined moveTarget string to this resource to be later used in .moveTo(moveTarget) to resolve the location of the move.
	// Experimental.
	AddMoveTarget(moveTarget *string)
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
	HasResourceMove() interface{}
	// Experimental.
	ImportFrom(id *string, provider cdktf.TerraformProvider)
	// Experimental.
	InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable
	// Move the resource corresponding to "id" to this resource.
	//
	// Note that the resource being moved from must be marked as moved using it's instance function.
	// Experimental.
	MoveFromId(id *string)
	// Moves this resource to the target resource given by moveTarget.
	// Experimental.
	MoveTo(moveTarget *string, index interface{})
	// Moves this resource to the resource corresponding to "id".
	// Experimental.
	MoveToId(id *string)
	// Overrides the auto-generated logical ID with a specific ID.
	// Experimental.
	OverrideLogicalId(newLogicalId *string)
	ResetAssumeRole()
	ResetBypassRowLevelSecurity()
	ResetConnectionLimit()
	ResetCreateDatabase()
	ResetCreateRole()
	ResetEncrypted()
	ResetEncryptedPassword()
	ResetId()
	ResetIdleInTransactionSessionTimeout()
	ResetInherit()
	ResetLogin()
	// Resets a previously passed logical Id to use the auto-generated logical id again.
	// Experimental.
	ResetOverrideLogicalId()
	ResetPassword()
	ResetReplication()
	ResetRoles()
	ResetSearchPath()
	ResetSkipDropRole()
	ResetSkipReassignOwned()
	ResetStatementTimeout()
	ResetSuperuser()
	ResetValidUntil()
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

// The jsii proxy struct for Role
type jsiiProxy_Role struct {
	internal.Type__cdktfTerraformResource
}

func (j *jsiiProxy_Role) AssumeRole() *string {
	var returns *string
	_jsii_.Get(
		j,
		"assumeRole",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) AssumeRoleInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"assumeRoleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) BypassRowLevelSecurity() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"bypassRowLevelSecurity",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) BypassRowLevelSecurityInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"bypassRowLevelSecurityInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) CdktfStack() cdktf.TerraformStack {
	var returns cdktf.TerraformStack
	_jsii_.Get(
		j,
		"cdktfStack",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Connection() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"connection",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) ConnectionLimit() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectionLimit",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) ConnectionLimitInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"connectionLimitInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) ConstructNodeMetadata() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"constructNodeMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Count() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"count",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) CreateDatabase() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"createDatabase",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) CreateDatabaseInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"createDatabaseInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) CreateRole() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"createRole",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) CreateRoleInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"createRoleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) DependsOn() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"dependsOn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Encrypted() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encrypted",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) EncryptedInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"encryptedInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) EncryptedPassword() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"encryptedPassword",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) EncryptedPasswordInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"encryptedPasswordInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) ForEach() cdktf.ITerraformIterator {
	var returns cdktf.ITerraformIterator
	_jsii_.Get(
		j,
		"forEach",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Fqn() *string {
	var returns *string
	_jsii_.Get(
		j,
		"fqn",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) FriendlyUniqueId() *string {
	var returns *string
	_jsii_.Get(
		j,
		"friendlyUniqueId",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Id() *string {
	var returns *string
	_jsii_.Get(
		j,
		"id",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) IdInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"idInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) IdleInTransactionSessionTimeout() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"idleInTransactionSessionTimeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) IdleInTransactionSessionTimeoutInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"idleInTransactionSessionTimeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Inherit() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"inherit",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) InheritInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"inheritInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Lifecycle() *cdktf.TerraformResourceLifecycle {
	var returns *cdktf.TerraformResourceLifecycle
	_jsii_.Get(
		j,
		"lifecycle",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Login() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"login",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) LoginInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"loginInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Name() *string {
	var returns *string
	_jsii_.Get(
		j,
		"name",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) NameInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"nameInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Node() constructs.Node {
	var returns constructs.Node
	_jsii_.Get(
		j,
		"node",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Password() *string {
	var returns *string
	_jsii_.Get(
		j,
		"password",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) PasswordInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"passwordInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Provider() cdktf.TerraformProvider {
	var returns cdktf.TerraformProvider
	_jsii_.Get(
		j,
		"provider",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Provisioners() *[]interface{} {
	var returns *[]interface{}
	_jsii_.Get(
		j,
		"provisioners",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) RawOverrides() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"rawOverrides",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Replication() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"replication",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) ReplicationInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"replicationInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Roles() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"roles",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) RolesInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"rolesInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) SearchPath() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"searchPath",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) SearchPathInput() *[]*string {
	var returns *[]*string
	_jsii_.Get(
		j,
		"searchPathInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) SkipDropRole() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"skipDropRole",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) SkipDropRoleInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"skipDropRoleInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) SkipReassignOwned() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"skipReassignOwned",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) SkipReassignOwnedInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"skipReassignOwnedInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) StatementTimeout() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"statementTimeout",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) StatementTimeoutInput() *float64 {
	var returns *float64
	_jsii_.Get(
		j,
		"statementTimeoutInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) Superuser() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"superuser",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) SuperuserInput() interface{} {
	var returns interface{}
	_jsii_.Get(
		j,
		"superuserInput",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) TerraformGeneratorMetadata() *cdktf.TerraformProviderGeneratorMetadata {
	var returns *cdktf.TerraformProviderGeneratorMetadata
	_jsii_.Get(
		j,
		"terraformGeneratorMetadata",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) TerraformMetaArguments() *map[string]interface{} {
	var returns *map[string]interface{}
	_jsii_.Get(
		j,
		"terraformMetaArguments",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) TerraformResourceType() *string {
	var returns *string
	_jsii_.Get(
		j,
		"terraformResourceType",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) ValidUntil() *string {
	var returns *string
	_jsii_.Get(
		j,
		"validUntil",
		&returns,
	)
	return returns
}

func (j *jsiiProxy_Role) ValidUntilInput() *string {
	var returns *string
	_jsii_.Get(
		j,
		"validUntilInput",
		&returns,
	)
	return returns
}


// Create a new {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role postgresql_role} Resource.
func NewRole(scope constructs.Construct, id *string, config *RoleConfig) Role {
	_init_.Initialize()

	if err := validateNewRoleParameters(scope, id, config); err != nil {
		panic(err)
	}
	j := jsiiProxy_Role{}

	_jsii_.Create(
		"@cdktf/provider-postgresql.role.Role",
		[]interface{}{scope, id, config},
		&j,
	)

	return &j
}

// Create a new {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role postgresql_role} Resource.
func NewRole_Override(r Role, scope constructs.Construct, id *string, config *RoleConfig) {
	_init_.Initialize()

	_jsii_.Create(
		"@cdktf/provider-postgresql.role.Role",
		[]interface{}{scope, id, config},
		r,
	)
}

func (j *jsiiProxy_Role)SetAssumeRole(val *string) {
	if err := j.validateSetAssumeRoleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"assumeRole",
		val,
	)
}

func (j *jsiiProxy_Role)SetBypassRowLevelSecurity(val interface{}) {
	if err := j.validateSetBypassRowLevelSecurityParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"bypassRowLevelSecurity",
		val,
	)
}

func (j *jsiiProxy_Role)SetConnection(val interface{}) {
	if err := j.validateSetConnectionParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connection",
		val,
	)
}

func (j *jsiiProxy_Role)SetConnectionLimit(val *float64) {
	if err := j.validateSetConnectionLimitParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"connectionLimit",
		val,
	)
}

func (j *jsiiProxy_Role)SetCount(val interface{}) {
	if err := j.validateSetCountParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"count",
		val,
	)
}

func (j *jsiiProxy_Role)SetCreateDatabase(val interface{}) {
	if err := j.validateSetCreateDatabaseParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"createDatabase",
		val,
	)
}

func (j *jsiiProxy_Role)SetCreateRole(val interface{}) {
	if err := j.validateSetCreateRoleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"createRole",
		val,
	)
}

func (j *jsiiProxy_Role)SetDependsOn(val *[]*string) {
	_jsii_.Set(
		j,
		"dependsOn",
		val,
	)
}

func (j *jsiiProxy_Role)SetEncrypted(val *string) {
	if err := j.validateSetEncryptedParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"encrypted",
		val,
	)
}

func (j *jsiiProxy_Role)SetEncryptedPassword(val interface{}) {
	if err := j.validateSetEncryptedPasswordParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"encryptedPassword",
		val,
	)
}

func (j *jsiiProxy_Role)SetForEach(val cdktf.ITerraformIterator) {
	_jsii_.Set(
		j,
		"forEach",
		val,
	)
}

func (j *jsiiProxy_Role)SetId(val *string) {
	if err := j.validateSetIdParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"id",
		val,
	)
}

func (j *jsiiProxy_Role)SetIdleInTransactionSessionTimeout(val *float64) {
	if err := j.validateSetIdleInTransactionSessionTimeoutParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"idleInTransactionSessionTimeout",
		val,
	)
}

func (j *jsiiProxy_Role)SetInherit(val interface{}) {
	if err := j.validateSetInheritParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"inherit",
		val,
	)
}

func (j *jsiiProxy_Role)SetLifecycle(val *cdktf.TerraformResourceLifecycle) {
	if err := j.validateSetLifecycleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"lifecycle",
		val,
	)
}

func (j *jsiiProxy_Role)SetLogin(val interface{}) {
	if err := j.validateSetLoginParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"login",
		val,
	)
}

func (j *jsiiProxy_Role)SetName(val *string) {
	if err := j.validateSetNameParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"name",
		val,
	)
}

func (j *jsiiProxy_Role)SetPassword(val *string) {
	if err := j.validateSetPasswordParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"password",
		val,
	)
}

func (j *jsiiProxy_Role)SetProvider(val cdktf.TerraformProvider) {
	_jsii_.Set(
		j,
		"provider",
		val,
	)
}

func (j *jsiiProxy_Role)SetProvisioners(val *[]interface{}) {
	if err := j.validateSetProvisionersParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"provisioners",
		val,
	)
}

func (j *jsiiProxy_Role)SetReplication(val interface{}) {
	if err := j.validateSetReplicationParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"replication",
		val,
	)
}

func (j *jsiiProxy_Role)SetRoles(val *[]*string) {
	if err := j.validateSetRolesParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"roles",
		val,
	)
}

func (j *jsiiProxy_Role)SetSearchPath(val *[]*string) {
	if err := j.validateSetSearchPathParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"searchPath",
		val,
	)
}

func (j *jsiiProxy_Role)SetSkipDropRole(val interface{}) {
	if err := j.validateSetSkipDropRoleParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"skipDropRole",
		val,
	)
}

func (j *jsiiProxy_Role)SetSkipReassignOwned(val interface{}) {
	if err := j.validateSetSkipReassignOwnedParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"skipReassignOwned",
		val,
	)
}

func (j *jsiiProxy_Role)SetStatementTimeout(val *float64) {
	if err := j.validateSetStatementTimeoutParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"statementTimeout",
		val,
	)
}

func (j *jsiiProxy_Role)SetSuperuser(val interface{}) {
	if err := j.validateSetSuperuserParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"superuser",
		val,
	)
}

func (j *jsiiProxy_Role)SetValidUntil(val *string) {
	if err := j.validateSetValidUntilParameters(val); err != nil {
		panic(err)
	}
	_jsii_.Set(
		j,
		"validUntil",
		val,
	)
}

// Generates CDKTF code for importing a Role resource upon running "cdktf plan <stack-name>".
func Role_GenerateConfigForImport(scope constructs.Construct, importToId *string, importFromId *string, provider cdktf.TerraformProvider) cdktf.ImportableResource {
	_init_.Initialize()

	if err := validateRole_GenerateConfigForImportParameters(scope, importToId, importFromId); err != nil {
		panic(err)
	}
	var returns cdktf.ImportableResource

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.role.Role",
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
func Role_IsConstruct(x interface{}) *bool {
	_init_.Initialize()

	if err := validateRole_IsConstructParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.role.Role",
		"isConstruct",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Role_IsTerraformElement(x interface{}) *bool {
	_init_.Initialize()

	if err := validateRole_IsTerraformElementParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.role.Role",
		"isTerraformElement",
		[]interface{}{x},
		&returns,
	)

	return returns
}

// Experimental.
func Role_IsTerraformResource(x interface{}) *bool {
	_init_.Initialize()

	if err := validateRole_IsTerraformResourceParameters(x); err != nil {
		panic(err)
	}
	var returns *bool

	_jsii_.StaticInvoke(
		"@cdktf/provider-postgresql.role.Role",
		"isTerraformResource",
		[]interface{}{x},
		&returns,
	)

	return returns
}

func Role_TfResourceType() *string {
	_init_.Initialize()
	var returns *string
	_jsii_.StaticGet(
		"@cdktf/provider-postgresql.role.Role",
		"tfResourceType",
		&returns,
	)
	return returns
}

func (r *jsiiProxy_Role) AddMoveTarget(moveTarget *string) {
	if err := r.validateAddMoveTargetParameters(moveTarget); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"addMoveTarget",
		[]interface{}{moveTarget},
	)
}

func (r *jsiiProxy_Role) AddOverride(path *string, value interface{}) {
	if err := r.validateAddOverrideParameters(path, value); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"addOverride",
		[]interface{}{path, value},
	)
}

func (r *jsiiProxy_Role) GetAnyMapAttribute(terraformAttribute *string) *map[string]interface{} {
	if err := r.validateGetAnyMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]interface{}

	_jsii_.Invoke(
		r,
		"getAnyMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetBooleanAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := r.validateGetBooleanAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"getBooleanAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetBooleanMapAttribute(terraformAttribute *string) *map[string]*bool {
	if err := r.validateGetBooleanMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*bool

	_jsii_.Invoke(
		r,
		"getBooleanMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetListAttribute(terraformAttribute *string) *[]*string {
	if err := r.validateGetListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*string

	_jsii_.Invoke(
		r,
		"getListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetNumberAttribute(terraformAttribute *string) *float64 {
	if err := r.validateGetNumberAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *float64

	_jsii_.Invoke(
		r,
		"getNumberAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetNumberListAttribute(terraformAttribute *string) *[]*float64 {
	if err := r.validateGetNumberListAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *[]*float64

	_jsii_.Invoke(
		r,
		"getNumberListAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetNumberMapAttribute(terraformAttribute *string) *map[string]*float64 {
	if err := r.validateGetNumberMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*float64

	_jsii_.Invoke(
		r,
		"getNumberMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetStringAttribute(terraformAttribute *string) *string {
	if err := r.validateGetStringAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *string

	_jsii_.Invoke(
		r,
		"getStringAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) GetStringMapAttribute(terraformAttribute *string) *map[string]*string {
	if err := r.validateGetStringMapAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns *map[string]*string

	_jsii_.Invoke(
		r,
		"getStringMapAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) HasResourceMove() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		r,
		"hasResourceMove",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) ImportFrom(id *string, provider cdktf.TerraformProvider) {
	if err := r.validateImportFromParameters(id); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"importFrom",
		[]interface{}{id, provider},
	)
}

func (r *jsiiProxy_Role) InterpolationForAttribute(terraformAttribute *string) cdktf.IResolvable {
	if err := r.validateInterpolationForAttributeParameters(terraformAttribute); err != nil {
		panic(err)
	}
	var returns cdktf.IResolvable

	_jsii_.Invoke(
		r,
		"interpolationForAttribute",
		[]interface{}{terraformAttribute},
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) MoveFromId(id *string) {
	if err := r.validateMoveFromIdParameters(id); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"moveFromId",
		[]interface{}{id},
	)
}

func (r *jsiiProxy_Role) MoveTo(moveTarget *string, index interface{}) {
	if err := r.validateMoveToParameters(moveTarget, index); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"moveTo",
		[]interface{}{moveTarget, index},
	)
}

func (r *jsiiProxy_Role) MoveToId(id *string) {
	if err := r.validateMoveToIdParameters(id); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"moveToId",
		[]interface{}{id},
	)
}

func (r *jsiiProxy_Role) OverrideLogicalId(newLogicalId *string) {
	if err := r.validateOverrideLogicalIdParameters(newLogicalId); err != nil {
		panic(err)
	}
	_jsii_.InvokeVoid(
		r,
		"overrideLogicalId",
		[]interface{}{newLogicalId},
	)
}

func (r *jsiiProxy_Role) ResetAssumeRole() {
	_jsii_.InvokeVoid(
		r,
		"resetAssumeRole",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetBypassRowLevelSecurity() {
	_jsii_.InvokeVoid(
		r,
		"resetBypassRowLevelSecurity",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetConnectionLimit() {
	_jsii_.InvokeVoid(
		r,
		"resetConnectionLimit",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetCreateDatabase() {
	_jsii_.InvokeVoid(
		r,
		"resetCreateDatabase",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetCreateRole() {
	_jsii_.InvokeVoid(
		r,
		"resetCreateRole",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetEncrypted() {
	_jsii_.InvokeVoid(
		r,
		"resetEncrypted",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetEncryptedPassword() {
	_jsii_.InvokeVoid(
		r,
		"resetEncryptedPassword",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetId() {
	_jsii_.InvokeVoid(
		r,
		"resetId",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetIdleInTransactionSessionTimeout() {
	_jsii_.InvokeVoid(
		r,
		"resetIdleInTransactionSessionTimeout",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetInherit() {
	_jsii_.InvokeVoid(
		r,
		"resetInherit",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetLogin() {
	_jsii_.InvokeVoid(
		r,
		"resetLogin",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetOverrideLogicalId() {
	_jsii_.InvokeVoid(
		r,
		"resetOverrideLogicalId",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetPassword() {
	_jsii_.InvokeVoid(
		r,
		"resetPassword",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetReplication() {
	_jsii_.InvokeVoid(
		r,
		"resetReplication",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetRoles() {
	_jsii_.InvokeVoid(
		r,
		"resetRoles",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetSearchPath() {
	_jsii_.InvokeVoid(
		r,
		"resetSearchPath",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetSkipDropRole() {
	_jsii_.InvokeVoid(
		r,
		"resetSkipDropRole",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetSkipReassignOwned() {
	_jsii_.InvokeVoid(
		r,
		"resetSkipReassignOwned",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetStatementTimeout() {
	_jsii_.InvokeVoid(
		r,
		"resetStatementTimeout",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetSuperuser() {
	_jsii_.InvokeVoid(
		r,
		"resetSuperuser",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) ResetValidUntil() {
	_jsii_.InvokeVoid(
		r,
		"resetValidUntil",
		nil, // no parameters
	)
}

func (r *jsiiProxy_Role) SynthesizeAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		r,
		"synthesizeAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) SynthesizeHclAttributes() *map[string]interface{} {
	var returns *map[string]interface{}

	_jsii_.Invoke(
		r,
		"synthesizeHclAttributes",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) ToHclTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		r,
		"toHclTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) ToMetadata() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		r,
		"toMetadata",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) ToString() *string {
	var returns *string

	_jsii_.Invoke(
		r,
		"toString",
		nil, // no parameters
		&returns,
	)

	return returns
}

func (r *jsiiProxy_Role) ToTerraform() interface{} {
	var returns interface{}

	_jsii_.Invoke(
		r,
		"toTerraform",
		nil, // no parameters
		&returns,
	)

	return returns
}

