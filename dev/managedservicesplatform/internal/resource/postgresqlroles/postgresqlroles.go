package postgresqlroles

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/grant"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/grantrole"
	postgresql "github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/provider"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/role"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/cloudsql"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/postgresqllogicalreplication"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	// WorkloadSuperuserGrant should be referenced as a dependency before
	// WorkloadUser is used.
	WorkloadSuperuserGrant cdktf.ITerraformDependable
}

type Config struct {
	Databases    []string
	CloudSQL     *cloudsql.Output
	Publications []postgresqllogicalreplication.PublicationOutput
}

// New applies PostgreSQL roles to a Cloud SQL database.
//
// When tearing down a database only (i.e. not destroying the entire environment),
// we must manually remove resources managed by this provider from state in order
// to apply the diff:
//
//	terraform state list | grep postgresql_grant | xargs terraform state rm
//
// This is because we cannot instantiate the provider when removing the
// database, causing plans and applies to fail. We'll likely be stuck with the
// workaround for a while, which is acceptable because CloudSQL-only teardowns
// should be rare - we'll more likely be removing entire environments in general.
//
// TODO(@bobheadxi): Improve documentation around this teardown scenario.
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	pgProvider := postgresql.NewPostgresqlProvider(scope, id.TerraformID("postgresql_provider"), &postgresql.PostgresqlProviderConfig{
		Scheme:    pointers.Ptr("gcppostgres"),
		Host:      config.CloudSQL.Instance.ConnectionName(),
		Username:  config.CloudSQL.AdminUser.Name(),
		Password:  config.CloudSQL.AdminUser.Password(),
		Port:      jsii.Number(5432),
		Superuser: jsii.Bool(false),
	})

	workloadSuperuserGrant := grantrole.NewGrantRole(scope, id.TerraformID("workload_service_account_role_cloudsqlsuperuser"), &grantrole.GrantRoleConfig{
		Provider:        pgProvider,
		Role:            config.CloudSQL.WorkloadUser.Name(),
		GrantRole:       jsii.String("cloudsqlsuperuser"),
		WithAdminOption: jsii.Bool(true),
	})

	// Operator access: grant restricted read-only permissions, based on
	// https://github.com/sourcegraph/deploy-sourcegraph-managed/blob/ded74a806bb6d1925cb894a8755ed52db7585a4f/modules/terraform-managed-instance-new/sql.tf#L153-L179
	for _, db := range config.Databases {
		_ = grant.NewGrant(scope, id.Group(db).TerraformID("operator_access_service_account_connect_grant"), &grant.GrantConfig{
			Provider:   pgProvider,
			Database:   &db,
			Role:       config.CloudSQL.OperatorAccessUser.Name(),
			ObjectType: pointers.Ptr("database"),
			Privileges: pointers.Ptr(pointers.Slice([]string{
				"CONNECT",
			})),
		})
		_ = grant.NewGrant(scope, id.Group(db).TerraformID("operator_access_service_account_table_grant"), &grant.GrantConfig{
			Provider: pgProvider,
			Database: &db,
			Role:     config.CloudSQL.OperatorAccessUser.Name(),
			Schema:   pointers.Ptr("public"),
			// All tables - objects is explicit empty array to indicate all tables
			ObjectType: pointers.Ptr("table"),
			Objects:    pointers.Ptr(pointers.Slice([]string{})),
			// Restricted privileges only
			Privileges: pointers.Ptr(pointers.Slice([]string{
				"SELECT",
			})),
		})
	}

	if len(config.Publications) > 0 {
		// Assign publication users permissions as required for GCP Datastream.
		// https://cloud.google.com/datastream/docs/configure-cloudsql-psql#cloudsqlforpostgres-create-datastream-user
		id := id.Group("publication")

		replicationRole := role.NewRole(scope, id.TerraformID("replicationrole"), &role.RoleConfig{
			Provider:    pgProvider,
			Name:        pointers.Ptr("replication_role"),
			Replication: pointers.Ptr(true),
		})

		for _, p := range config.Publications {
			id := id.Group(p.Name)

			// 	CREATE USER USER_NAME WITH REPLICATION LOGIN <...>;
			_ = grantrole.NewGrantRole(scope, id.TerraformID("user_replicationrole_grant"), &grantrole.GrantRoleConfig{
				Provider:  pgProvider,
				Role:      p.User.Name(),
				GrantRole: replicationRole.Name(),
			})

			// 	GRANT SELECT ON ALL TABLES IN SCHEMA SCHEMA_NAME TO USER_NAME;
			_ = grant.NewGrant(scope, id.TerraformID("user_table_select_grant"), &grant.GrantConfig{
				Provider:   pgProvider,
				Database:   &p.Database,
				Role:       p.User.Name(),
				Schema:     pointers.Ptr("public"),
				ObjectType: pointers.Ptr("table"),
				Objects:    pointers.Ptr(pointers.Slice(p.Tables)),
				// Restricted privileges only
				Privileges: pointers.Ptr(pointers.Slice([]string{
					"SELECT",
				})),
			})
			// 	GRANT USAGE ON SCHEMA SCHEMA_NAME TO USER_NAME;
			_ = grant.NewGrant(scope, id.TerraformID("user_schema_usage_grant"), &grant.GrantConfig{
				Provider:   pgProvider,
				Database:   &p.Database,
				Role:       p.User.Name(),
				ObjectType: pointers.Ptr("schema"),
				Schema:     pointers.Ptr("public"),
				Privileges: pointers.Ptr(pointers.Slice([]string{
					"USAGE",
				})),
			})
			// Unnecessary?
			//
			// 	ALTER DEFAULT PRIVILEGES IN SCHEMA SCHEMA_NAME
			// 		GRANT SELECT ON TABLES TO USER_NAME;
		}
	}

	return &Output{
		WorkloadSuperuserGrant: workloadSuperuserGrant,
	}, nil
}
