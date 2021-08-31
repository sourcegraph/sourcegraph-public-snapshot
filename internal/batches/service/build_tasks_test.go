package service

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

func TestBuildTask_IfConditions(t *testing.T) {
	tests := map[string]struct {
		spec *batcheslib.BatchSpec

		wantSteps []batcheslib.Step
	}{
		"no if": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
			},
		},

		"if has static true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: "true"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: "true"},
			},
		},

		"one of many steps has if with static true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "true"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
				{Run: "echo 2", If: "true"},
				{Run: "echo 3"},
			},
		},

		"if has static non-true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: "this is not true"},
				},
			},
			wantSteps: nil,
		},

		"one of many steps has if with static non-true value": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1"},
					{Run: "echo 2", If: "every type system needs generics"},
					{Run: "echo 3"},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1"},
				{Run: "echo 3"},
			},
		},

		"if expression that can be partially evaluated to true": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: `${{ matches repository.name "github.com/sourcegraph/src*" }}`},
			},
		},

		"if expression that can be partially evaluated to false": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: nil,
		},

		"one of many steps has if expression that can be evaluated to true": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ matches repository.name "horse" }}`},
				},
			},
			wantSteps: nil,
		},

		"if expression that can NOT be partially evaluated": {
			spec: &batcheslib.BatchSpec{
				Steps: []batcheslib.Step{
					{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
				},
			},
			wantSteps: []batcheslib.Step{
				{Run: "echo 1", If: `${{ eq outputs.value "foobar" }}`},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			task, ok, err := buildTask(tt.spec, testRepo1, "", false)
			if err != nil {
				t.Fatalf("unexpected err: %s", err)
			}

			if !ok {
				if tt.wantSteps != nil {
					t.Fatalf("no task built, but steps expected")
				}
				return
			}

			opts := cmpopts.IgnoreUnexported(batcheslib.Step{})
			if diff := cmp.Diff(tt.wantSteps, task.Steps, opts); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
