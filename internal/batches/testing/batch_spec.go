package testing

import (
	"context"
	"testing"

	"gopkg.in/yaml.v2"

	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

type CreateBatchSpecer interface {
	CreateBatchSpec(ctx context.Context, batchSpec *btypes.BatchSpec) error
}

func CreateBatchSpec(t *testing.T, ctx context.Context, store CreateBatchSpecer, name string, userID int32, bcID int64) *btypes.BatchSpec {
	t.Helper()

	rawSpec, err := yaml.Marshal(struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
	}{Name: name, Description: "the description"})
	if err != nil {
		t.Fatal(err)
	}

	s := &btypes.BatchSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec: &batcheslib.BatchSpec{
			Name:        name,
			Description: "the description",
			ChangesetTemplate: &batcheslib.ChangesetTemplate{
				Branch: "branch-name",
			},
		},
		RawSpec:       string(rawSpec),
		BatchChangeID: bcID,
	}

	if err := store.CreateBatchSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}

func CreateEmptyBatchSpec(t *testing.T, ctx context.Context, store CreateBatchSpecer, name string, userID int32, bcID int64) *btypes.BatchSpec {
	t.Helper()

	rawSpec, err := yaml.Marshal(struct {
		Name string `yaml:"name"`
	}{Name: name})
	if err != nil {
		t.Fatal(err)
	}

	s := &btypes.BatchSpec{
		UserID:          userID,
		NamespaceUserID: userID,
		Spec:            &batcheslib.BatchSpec{Name: name},
		RawSpec:         string(rawSpec),
		BatchChangeID:   bcID,
	}

	if err := store.CreateBatchSpec(ctx, s); err != nil {
		t.Fatal(err)
	}

	return s
}
