package tfvar

import (
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/hashicorp/terraform-cdk-go/cdktf"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type Config struct {
	VariableKey string
	Description string
}

type Output struct {
	StringValue *string // only support strings for now
}

func New(scope constructs.Construct, id resourceid.ID, config Config) *Output {
	v := cdktf.NewTerraformVariable(scope, id.TerraformID(config.VariableKey),
		&cdktf.TerraformVariableConfig{
			Type:        pointers.Ptr("string"), // only strings for now
			Description: pointers.Ptr(config.Description),
			Sensitive:   pointers.Ptr(false),
		})
	v.OverrideLogicalId(&config.VariableKey)
	return &Output{StringValue: v.StringValue()}
}
