package stack

import (
	"bytes"
	"sort"
	"strconv"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type TFVars map[string]string

func (v TFVars) RenderTFVarsFile() []byte {
	if len(v) == 0 {
		return []byte{'\n'}
	}
	keys := maps.Keys(v)
	sort.Strings(keys)
	var b bytes.Buffer
	for _, k := range keys {
		b.WriteString(k)
		b.WriteString(" = ")
		b.WriteString(strconv.Quote(v[k]))
		b.WriteByte('\n')
	}
	return b.Bytes()
}

// Stack encapsulates a CDKTF stack and the name the stack was originally
// created with.
type Stack struct {
	Name  string
	Stack cdktf.TerraformStack
	// Metadata is arbitrary metadata that can be attached by stack options.
	Metadata map[string]string
	// DynamicVariables are rendered into a tfvar file, and should be used to
	// track anything that requires an external datasource to populate up-front
	// at generation time.
	DynamicVariables TFVars
}

// Set collects the stacks that comprise a CDKTF application.
type Set struct {
	// app represents a CDKTF application that is comprised of the stacks in
	// this set.
	//
	// The App can be extracted with stack.ExtractApp(*Set)
	app cdktf.App
	// opts are applied to all the stacks created from (*Set).New()
	opts []NewStackOption
	// stacks is all the stacks created from (*Set).New()
	//
	// Names of created stacks can be extracted with stack.ExtractStacks(*Set)
	stacks []Stack
}

// NewStackOption applies modifications to cdktf.TerraformStacks when they are
// created.
type NewStackOption func(s Stack) error

// NewSet creates a new stack.Set, which collects the stacks that comprise a
// CDKTF application.
func NewSet(renderDir string, opts ...NewStackOption) *Set {
	return &Set{
		app: cdktf.NewApp(&cdktf.AppConfig{
			Outdir: pointers.Ptr(renderDir),
		}),
		opts:   opts,
		stacks: []Stack{},
	}
}

// ExplicitStackOutputs adds an explicit output to the Terraform stack for human
// reference. This is separate from outputs returned explicitly from Stack
// constructors, which are intended for programatic access.
type ExplicitStackOutputs struct{ s cdktf.TerraformStack }

// Add renders a non-sensitive key-value pair as part of the workspace outputs.
func (o *ExplicitStackOutputs) Add(name string, value any) {
	_ = cdktf.NewTerraformOutput(o.s,
		resourceid.New("output").ResourceID(name),
		&cdktf.TerraformOutputConfig{
			Value:     value,
			Sensitive: pointers.Ptr(false),
		})
}

// New creates a new stack belonging to this set.
func (s *Set) New(name string, opts ...NewStackOption) (cdktf.TerraformStack, ExplicitStackOutputs, error) {
	stack := Stack{
		Name:             name,
		Stack:            cdktf.NewTerraformStack(s.app, &name),
		Metadata:         make(map[string]string),
		DynamicVariables: make(map[string]string),
	}
	for _, opt := range append(s.opts, opts...) {
		if err := opt(stack); err != nil {
			return nil, ExplicitStackOutputs{}, err
		}
	}
	s.stacks = append(s.stacks, stack)
	return stack.Stack, ExplicitStackOutputs{stack.Stack}, nil
}

// ExtractApp returns the underlying CDKTF application of this stack.Set for
// synthesizing resources.
//
// It is intentionally not part of the stack.Set interface as it should not
// generally be needed.
func ExtractApp(set *Set) cdktf.App { return set.app }

// ExtractStacks returns all the stacks created so far in this stack.Set.
//
// It is intentionally not part of the stack.Set interface as it should not
// generally be needed.
func ExtractStacks(set *Set) []Stack { return set.stacks }
