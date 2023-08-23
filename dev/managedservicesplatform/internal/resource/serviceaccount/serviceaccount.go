package serviceaccount

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectiammember"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/serviceaccount"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Role struct {
	// ID is used to generate a resource ID for the role grant. Must be provided
	ID string
	// Role is sourced from https://cloud.google.com/iam/docs/understanding-roles#predefined
	Role string
}

type Config struct {
	Project project.Project

	AccountID   string
	DisplayName string
	Roles       []Role
}

type Output struct {
	Email string
}

// New provisions a service account, including roles for it to inherit.
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	serviceAccount := serviceaccount.NewServiceAccount(scope,
		id.ResourceID("account"),
		&serviceaccount.ServiceAccountConfig{
			Project: config.Project.ProjectId(),

			AccountId:   pointers.Ptr(config.AccountID),
			DisplayName: pointers.Ptr(config.DisplayName),
		})
	for _, role := range config.Roles {
		_ = projectiammember.NewProjectIamMember(scope,
			id.ResourceID("member_%s", role.ID),
			&projectiammember.ProjectIamMemberConfig{
				Project: config.Project.ProjectId(),

				Role:   pointers.Ptr(role.Role),
				Member: serviceAccount.Member(),
			})
	}
	return &Output{
		Email: *serviceAccount.Email(),
	}
}
