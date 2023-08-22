package serviceaccount

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectiammember"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/serviceaccount"

	"github.com/sourcegraph/sourcegraph/internal/pointer"
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
func New(scope constructs.Construct, id string, config Config) *Output {
	serviceAccount := serviceaccount.NewServiceAccount(scope,
		pointer.Stringf("%s-sa", id),
		&serviceaccount.ServiceAccountConfig{
			Project: config.Project.ProjectId(),

			AccountId:   pointer.Value(config.AccountID),
			DisplayName: pointer.Value(config.DisplayName),
		})
	for _, role := range config.Roles {
		_ = projectiammember.NewProjectIamMember(scope,
			pointer.Stringf("%s_sa_%s", id, role.ID),
			&projectiammember.ProjectIamMemberConfig{
				Project: config.Project.ProjectId(),

				Role:   pointer.Value(role.Role),
				Member: serviceAccount.Member(),
			})
	}
	return &Output{
		Email: *serviceAccount.Email(),
	}
}
