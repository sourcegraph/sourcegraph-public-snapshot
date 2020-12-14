package command

import (
	"context"
	"testing"
)

func TestRunCommandEmptyCommand(t *testing.T) {
	if err := runCommand(context.Background(), nil, command{Command: []string{}}); err != ErrIllegalCommand {
		t.Errorf("unexpected error. want=%q have=%q", ErrIllegalCommand, err)
	}
}

func TestRunCommandIllegalCommand(t *testing.T) {
	if err := runCommand(context.Background(), nil, command{Command: []string{"kill"}}); err != ErrIllegalCommand {
		t.Errorf("unexpected error. want=%q have=%q", ErrIllegalCommand, err)
	}
}
