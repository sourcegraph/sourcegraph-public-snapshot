package worker

import (
	"fmt"
	"io"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

type buildState struct {
	build sourcegraph.BuildSpec
}

// newBuildState creates a new buildState object. It must be closed by
// calling end (to close the log file).
func newBuildState(build sourcegraph.BuildSpec) buildState {
	return buildState{
		build: build,
	}
}

// start marks the build as started.
func (s buildState) start(ctx context.Context) error {
	_, err := sourcegraph.NewClientFromContext(ctx).Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
		Build: s.build,
		Info:  sourcegraph.BuildUpdate{StartedAt: now()},
	})
	return err
}

// end updates the build's final state.
func (s buildState) end(ctx context.Context, execErr error) error {
	_, err := sourcegraph.NewClientFromContext(ctx).Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
		Build: s.build,
		Info: sourcegraph.BuildUpdate{
			Success: execErr == nil,
			Failure: execErr != nil,
			EndedAt: now(),
		},
	})
	return err
}

// createTasks creates new tasks.
func (s buildState) createTasks(ctx context.Context, labels ...string) ([]buildTaskState, error) {
	tasks := make([]*sourcegraph.BuildTask, len(labels))
	for i, label := range labels {
		tasks[i] = &sourcegraph.BuildTask{Label: label}
	}
	createdTasks, err := sourcegraph.NewClientFromContext(ctx).Builds.CreateTasks(ctx, &sourcegraph.BuildsCreateTasksOp{
		Build: s.build,
		Tasks: tasks,
	})
	if err != nil {
		return nil, err
	}
	states := make([]buildTaskState, len(createdTasks.BuildTasks))
	for i, task := range createdTasks.BuildTasks {
		states[i] = buildTaskState{
			task: task.Spec(),
			log:  newLogger(task.Spec()),
		}
	}
	return states, nil
}

type buildTaskState struct {
	task sourcegraph.TaskSpec

	// log is where task logs are written. Internal errors
	// encountered by the builder are not written to w but are
	// returned as errors from its methods.
	log io.WriteCloser
}

// start marks the task as having started.
func (s buildTaskState) start(ctx context.Context) error {
	_, err := sourcegraph.NewClientFromContext(ctx).Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
		Task: s.task,
		Info: sourcegraph.TaskUpdate{
			StartedAt: now(),
		},
	})
	if err != nil {
		fmt.Fprintf(s.log, "Error starting task: %s\n", err)
	}
	return err
}

// skip marks the task as having been skipped.
func (s buildTaskState) skip(ctx context.Context) error {
	_, err := sourcegraph.NewClientFromContext(ctx).Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
		Task: s.task,
		Info: sourcegraph.TaskUpdate{
			Skipped: true,
			EndedAt: now(),
		},
	})
	if err != nil {
		fmt.Fprintf(s.log, "Error marking task as skipped: %s\n", err)
	}
	return err
}

// warnings marks the task as having warnings.
func (s buildTaskState) warnings(ctx context.Context) error {
	_, err := sourcegraph.NewClientFromContext(ctx).Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
		Task: s.task,
		Info: sourcegraph.TaskUpdate{Warnings: true},
	})
	if err != nil {
		fmt.Fprintf(s.log, "Error marking task as having warnings: %s\n", err)
	}
	return err
}

// end updates the task's final state.
func (s buildTaskState) end(ctx context.Context, execErr error) error {
	defer s.log.Close()

	_, err := sourcegraph.NewClientFromContext(ctx).Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
		Task: s.task,
		Info: sourcegraph.TaskUpdate{
			Success: execErr == nil,
			Failure: execErr != nil,
			EndedAt: now(),
		},
	})
	if err != nil {
		fmt.Fprintf(s.log, "Error ending build task: %s\n", err)
	}
	return err
}

func (s buildTaskState) createSubtask(ctx context.Context, label string) (*buildTaskState, error) {
	tasks, err := sourcegraph.NewClientFromContext(ctx).Builds.CreateTasks(ctx, &sourcegraph.BuildsCreateTasksOp{
		Build: s.task.Build,
		Tasks: []*sourcegraph.BuildTask{
			{Label: label, ParentID: s.task.ID},
		},
	})
	if err != nil {
		fmt.Fprintf(s.log, "Error creating subtask with label %q: %s\n", label, err)
		return nil, err
	}
	subtask := tasks.BuildTasks[0].Spec()
	return &buildTaskState{
		task: subtask,
		log:  newLogger(subtask),
	}, nil
}
