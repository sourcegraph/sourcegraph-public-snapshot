package terraformversion

import (
	"fmt"
	"reflect"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/stack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// With applies an aspect enforcing the given Terraform version on a new stack.
//
// CDKTF does not provide a native way to configure terraform version,
// so we use an aspect to enforce it.
// Learn more: https://developer.hashicorp.com/terraform/cdktf/concepts/aspects
func With(terraformVersion string) stack.NewStackOption {
	return func(s stack.Stack) error {
		cdktf.Aspects_Of(s.Stack).Add(&enforceTerraformVersion{
			TerraformVersion: terraformVersion,
		})
		return nil
	}
}

type enforceTerraformVersion struct {
	TerraformVersion string
}

var _ cdktf.IAspect = (*enforceTerraformVersion)(nil)

// Visit implements the aspect interface.
func (e *enforceTerraformVersion) Visit(node constructs.IConstruct) {
	switch reflect.TypeOf(node).String() {
	// It is not possible to check the type because the type is not exported.
	case "*cdktf.jsiiProxy_TerraformStack":
		s := node.(cdktf.TerraformStack)
		s.AddOverride(pointers.Ptr("terraform.required_version"),
			fmt.Sprintf("~> %s", e.TerraformVersion))
	}
}
