package iam

import (
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/projects_iam"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Output struct{}

type Variables struct {
	Project project.Project
}

const StackName = "iam"

func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
	stack := stacks.New(StackName)

	_ = projects_iam.NewProjectsIam(stack, pointers.Ptr("iam_binding"), &projects_iam.ProjectsIamConfig{
		Mode:     pointers.Ptr("authoritative"),
		Projects: &[]*string{vars.Project.Id()},
		Bindings: &map[string]*[]*string{
			// TODO: Is this static?
			"organizations/1006954638239/roles/EntitlePermissions":
			// TODO: What do these values mean? Are they sensitive?
			// https://sourcegraph.sourcegraph.com/github.com/sourcegraph/infrastructure/-/blob/cody-gateway/envs/prod/iam/main.tf?L7:1-10:6
			{},
		},
	})

	return &Output{}, nil
}
