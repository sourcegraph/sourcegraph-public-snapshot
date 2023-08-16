package stack

import (
	"github.com/aws/jsii-runtime-go"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type Stack struct {
	Name  string
	Stack cdktf.TerraformStack
}

type Set struct {
	// App represents a CDKTF application that is comprised of the stacks in
	// this set.
	App cdktf.App
	// stacks is all the stacks created from (*Set).New()
	stacks []Stack
}

func NewSet(renderDir string) *Set {
	return &Set{
		App: cdktf.NewApp(&cdktf.AppConfig{
			Outdir: jsii.String(renderDir),
		}),
		stacks: []Stack{},
	}
}

type NewStackOption func(s cdktf.TerraformStack)

// New creates a new stack belonging to this set.
func (s *Set) New(name string, opts ...NewStackOption) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(s.App, &name)
	for _, opt := range opts {
		opt(stack)
	}
	s.stacks = append(s.stacks, Stack{
		Name:  name,
		Stack: stack,
	})
	return stack
}

// GetStacks returns all the stacks created so far.
func (s *Set) GetStacks() []Stack { return s.stacks }
