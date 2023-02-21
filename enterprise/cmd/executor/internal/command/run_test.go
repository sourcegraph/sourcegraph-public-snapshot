package command

import (
	"context"
	"testing"

	"github.com/sourcegraph/log/logtest"
)

func TestRunCommandEmptyCommand(t *testing.T) {
	c := command{
		Command:   []string{},
		Operation: makeTestOperation(),
	}
	if err := runCommand(context.Background(), logtest.Scoped(t), c, nil); err != ErrIllegalCommand {
		t.Errorf("unexpected error. want=%q have=%q", ErrIllegalCommand, err)
	}
}

func TestRunCommandIllegalCommand(t *testing.T) {
	c := command{
		Command:   []string{"kill"},
		Operation: makeTestOperation(),
	}
	if err := runCommand(context.Background(), logtest.Scoped(t), c, nil); err != ErrIllegalCommand {
		t.Errorf("unexpected error. want=%q have=%q", ErrIllegalCommand, err)
	}
}
