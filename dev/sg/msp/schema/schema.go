package schema

import (
	"encoding/json"
	"os"

	"github.com/invopop/jsonschema"

	"github.com/sourcegraph/sourcegraph/dev/managedservicesplatform/spec"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Render renders a JSON schema for spec.Spec, using the same mechanism used
// in sourcegraph/controller. It must be run from the sourcegraph/sourcegraph
// repository root.
func Render() ([]byte, error) {
	// We must be in repo root to extract Go comments correctly
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, errors.Wrap(err, "must be in sourcegraph/sourcegraph repository")
	}
	if err := os.Chdir(repoRoot); err != nil {
		return nil, errors.Wrap(err, "must be in sourcegraph/sourcegraph repository")
	}

	var r jsonschema.Reflector
	if err := r.AddGoComments(
		"github.com/sourcegraph/sourcegraph",
		"./dev/managedservicesplatform/spec",
	); err != nil {
		return nil, errors.Wrap(err, "failed to extract Go comments")
	}
	if len(r.CommentMap) == 0 {
		return nil, errors.New("failed to extract Go comments")
	}

	jsonSchema := r.Reflect(spec.Spec{})
	if jsonSchema == nil {
		return nil, errors.Newf("failed to reflect on %T", spec.Spec{})
	}
	b, err := json.MarshalIndent(jsonSchema, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal jsonschema")
	}
	return b, nil
}
