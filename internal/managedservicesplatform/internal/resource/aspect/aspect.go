package aspect

import (
	"fmt"
	"reflect"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

// EnforceTerraformVersion is an aspect that enforces terraform version in a cdktf stack.
// cdktf does not provide a native way to configure terraform version,
// so we use an aspect to enforce it.
// Learn more https://developer.hashicorp.com/terraform/cdktf/concepts/aspects
type EnforceTerraformVersion struct {
	TerraformVersion string
}

var _ cdktf.IAspect = (*EnforceTerraformVersion)(nil)

// Visit implements the aspect interface.
func (e *EnforceTerraformVersion) Visit(node constructs.IConstruct) {
	switch reflect.TypeOf(node).String() {
	// It is not possible to check the type because the type is not exported.
	case "*cdktf.jsiiProxy_TerraformStack":
		s := node.(cdktf.TerraformStack)
		s.AddOverride(jsii.String("terraform.required_version"),
			fmt.Sprintf("~> %s", e.TerraformVersion))
	}
}
