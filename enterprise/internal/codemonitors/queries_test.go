package codemonitors

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestQueryByRecordID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, s := newTestStore(t)
	_, id, _, userCTX := newTestUser(ctx, t)
	m, err := s.insertTestMonitor(userCTX, t)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueTriggerQueries(ctx)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetQueryByRecordID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := &MonitorQuery{
		Id:           1,
		Monitor:      m.ID,
		QueryString:  testQuery,
		NextRun:      s.Now(),
		LatestResult: nil,
		CreatedBy:    id,
		CreatedAt:    s.Now(),
		ChangedBy:    id,
		ChangedAt:    s.Now(),
	}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("diff: %s", diff)
	}
}

func TestTriggerQueryNextRun(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	ctx, s := newTestStore(t)
	_, id, _, userCTX := newTestUser(ctx, t)
	m, err := s.insertTestMonitor(userCTX, t)
	if err != nil {
		t.Fatal(err)
	}
	err = s.EnqueueTriggerQueries(ctx)
	if err != nil {
		t.Fatal(err)
	}

	wantLatestResult := s.Now().Add(time.Minute)
	wantNextRun := s.Now().Add(time.Hour)

	err = s.SetTriggerQueryNextRun(ctx, 1, wantNextRun, wantLatestResult)
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.GetQueryByRecordID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := &MonitorQuery{
		Id:           1,
		Monitor:      m.ID,
		QueryString:  testQuery,
		NextRun:      wantNextRun,
		LatestResult: &wantLatestResult,
		CreatedBy:    id,
		CreatedAt:    s.Now(),
		ChangedBy:    id,
		ChangedAt:    s.Now(),
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Fatalf("diff: %s", diff)
	}
}
