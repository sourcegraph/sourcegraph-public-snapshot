package images

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// UpdateOperation returns a new container repository raw image field to be used in
// manifests in the given registry.
type UpdateOperation func(registry Registry, repo *Repository) (*Repository, error)

func UpdateK8sManifest(ctx context.Context, registry Registry, path string, op UpdateOperation) error {
	updater := imageUpdater{
		registry: registry,
		op:       op,
	}

	rw := &kio.LocalPackageReadWriter{
		KeepReaderAnnotations: false,
		PreserveSeqIndent:     true,
		PackagePath:           path,
		PackageFileName:       "",
		IncludeSubpackages:    true,
		ErrorIfNonResources:   false,
		OmitReaderAnnotations: false,
		SetAnnotations:        nil,
		NoDeleteFiles:         true, //modify in place
		WrapBareSeqNode:       false,
	}

	err := kio.Pipeline{
		Inputs:                []kio.Reader{rw},
		Filters:               []kio.Filter{updater},
		Outputs:               []kio.Writer{rw},
		ContinueOnEmptyResult: true,
	}.Execute()

	return err
}

var conventionalInitContainerPaths = [][]string{
	{"spec", "initContainers"},
	{"spec", "template", "spec", "initContainers"},
}

type imageUpdater struct {
	registry Registry
	op       UpdateOperation
}

var _ kio.Filter = &imageUpdater{}

func (f imageUpdater) Filter(inputs []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, node := range inputs {
		switch kind := node.GetKind(); kind {
		case "Deployment", "StatefulSet", "DaemonSet", "Job": // We only care about those.
			// Find all containers "image" fields
			containers, err := node.Pipe(yaml.LookupFirstMatch(yaml.ConventionalContainerPaths))
			if err != nil {
				return nil, errors.Wrapf(err, "containers %q", node.GetName())
			}
			initContainers, err := node.Pipe(yaml.LookupFirstMatch(conventionalInitContainerPaths))
			if err != nil {
				return nil, errors.Wrapf(err, "init containers %q", node.GetName())
			}
			if containers == nil && initContainers == nil {
				return nil, ErrNoImage{
					Kind: node.GetKind(),
					Name: node.GetName(),
				}
			}

			// Wrap access to the image field
			parentNode := node
			updateFn := func(node *yaml.RNode) error {
				// Find the previous value.
				oldImage, err := f.lookup(node, parentNode)
				if err != nil {
					if errors.As(err, &ErrNoImage{Kind: parentNode.GetKind(), Name: parentNode.GetName()}) {
						return nil
					}
					return err
				}
				r, err := ParseRepository(oldImage)
				if err != nil {
					return err
				}

				// Compute the new image field value, using the given UpdateOperation
				newRepo, err := f.op(f.registry, r)
				if err != nil {
					if errors.Is(err, ErrNoUpdateNeeded) {
						std.Out.WriteLine(output.Styled(output.StyleWarning, fmt.Sprintf("skipping %q", oldImage)))
						return nil
					} else {
						return err
					}
				}
				// Update the field in-place.
				return node.PipeE(yaml.Lookup("image"), yaml.Set(yaml.NewStringRNode(newRepo.Ref())))
			}

			// Apply the above on normal containers fields..
			if err := containers.VisitElements(updateFn); err != nil {
				return nil, err
			}
			// And if needed, also apply it on init containers.
			if initContainers != nil {
				if err := initContainers.VisitElements(updateFn); err != nil {
					return nil, err
				}
			}
		default:
			std.Out.Verbosef("Skipping manifest of kind: %v", kind)
			continue
		}
	}

	return inputs, nil
}

func (f imageUpdater) lookup(node *yaml.RNode, parentNode *yaml.RNode) (string, error) {
	imageNode := node.Field("image")
	if imageNode == nil {
		return "", errors.Wrapf(ErrNoImage{parentNode.GetKind(), parentNode.GetName()}, "couldn't find image for container %s", parentNode.GetName())
	}
	image, err := imageNode.Value.String()
	if err != nil {
		return "", errors.Wrapf(err, "%s: invalid image", node.GetName())
	}
	return image, nil
}
