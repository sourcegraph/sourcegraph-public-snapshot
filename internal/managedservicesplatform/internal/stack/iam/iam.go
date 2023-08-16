package iam

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/google/project"
	"github.com/sourcegraph/managed-services-platform-cdktf/gen/projects_iam"

	"github.com/sourcegraph/sourcegraph/internal/managedservicesplatform/internal/stack"
)

type Output struct{}

type Variables struct {
	Project project.Project
}

const StackName = "iam"

func NewStack(stacks *stack.Set, vars Variables) (*Output, error) {
	stack := stacks.New(StackName)

	_ = projects_iam.NewProjectsIam(stack, jsii.String("iam_binding"), &projects_iam.ProjectsIamConfig{
		Mode:     jsii.String("authoritative"),
		Projects: &[]*string{vars.Project.Id()},
		Bindings: &map[string]*[]*string{
			*jsii.String("organizations/1006954638239/roles/EntitlePermissions"):
			// TODO: What do these values mean? https://sourcegraph.sourcegraph.com/github.com/sourcegraph/infrastructure/-/blob/cody-gateway/envs/prod/iam/main.tf?L7:1-10:6
			jsii.Strings(""),
		},
	})

	return &Output{}, nil
}
