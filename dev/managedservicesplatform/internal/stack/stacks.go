package stack

import (
	"bytes"
	"sort"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"golang.org/x/exp/maps"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resource/gsmsecret"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/internal/resourceid"
	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/stacks"
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

// Locals allows stack options to add local variables for reference in custom
// Terraform and outputs.
func (s Stack) Locals() *StackLocals { return &StackLocals{s} }

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

// StackLocals adds an explicit output to the Terraform stack for human
// reference. This is separate from outputs returned explicitly from Stack
// constructors, which are intended for programatic access. It also adds
// local variables for reference by custom resources.
type StackLocals struct{ s Stack }

// MetadataKeyStackLocalsGSMProjectID, if set on stack.Metadata, configures
// StackLocals to also add locals to GSM in the given project.
const MetadataKeyStackLocalsGSMProjectID = "locals_gsm_project_id"

// Add renders a non-sensitive key-value pair as part of the workspace outputs,
// under the resource ID 'output-${name}'.
//
// The value is also available to locals under '${name}', so that they can
// accessed under 'local.${name}' in custom resources.
//
// If a StackOption is used to create the stack that configures
// MetadataKeyStackLocalsGSMProjectID, then the added value will also be stored
// in GSM. This allows tooling to access stack outputs and locals from GSM.
func (l *StackLocals) Add(name string, value string, description string) {
	id := resourceid.New("output")
	_ = cdktf.NewTerraformOutput(l.s.Stack,
		id.TerraformID(name),
		&cdktf.TerraformOutputConfig{
			Value:       value,
			Sensitive:   pointers.Ptr(false),
			Description: &description,
		})
	_ = cdktf.NewTerraformLocal(l.s.Stack, &name, value)

	l.maybeEmitToGSM(id, name, value)
}

// AddSlice is the same as Add, but accepts a slice instead. The value is
// represented as a comma-separated string in GSM.
func (l *StackLocals) AddSlice(name string, value []string, description string) {
	id := resourceid.New("output")
	_ = cdktf.NewTerraformOutput(l.s.Stack,
		id.TerraformID(name),
		&cdktf.TerraformOutputConfig{
			Value:       value,
			Sensitive:   pointers.Ptr(false),
			Description: &description,
		})
	_ = cdktf.NewTerraformLocal(l.s.Stack, &name, value)

	l.maybeEmitToGSM(id, name, strings.Join(value, ","))
}

func (l *StackLocals) maybeEmitToGSM(id resourceid.ID, name, value string) {
	// If MetadataKeyStackLocalsGSMProjectID is set, emit to GSM
	if project, ok := l.s.Metadata[MetadataKeyStackLocalsGSMProjectID]; ok {
		_ = gsmsecret.New(l.s.Stack, id.Group("gsm").Group(name), gsmsecret.Config{
			ID:        stacks.OutputSecretID(l.s.Name, name),
			ProjectID: project,
			Value:     value,
		})
	}
}

// New creates a new stack belonging to this set.
func (s *Set) New(name string, opts ...NewStackOption) (cdktf.TerraformStack, StackLocals, error) {
	stack := Stack{
		Name:             name,
		Stack:            cdktf.NewTerraformStack(s.app, &name),
		Metadata:         make(map[string]string),
		DynamicVariables: make(map[string]string),
	}
	for _, opt := range append(s.opts, opts...) {
		if err := opt(stack); err != nil {
			return nil, StackLocals{}, err
		}
	}
	s.stacks = append(s.stacks, stack)
	return stack.Stack, StackLocals{stack}, nil
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

// ExtractStacks returns the "current" (last) stack in this stack.Set.
//
// It is intentionally not part of the stack.Set interface as it should not
// generally be needed.
func ExtractCurrentStack(set *Set) Stack { return set.stacks[len(set.stacks)-1] }
