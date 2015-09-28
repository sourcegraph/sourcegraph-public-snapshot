package pgsql

import (
	"log"
	"math/rand"
	"time"

	"github.com/sqs/modl"
)

var (
	BuildTimeout          = 90 * time.Minute
	BuildHeartbeatTimeout = 30 * time.Second

	buildReapersStarted = false
)

// StartBuildReapers starts background processes to periodically kill
// timed out builds. Builds can time out for 2 reasons: (1) they
// started more than BuildTimeout ago, or (2) the worker building them
// hasn't updated their HeartbeatAt in BuildHeartbeatTimeout.
func StartBuildReapers(dbh modl.SqlExecutor) {
	if buildReapersStarted {
		panic("build reapers have already been started")
	}
	buildReapersStarted = true

	go func() {
		for {
			time.Sleep(time.Millisecond * time.Duration(3000+rand.Intn(8000)))

			res, err := dbh.Exec(`
UPDATE repo_build SET failure=true, ended_at=current_timestamp
WHERE (NOT failure) AND ended_at IS NULL AND started_at IS NOT NULL AND (started_at < current_timestamp - ($1 * interval '1 second'))`, int(BuildTimeout/time.Second))
			if err != nil {
				log.Println("Error in background queue timeout process: ", err)
			} else if n, _ := res.RowsAffected(); n > 0 {
				log.Printf("Failed %d builds that exceeded the timeout (%s).", n, BuildTimeout)
			}

			res, err = dbh.Exec(`
UPDATE repo_build SET failure=true, killed=true, ended_at=current_timestamp
WHERE (NOT failure) AND queue AND ended_at IS NULL AND started_at IS NOT NULL AND (greatest(started_at, heartbeat_at) < current_timestamp - ($1 * interval '1 second'))`, int(BuildHeartbeatTimeout/time.Second))
			if err != nil {
				log.Println("Error in background queue heartbeat timeout process: ", err)
			} else if n, _ := res.RowsAffected(); n > 0 {
				log.Printf("Failed %d builds that exceeded the heartbeat timeout (%s).", n, BuildHeartbeatTimeout)
			}

			time.Sleep(time.Second * time.Duration(10+rand.Intn(10)))
		}
	}()
}
