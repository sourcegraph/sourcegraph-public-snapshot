package postgresqlroles

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/grantrole"
	postgresql "github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/provider"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/cloudsql"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct {
	// WorkloadSuperuserGrant should be referenced as a dependency before
	// WorkloadUser is used.
	WorkloadSuperuserGrant cdktf.ITerraformDependable
}

type Config *cloudsql.Output

// New applies PostgreSQL roles to a Cloud SQL database.
//
// When tearing down a database only (i.e. not destroying the entire environment),
// we must manually remove resources managed by this provider from state in order
// to apply the diff:
//
//	terraform state list | grep postgresql_grant_role | xargs terraform state rm
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
		Host:      config.Instance.ConnectionName(),
		Username:  config.AdminUser.Name(),
		Password:  config.AdminUser.Password(),
		Port:      jsii.Number(5432),
		Superuser: jsii.Bool(false),
	})

	workloadSuperuserGrant := grantrole.NewGrantRole(scope, id.TerraformID("workload_service_account_role_cloudsqlsuperuser"), &grantrole.GrantRoleConfig{
		Provider:        pgProvider,
		Role:            config.WorkloadUser.Name(),
		GrantRole:       jsii.String("cloudsqlsuperuser"),
		WithAdminOption: jsii.Bool(true),
	})

	return &Output{
		WorkloadSuperuserGrant: workloadSuperuserGrant,
	}, nil
}
