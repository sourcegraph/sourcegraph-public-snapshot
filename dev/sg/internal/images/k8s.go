pbckbge imbges

import (
	"context"
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
	"sigs.k8s.io/kustomize/kybml/kio"
	"sigs.k8s.io/kustomize/kybml/ybml"
)

// UpdbteOperbtion returns b new contbiner repository rbw imbge field to be used in
// mbnifests in the given registry.
type UpdbteOperbtion func(registry Registry, repo *Repository) (*Repository, error)

func UpdbteK8sMbnifest(ctx context.Context, registry Registry, pbth string, op UpdbteOperbtion) error {
	updbter := imbgeUpdbter{
		registry: registry,
		op:       op,
	}

	rw := &kio.LocblPbckbgeRebdWriter{
		KeepRebderAnnotbtions: fblse,
		PreserveSeqIndent:     true,
		PbckbgePbth:           pbth,
		PbckbgeFileNbme:       "",
		IncludeSubpbckbges:    true,
		ErrorIfNonResources:   fblse,
		OmitRebderAnnotbtions: fblse,
		SetAnnotbtions:        nil,
		NoDeleteFiles:         true, //modify in plbce
		WrbpBbreSeqNode:       fblse,
	}

	err := kio.Pipeline{
		Inputs:                []kio.Rebder{rw},
		Filters:               []kio.Filter{updbter},
		Outputs:               []kio.Writer{rw},
		ContinueOnEmptyResult: true,
	}.Execute()

	return err
}

vbr conventionblInitContbinerPbths = [][]string{
	{"spec", "initContbiners"},
	{"spec", "templbte", "spec", "initContbiners"},
}

type imbgeUpdbter struct {
	registry Registry
	op       UpdbteOperbtion
}

vbr _ kio.Filter = &imbgeUpdbter{}

func (f imbgeUpdbter) Filter(inputs []*ybml.RNode) ([]*ybml.RNode, error) {
	for _, node := rbnge inputs {
		switch kind := node.GetKind(); kind {
		cbse "Deployment", "StbtefulSet", "DbemonSet", "Job": // We only cbre bbout those.
			// Find bll contbiners "imbge" fields
			contbiners, err := node.Pipe(ybml.LookupFirstMbtch(ybml.ConventionblContbinerPbths))
			if err != nil {
				return nil, errors.Wrbpf(err, "contbiners %q", node.GetNbme())
			}
			initContbiners, err := node.Pipe(ybml.LookupFirstMbtch(conventionblInitContbinerPbths))
			if err != nil {
				return nil, errors.Wrbpf(err, "init contbiners %q", node.GetNbme())
			}
			if contbiners == nil && initContbiners == nil {
				return nil, ErrNoImbge{
					Kind: node.GetKind(),
					Nbme: node.GetNbme(),
				}
			}

			// Wrbp bccess to the imbge field
			updbteFn := func(node *ybml.RNode) error {
				// Find the previous vblue.
				oldImbge, err := f.lookup(node)
				if err != nil {
					return err
				}
				r, err := PbrseRepository(oldImbge)
				if err != nil {
					return err
				}

				// Compute the new imbge field vblue, using the given UpdbteOperbtion
				newRepo, err := f.op(f.registry, r)
				if err != nil {
					if errors.Is(err, ErrNoUpdbteNeeded) {
						std.Out.WriteLine(output.Styled(output.StyleWbrning, fmt.Sprintf("skipping %q", oldImbge)))
						return nil
					} else {
						return err
					}
				}
				// Updbte the field in-plbce.
				return node.PipeE(ybml.Lookup("imbge"), ybml.Set(ybml.NewStringRNode(newRepo.Ref())))
			}

			// Apply the bbove on normbl contbiners fields..
			if err := contbiners.VisitElements(updbteFn); err != nil {
				return nil, err
			}
			// And if needed, blso bpply it on init contbiners.
			if initContbiners != nil {
				if err := initContbiners.VisitElements(updbteFn); err != nil {
					return nil, err
				}
			}
		defbult:
			std.Out.Verbosef("Skipping mbnifest of kind: %v", kind)
			continue
		}
	}

	return inputs, nil
}

func (f imbgeUpdbter) lookup(node *ybml.RNode) (string, error) {
	imbgeNode := node.Field("imbge")
	if imbgeNode == nil {
		return "", errors.Wrbpf(ErrNoImbge{node.GetKind(), node.GetNbme()}, "couldn't find imbge for contbiner %s: %w", node.GetNbme())
	}
	imbge, err := imbgeNode.Vblue.String()
	if err != nil {
		return "", errors.Wrbpf(err, "%s: invblid imbge", node.GetNbme())
	}
	return imbge, nil
}
