package role

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type RoleConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// The name of the role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#name Role#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// Role to switch to at login.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#assume_role Role#assume_role}
	AssumeRole *string `field:"optional" json:"assumeRole" yaml:"assumeRole"`
	// Determine whether a role bypasses every row-level security (RLS) policy.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#bypass_row_level_security Role#bypass_row_level_security}
	BypassRowLevelSecurity interface{} `field:"optional" json:"bypassRowLevelSecurity" yaml:"bypassRowLevelSecurity"`
	// How many concurrent connections can be made with this role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#connection_limit Role#connection_limit}
	ConnectionLimit *float64 `field:"optional" json:"connectionLimit" yaml:"connectionLimit"`
	// Define a role's ability to create databases.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#create_database Role#create_database}
	CreateDatabase interface{} `field:"optional" json:"createDatabase" yaml:"createDatabase"`
	// Determine whether this role will be permitted to create new roles.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#create_role Role#create_role}
	CreateRole interface{} `field:"optional" json:"createRole" yaml:"createRole"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#encrypted Role#encrypted}.
	Encrypted *string `field:"optional" json:"encrypted" yaml:"encrypted"`
	// Control whether the password is stored encrypted in the system catalogs.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#encrypted_password Role#encrypted_password}
	EncryptedPassword interface{} `field:"optional" json:"encryptedPassword" yaml:"encryptedPassword"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#id Role#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Terminate any session with an open transaction that has been idle for longer than the specified duration in milliseconds.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#idle_in_transaction_session_timeout Role#idle_in_transaction_session_timeout}
	IdleInTransactionSessionTimeout *float64 `field:"optional" json:"idleInTransactionSessionTimeout" yaml:"idleInTransactionSessionTimeout"`
	// Determine whether a role "inherits" the privileges of roles it is a member of.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#inherit Role#inherit}
	Inherit interface{} `field:"optional" json:"inherit" yaml:"inherit"`
	// Determine whether a role is allowed to log in.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#login Role#login}
	Login interface{} `field:"optional" json:"login" yaml:"login"`
	// Sets the role's password.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#password Role#password}
	Password *string `field:"optional" json:"password" yaml:"password"`
	// Determine whether a role is allowed to initiate streaming replication or put the system in and out of backup mode.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#replication Role#replication}
	Replication interface{} `field:"optional" json:"replication" yaml:"replication"`
	// Role(s) to grant to this new role.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#roles Role#roles}
	Roles *[]*string `field:"optional" json:"roles" yaml:"roles"`
	// Sets the role's search path.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#search_path Role#search_path}
	SearchPath *[]*string `field:"optional" json:"searchPath" yaml:"searchPath"`
	// Skip actually running the DROP ROLE command when removing a ROLE from PostgreSQL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#skip_drop_role Role#skip_drop_role}
	SkipDropRole interface{} `field:"optional" json:"skipDropRole" yaml:"skipDropRole"`
	// Skip actually running the REASSIGN OWNED command when removing a role from PostgreSQL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#skip_reassign_owned Role#skip_reassign_owned}
	SkipReassignOwned interface{} `field:"optional" json:"skipReassignOwned" yaml:"skipReassignOwned"`
	// Abort any statement that takes more than the specified number of milliseconds.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#statement_timeout Role#statement_timeout}
	StatementTimeout *float64 `field:"optional" json:"statementTimeout" yaml:"statementTimeout"`
	// Determine whether the new role is a "superuser".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#superuser Role#superuser}
	Superuser interface{} `field:"optional" json:"superuser" yaml:"superuser"`
	// Sets a date and time after which the role's password is no longer valid.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs/resources/role#valid_until Role#valid_until}
	ValidUntil *string `field:"optional" json:"validUntil" yaml:"validUntil"`
}

