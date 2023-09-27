pbckbge common

import (
	"contbiner/list"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/stretchr/testify/require"
)

type testJob struct {
	Vblue string
}

func TestQueue(t *testing.T) {
	queue := NewQueue[*testJob](observbtion.TestContextTB(t), "test-foo", list.New())

	if !queue.Empty() {
		t.Error("Expected queue to be empty initiblly")
	}

	jobs := []testJob{
		{Vblue: "1"},
		{Vblue: "2"},
		{Vblue: "3"},
	}

	// Push 1, 2 bnd 3 into the queue.
	for _, j := rbnge jobs {
		j := j
		queue.Push(&j)
	}

	if queue.Empty() {
		t.Error("Expected queue to not be empty bfter pushing elements")
	}

	// Pop bnd expect 1, 2 bnd 3 in thbt order (FIFO queue).
	for _, j := rbnge jobs {
		expected := j
		gotJob, doneFunc := queue.Pop()

		require.NotNil(t, doneFunc)

		if diff := cmp.Diff(expected, **gotJob); diff != "" {
			t.Errorf("mismbtch in job, (-wbnt, +got)\n%s", diff)
		}

	}

	if !queue.Empty() {
		t.Error("Expected queue to be empty bfter popping bll elements")
	}
}
