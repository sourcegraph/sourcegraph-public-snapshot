package codehost_testing

import (
	"context"
	"errors"
	"testing"
)

func testAction(name string) *Action {
	return &Action{
		Name: name,
		Apply: func(ctx context.Context) error {
			return nil
		},
		Teardown: func(ctx context.Context) error {
			return nil
		},
	}
}

func TestScenarioApplyAndTeardown(t *testing.T) {
	t.Run("Applied actions updates nextActionIdx", func(t *testing.T) {
		scenario := &GitHubScenario{
			id:            "testing-id",
			t:             t,
			client:        nil,
			actions:       []*Action{},
			reporter:      NoopReporter{},
			nextActionIdx: 0,
		}

		scenario.Append(testAction("t1"), testAction("t2"))

		if len(scenario.actions) != 2 {
			t.Errorf("actions not appended - got %d wanted %d", len(scenario.actions), 2)
		}

		err := scenario.Apply(context.TODO())
		if err != nil {
			t.Fatalf("failed to apply test scenario with mock actions: %v", err)
		}

		if scenario.nextActionIdx != 2 {
			t.Errorf("actions applied count mismatch - got %d wanted %d", scenario.nextActionIdx, 2)
		}

		if !scenario.IsApplied() {
			t.Error("all actions have been applied thus IsApplied should be true")
		}
	})
	t.Run("Next Action Idx at 1 if remaining action errors", func(t *testing.T) {
		scenario := &GitHubScenario{
			id:            "testing-id",
			t:             t,
			client:        nil,
			actions:       []*Action{},
			reporter:      NoopReporter{},
			nextActionIdx: 0,
		}

		errAction := testAction("err1")
		fakeErr := errors.New("fake error")
		errAction.Apply = func(ctx context.Context) error {
			return fakeErr
		}
		scenario.Append(testAction("t1"), errAction)

		if len(scenario.actions) != 2 {
			t.Errorf("actions not appended - got %d wanted %d", len(scenario.actions), 2)
		}

		err := scenario.Apply(context.TODO())
		if err != nil && !errors.Is(err, fakeErr) {
			t.Fatalf("failed to apply test scenario with mock actions: %v", err)
		}

		if scenario.nextActionIdx != 1 {
			t.Errorf("actions applied count mismatch - got %d wanted %d", scenario.nextActionIdx, 1)
		}

		if scenario.IsApplied() {
			t.Error("not all actions have been applied thus IsApplied should be false")
		}

		err = scenario.Teardown(context.TODO())
		if err != nil {
			t.Fatalf("teardown not expected to fail here: %v", err)
		}
		if scenario.IsApplied() {
			t.Error("after teardown, IsApplied should be false")
		}
		if scenario.nextActionIdx != 0 {
			t.Errorf("after teardown nextActionIdx should be 0 - got %d", scenario.nextActionIdx)
		}
	})
	t.Run("3 Actions with 1 skipped teardown", func(t *testing.T) {
		scenario := &GitHubScenario{
			id:            "testing-id",
			t:             t,
			client:        nil,
			actions:       []*Action{},
			reporter:      NoopReporter{},
			nextActionIdx: 0,
		}

		skipAction := testAction("s2")
		skipAction.Teardown = nil
		scenario.Append(testAction("t1"), skipAction, testAction("t3"))

		if len(scenario.actions) != 3 {
			t.Errorf("actions not appended - got %d wanted %d", len(scenario.actions), 2)
		}

		scenario.Apply(context.TODO())

		if scenario.nextActionIdx != 3 {
			t.Errorf("actions applied count mismatch - got %d wanted %d", scenario.nextActionIdx, 1)
		}

		if !scenario.IsApplied() {
			t.Error("all actions should be applied")
		}

		err := scenario.Teardown(context.TODO())
		if err != nil {
			t.Fatalf("teardown not expected to fail here: %v", err)
		}
		if scenario.IsApplied() {
			t.Error("after teardown, scenario should not be Applied")
		}
		if scenario.nextActionIdx != 0 {
			t.Errorf("after teardown nextActionIdx should be 0 - got %d", scenario.nextActionIdx)
		}
	})
	t.Run("4 Actions with 2 failed teardown", func(t *testing.T) {
		scenario := &GitHubScenario{
			id:            "testing-id",
			t:             t,
			client:        nil,
			actions:       []*Action{},
			reporter:      NoopReporter{},
			nextActionIdx: 0,
		}

		errTeardown := func(_ context.Context) error { return errors.New("fake") }
		errAction := testAction("e2")
		errAction.Teardown = errTeardown
		scenario.Append(testAction("t1"), errAction, testAction("t3"))
		errTeardown = func(_ context.Context) error { return errors.New("fake") }
		errAction = testAction("e4")
		errAction.Teardown = errTeardown
		scenario.Append(errAction)

		if len(scenario.actions) != 4 {
			t.Errorf("actions not appended - got %d wanted %d", len(scenario.actions), 2)
		}

		scenario.Apply(context.TODO())

		if scenario.nextActionIdx != 4 {
			t.Errorf("actions applied count mismatch - got %d wanted %d", scenario.nextActionIdx, 4)
		}

		if !scenario.IsApplied() {
			t.Error("all actions should be applied")
		}

		scenario.Teardown(context.TODO())
		if scenario.IsApplied() {
			t.Error("after teardown, scenario should not be Applied")
		}

		if scenario.nextActionIdx != 0 {
			t.Errorf("after teardown nextActionIdx should be 0 - got %d", scenario.nextActionIdx)
		}
	})
}
