package worker

import (
	"math/rand"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sqs/pbtypes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"

	"context"
)

// buildReaper periodically removes builds that have not sent a
// heartbeat within the allowable timeout. It should be run in a
// goroutine by itself.
func buildReaper(ctx context.Context) {
	// Add margin for clock skew.
	const heartbeatTimeout = serverHeartbeatInterval + 2*time.Minute

	listAllHeartbeatExpiredBuilds := func(ctx context.Context) ([]*sourcegraph.Build, error) {
		var expired []*sourcegraph.Build

		cl, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			return nil, err
		}
		for page := int32(1); ; page++ {
			builds, err := cl.Builds.List(ctx, &sourcegraph.BuildListOptions{
				Active:      true,
				ListOptions: sourcegraph.ListOptions{Page: page, PerPage: 100},
			})
			if err != nil {
				return nil, err
			}

			if len(builds.Builds) == 0 {
				break
			}

			for _, b := range builds.Builds {
				// A build that was just started might not yet have a
				// heartbeat, so accept the StartedAt time as the last
				// heartbeat-ish time.
				var last time.Time
				if b.HeartbeatAt != nil && b.HeartbeatAt.Time().After(last) {
					last = b.HeartbeatAt.Time()
				}
				if b.StartedAt != nil && b.StartedAt.Time().After(last) {
					last = b.StartedAt.Time()
				}
				if last.IsZero() {
					log15.Error("Active build has no StartedAt nor HeartbeatAt date", "build", b.Spec())
					continue
				}

				if time.Since(last) > heartbeatTimeout {
					expired = append(expired, b)
				}
			}
		}

		return expired, nil
	}

	for {
		// Random sleep to avoid thundering herd.
		time.Sleep(time.Minute * time.Duration(3+rand.Intn(5)))

		cl, err := sourcegraph.NewClientFromContext(ctx)
		if err != nil {
			log15.Error("Build reaper failed to create client", "err", err)
			continue
		}

		expiredBuilds, err := listAllHeartbeatExpiredBuilds(ctx)
		if err != nil {
			log15.Error("Error listing all heartbeat-expired builds", "err", err)
			continue
		}

		for _, b := range expiredBuilds {
			log15.Info("Marking heartbeat-expired build (and its unfinished tasks) as failed", "build", b.Spec())
			now := pbtypes.NewTimestamp(time.Now())

			_, err := cl.Builds.Update(ctx, &sourcegraph.BuildsUpdateOp{
				Build: b.Spec(),
				Info:  sourcegraph.BuildUpdate{Failure: true, Killed: true, EndedAt: &now},
			})
			if err != nil {
				log15.Error("Error updating heartbeat-expired build", "build", b.Spec(), "err", err)
			}

			// Mark all of the build's unfinished tasks as failed, too.
			for page := int32(1); ; page++ {
				tasks, err := cl.Builds.ListBuildTasks(ctx, &sourcegraph.BuildsListBuildTasksOp{
					Build: b.Spec(),
					Opt:   &sourcegraph.BuildTaskListOptions{ListOptions: sourcegraph.ListOptions{Page: page}},
				})
				if err != nil {
					log15.Error("Error listing heartbeat-expired build tasks", "build", b.Spec(), "err", err)
					break
				}

				for _, task := range tasks.BuildTasks {
					if task.EndedAt != nil {
						continue
					}
					_, err := cl.Builds.UpdateTask(ctx, &sourcegraph.BuildsUpdateTaskOp{
						Task: task.Spec(),
						Info: sourcegraph.TaskUpdate{Failure: true, EndedAt: &now},
					})
					if err != nil {
						log15.Error("Error updating heartbeat-expired build task", "task", task.Spec(), "err", err)
					}
				}
				if len(tasks.BuildTasks) == 0 {
					break
				}
			}
		}
	}
}
