package worker

import (
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"

	"golang.org/x/net/context"
)

func buildCleanup(ctx context.Context, activeBuilds *activeBuilds) {
	activeBuilds.RLock()
	defer activeBuilds.RUnlock()
	// Mark all active builds (and their tasks) as killed. But set
	// an aggressive timeout so we don't block the termination for
	// too long.
	if len(activeBuilds.Builds) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	time.AfterFunc(500*time.Millisecond, func() {
		activeBuilds.RLock()
		defer activeBuilds.RUnlock()
		// Log if it's taking a noticeable amount of time.
		builds := make([]string, 0, len(activeBuilds.Builds))
		for b := range activeBuilds.Builds {
			builds = append(builds, b.Spec.IDString())
		}
		log15.Info("Marking active builds as killed before terminating...", "builds", builds)
	})
	for b := range activeBuilds.Builds {
		if err := markBuildAsKilled(ctx, b.Spec); err != nil {
			log15.Error("Error marking build as killed upon process termination", "build", b.Spec, "err", err)
		}
	}
}

func markBuildAsKilled(ctx context.Context, b sourcegraph.BuildSpec) error {
	cl, err := sourcegraph.NewClientFromContext(ctx)
	if err != nil {
		return err
	}

	_, err = cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
		Build: b,
		Info: sourcegraph.BuildUpdate{
			EndedAt: now(),
			Killed:  true,
		},
	})
	if err != nil {
		return err
	}

	// Mark all of the build's unfinished tasks as failed, too.
	for page := int32(1); ; page++ {
		tasks, err := cl.Builds.ListBuildTasks(ctx, &sourcegraph.BuildsListBuildTasksOp{
			Build: b,
			Opt:   &sourcegraph.BuildTaskListOptions{ListOptions: sourcegraph.ListOptions{Page: page}},
		})
		if err != nil {
			return err
		}

		for _, task := range tasks.BuildTasks {
			if task.EndedAt != nil {
				continue
			}
			_, err := cl.Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
				Task: task.Spec(),
				Info: sourcegraph.TaskUpdate{Failure: true, EndedAt: now()},
			})
			if err != nil {
				return err
			}
		}
		if len(tasks.BuildTasks) == 0 {
			break
		}
	}

	return nil
}
