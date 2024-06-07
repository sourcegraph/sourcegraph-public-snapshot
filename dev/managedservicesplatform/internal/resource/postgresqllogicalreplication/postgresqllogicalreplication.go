package postgresqllogicalreplication

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/sqluser"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/publication"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/postgresql/replicationslot"
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
	User sqluser.SqlUser
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
	var publicationOutputs []PublicationOutput
	for _, p := range config.Spec.Publications {
		id := id.Group("publications").Group(p.Name)

		// Create user for Datastream:
		// https://cloud.google.com/datastream/docs/configure-cloudsql-psql#cloudsqlforpostgres-create-datastream-user
		logicalReplicationUserPassword := password.NewPassword(scope, id.TerraformID("user_password"), &password.PasswordConfig{
			Length:  pointers.Float64(32),
			Special: pointers.Ptr(false),
		})
		logicalReplicationUser := sqluser.NewSqlUser(scope, id.TerraformID("user"), &sqluser.SqlUserConfig{
			Instance: config.CloudSQL.Instance.Name(),
			Project:  config.CloudSQL.Instance.Project(),
			Name:     pointers.Stringf("publication_%s_user", p.Name),
			Password: logicalReplicationUserPassword.Result(),

			// PostgreSQL cannot delete users with roles, so we just abandon the
			// users in a deletion event.
			DeletionPolicy: pointers.Ptr("ABANDON"),
		})

		// Provision publication and replication slot:
		// https://cloud.google.com/datastream/docs/configure-cloudsql-psql#cloudsqlforpostgres-create-publication-and-replication-slot
		publicationOutputs = append(publicationOutputs, PublicationOutput{
			EnvironmentResourcePostgreSQLLogicalReplicationPublicationsSpec: p,
			PublicationName: publication.NewPublication(scope, id.TerraformID("publication"), &publication.PublicationConfig{
				Provider: config.PostgreSQLProvider,
				Name:     pointers.Ptr(p.Name),
				Database: pointers.Ptr(p.Database),
				Tables:   pointers.Ptr(pointers.Slice(p.Tables)),
			}).Name(),
			ReplicationSlotName: replicationslot.NewReplicationSlot(scope, id.TerraformID("replication_slot"), &replicationslot.ReplicationSlotConfig{
				Provider: config.PostgreSQLProvider,
				Name:     pointers.Ptr(p.Name + "_pgoutput"),
				Database: pointers.Ptr(p.Database),
				Plugin:   pointers.Ptr("pgoutput"),
			}).Name(),
			User: logicalReplicationUser,
		})
	}

	return &Output{
		Publications: publicationOutputs,
	}, nil
}
