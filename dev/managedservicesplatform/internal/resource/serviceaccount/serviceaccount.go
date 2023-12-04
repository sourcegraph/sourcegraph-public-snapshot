package serviceaccount

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/projectiammember"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/serviceaccount"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Role struct {
	// ID is used to generate a resource ID for the role grant. Must be provided
	ID resourceid.ID
	// Role is sourced from https://cloud.google.com/iam/docs/understanding-roles#predefined
	Role string
}

type Config struct {
	ProjectID string

	AccountID   string
	DisplayName string
	Roles       []Role
}

type Output struct {
	Email  string
	Member string
}

// New provisions a service account, including roles for it to inherit.
func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	serviceAccount := serviceaccount.NewServiceAccount(scope,
		id.TerraformID("account"),
		&serviceaccount.ServiceAccountConfig{
			Project: pointers.Ptr(config.ProjectID),

			AccountId:   pointers.Ptr(config.AccountID),
			DisplayName: pointers.Ptr(config.DisplayName),
		})
	for _, role := range config.Roles {
		_ = projectiammember.NewProjectIamMember(scope,
			id.Append(role.ID).TerraformID("member"),
			&projectiammember.ProjectIamMemberConfig{
				Project: pointers.Ptr(config.ProjectID),

				Role:   pointers.Ptr(role.Role),
				Member: serviceAccount.Member(),
			})
	}
	return &Output{
		Email:  *serviceAccount.Email(),
		Member: *serviceAccount.Member(),
	}
}
