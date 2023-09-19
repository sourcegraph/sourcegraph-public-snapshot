package tst

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type actionRunner struct {
	T        *testing.T
	Reporter Reporter
	setup    []Action
	teardown []Action
}

type ActionResult interface {
	Get() any
}

type ActionFn func(ctx context.Context, store *ScenarioStore) (ActionResult, error)

type Action interface {
	Name() string
	Hash() []byte
	Complete() bool
	Do(ctx context.Context, t *testing.T, store *ScenarioStore) (ActionResult, error)
	String() string
}

type action struct {
	id       string
	name     string
	hash     []byte
	complete bool
	fn       ActionFn
}

func (a *action) Do(ctx context.Context, t *testing.T, store *ScenarioStore) (ActionResult, error) {
	t.Helper()
	result, err := a.fn(ctx, store)
	a.complete = true
	return result, err
}

func (a *action) Hash() []byte {
	return a.hash
}

func (a *action) Name() string {
	return a.name
}

func (a *action) Complete() bool {
	return a.complete
}

func (a *action) String() string {
	return fmt.Sprintf("%s (%s)", a.name, a.id)
}

type actionResult[T any] struct {
	item T
}

func (a *actionResult[T]) Get() any {
	return a.item
}

func NewActionManager(t *testing.T) *actionRunner {
	return &actionRunner{
		T:        t,
		setup:    make([]Action, 0),
		teardown: make([]Action, 0),
		Reporter: NoopReporter{},
	}
}

func (m *actionRunner) AddSetup(actions ...Action) {
	m.setup = append(m.setup, actions...)
}
func (m *actionRunner) AddTeardown(actions ...Action) {
	m.teardown = append(m.teardown, actions...)
}

func (m *actionRunner) setupPlan() string {
	b := strings.Builder{}
	for _, a := range m.setup {
		b.WriteString(a.String())
		b.WriteByte('\n')
	}

	return b.String()
}

func (m *actionRunner) teardownPlan() string {
	b := strings.Builder{}
	actions := m.teardown
	for i := len(actions) - 1; i >= 0; i-- {
		b.WriteString(actions[i].String())
		b.WriteByte('\n')
	}

	return b.String()
}

func (m *actionRunner) String() string {
	b := strings.Builder{}
	b.WriteString("Setup\n")
	b.WriteString("======\n")
	b.WriteString(m.setupPlan())
	b.WriteByte('\n')
	b.WriteString("Teardown\n")
	b.WriteString("========\n")
	b.WriteString(m.teardownPlan())
	return b.String()
}

func (m *actionRunner) Apply(ctx context.Context, store *ScenarioStore, actions []Action, failFast bool) error {
	m.T.Helper()
	var errs errors.MultiError
	for _, action := range actions {
		m.Reporter.Writef("Applying '%s' = ", action)
		now := time.Now().UTC()

		var err error
		if !action.Complete() {
			// TODO(burmudar): ActionsFn currently returns a result and error and we currently ignore the result, since
			// it is inserted into the store by the action instead. Should we rework this so that we actually use the
			// action result? Or should we just let actions return an err and manage the store?
			_, err = action.Do(ctx, m.T, store)
		} else {
			m.Reporter.Writeln("[SKIPPED]")
			continue
		}

		duration := time.Now().UTC().Sub(now)
		if err != nil {
			if failFast {
				m.Reporter.Writef("[FAILED] (%s)\n", duration.String())
				return err
			} else {
				m.Reporter.Writef("[FAILED] (%s)\n", duration.String())
				errs = errors.Append(errs, err)
			}
		} else {
			m.Reporter.Writef("[SUCCESS] (%s)\n", duration.String())
		}
	}
	return errs
}
