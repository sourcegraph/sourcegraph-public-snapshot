package postgresqllogicalreplication

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	postgresql "github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/provider"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/publication"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/replicationslot"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/role"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/random/password"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/cloudsql"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	PostgreSQLProvider cdktf.TerraformProvider

	CloudSQL *cloudsql.Output
	Spec     spec.EnvironmentResourcePostgreSQLLogicalReplicationSpec
}

type PublicationOutput struct {
	// The name of your publication. You'll need to provide this name when you
	// create a stream in the Datastream stream creation wizard.
	PublicationName *string
	// The name of your replication slot. You'll need to provide this name when
	// you create a stream in the Datastream stream creation wizard.
	ReplicationSlotName *string
	// User for subscribing to the publication.
	User role.Role
	// The original publication spec.
	spec.EnvironmentResourcePostgreSQLLogicalReplicationPublicationsSpec
}

type Output struct {
	Publications []PublicationOutput
}

// New applies PostgreSQL runtime configuration for PostgreSQL logical replication.
//
// When tearing down a database only (i.e. not destroying the entire environment),
// we must manually remove resources managed by this provider from state in order
// to apply the diff:
//
//	terraform state list | grep postgresql_publication | xargs terraform state rm
//	terraform state list | grep postgresql_replication_slot | xargs terraform state rm
//
// This is because we cannot instantiate the provider when removing the
// database, causing plans and applies to fail. We'll likely be stuck with the
// workaround for a while, which is acceptable because CloudSQL-only teardowns
// should be rare - we'll more likely be removing entire environments in general.
//
// TODO(@bobheadxi): Improve documentation around this teardown scenario.
func New(scope constructs.Construct, id resourceid.ID, config Config) (*Output, error) {
	// TODO explain
	replicationSlotCreator := role.NewRole(scope, id.TerraformID("replicationslotcreator"), &role.RoleConfig{
		Provider: config.PostgreSQLProvider,
		Name:     pointers.Ptr("msp-replicationslotcreator"),
		Password: password.NewPassword(scope, id.TerraformID("replicationslotcreator_password"), &password.PasswordConfig{
			Length:  pointers.Float64(32),
			Special: pointers.Ptr(false),
		}).Result(),
		Replication: pointers.Ptr(true),
	})
	replicationSlotProvider := postgresql.NewPostgresqlProvider(scope,
		id.TerraformID("postgresql_replicationslotcreator_provider"),
		&postgresql.PostgresqlProviderConfig{
			Scheme:    pointers.Ptr("gcppostgres"),
			Host:      config.CloudSQL.Instance.ConnectionName(),
			Username:  replicationSlotCreator.Name(),
			Password:  replicationSlotCreator.Password(),
			Port:      jsii.Number(5432),
			Superuser: jsii.Bool(false),
			Alias:     pointers.Ptr("postgresql_replicationslotcreator_provider"),
		})

	var publicationOutputs []PublicationOutput
	for _, p := range config.Spec.Publications {
		id := id.Group("publications").Group(p.Name)

		// Create user for Datastream:
		// https://cloud.google.com/datastream/docs/configure-cloudsql-psql#cloudsqlforpostgres-create-datastream-user
		logicalReplicationUser := role.NewRole(scope, id.TerraformID("user"), &role.RoleConfig{
			Provider: config.PostgreSQLProvider,
			Name:     pointers.Stringf("publication_%s_user", p.Name),
			Password: password.NewPassword(scope, id.TerraformID("user_password"), &password.PasswordConfig{
				Length:  pointers.Float64(32),
				Special: pointers.Ptr(false),
			}).Result(),
			Replication: pointers.Ptr(true),
		})

		// Provision publication and replication slot:
		// https://cloud.google.com/datastream/docs/configure-cloudsql-psql#cloudsqlforpostgres-create-publication-and-replication-slot
		publicationOutputs = append(publicationOutputs, PublicationOutput{
			EnvironmentResourcePostgreSQLLogicalReplicationPublicationsSpec: p,
			PublicationName: publication.NewPublication(scope,
				id.TerraformID("publication"),
				&publication.PublicationConfig{
					// Tables are created (and therefore owned) by the application
					// workload user by default, so we use the provider authenticated
					// as the workload user.
					Provider: config.PostgreSQLProvider,
					Name:     pointers.Ptr(p.Name),
					Database: pointers.Ptr(p.Database),
					Tables:   pointers.Ptr(pointers.Slice(p.Tables)),
				}).Name(),
			ReplicationSlotName: replicationslot.NewReplicationSlot(scope,
				id.TerraformID("replication_slot"),
				&replicationslot.ReplicationSlotConfig{
					Provider:  replicationSlotProvider,
					Name:      pointers.Ptr(p.Name + "_pgoutput"),
					Database:  pointers.Ptr(p.Database),
					Plugin:    pointers.Ptr("pgoutput"),
					DependsOn: &[]cdktf.ITerraformDependable{replicationSlotCreator},
				}).Name(),
			User: logicalReplicationUser,
		})
	}

	return &Output{
		Publications: publicationOutputs,
	}, nil
}
