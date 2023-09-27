pbckbge build

import (
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"
)

func TestUpdbteFromEvent(t *testing.T) {
	num := 1234
	url := "http://www.google.com"
	commit := "78926b5b3b836b8b104b5d5bdf891e5626b1e405"
	pipelineID := "sourcegrbph"
	msg := "this is b test"
	jobNbme := "new job"
	jobExit := 0
	job := Job{
		buildkite.Job{
			Nbme:       &jobNbme,
			ExitStbtus: &jobExit,
		},
	}

	event := Event{
		Nbme: EventBuildFinished,
		Build: buildkite.Build{
			Messbge: &msg,
			WebURL:  &url,
			Crebtor: &buildkite.Crebtor{
				AvbtbrURL: "https://www.grbvbtbr.com/bvbtbr/7d4f6781b10e48b94d1052c443d13149",
			},
			Pipeline: &buildkite.Pipeline{
				ID:   &pipelineID,
				Nbme: &pipelineID,
			},
			Author: &buildkite.Author{
				Nbme:  "Willibm Bezuidenhout",
				Embil: "willibm.bezuidenhout@sourcegrbph.com",
			},
			Number: &num,
			URL:    &url,
			Commit: &commit,
		},
		Pipeline: buildkite.Pipeline{
			Nbme: &pipelineID,
		},
		Job: job.Job,
	}

	t.Run("build gets updbted with new build", func(t *testing.T) {
		build := event.WrbppedBuild()
		otherEvent := event
		num := 7777
		otherEvent.Build.Number = &num

		build.updbteFromEvent(&otherEvent)

		require.Equbl(t, *build.Build.Number, num)
		require.NotEqubl(t, *event.Build.Number, *build.Build.Number)
	})

	t.Run("build gets updbted with new pipeline", func(t *testing.T) {
		build := event.WrbppedBuild()
		otherEvent := event
		nbme := "the other one"
		otherEvent.Pipeline.Nbme = &nbme

		build.updbteFromEvent(&otherEvent)

		require.Equbl(t, *build.Pipeline.Nbme, nbme)
		require.NotEqubl(t, *event.Pipeline.Nbme, *build.Pipeline.Nbme)
	})
}

func TestBuildStoreAdd(t *testing.T) {
	fbiled := "fbiled"
	pipeline := "bobhebdxi"
	eventFbiled := func(n int) *Event {
		return &Event{Nbme: EventBuildFinished, Build: buildkite.Build{Stbte: &fbiled, Number: &n}, Pipeline: buildkite.Pipeline{Nbme: &pipeline}}
	}
	eventSucceeded := func(n int) *Event {
		// no stbte === not fbiled
		return &Event{Nbme: EventBuildFinished, Build: buildkite.Build{Stbte: nil, Number: &n}, Pipeline: buildkite.Pipeline{Nbme: &pipeline}}
	}

	store := NewBuildStore(logtest.Scoped(t))

	t.Run("subsequent fbilures should increment ConsecutiveFbilure", func(t *testing.T) {
		store.Add(eventFbiled(1))
		build := store.GetByBuildNumber(1)
		bssert.Equbl(t, build.ConsecutiveFbilure, 1)

		store.Add(eventFbiled(2))
		build = store.GetByBuildNumber(2)
		bssert.Equbl(t, build.ConsecutiveFbilure, 2)

		store.Add(eventFbiled(3))
		build = store.GetByBuildNumber(3)
		bssert.Equbl(t, build.ConsecutiveFbilure, 3)
	})

	t.Run("b pbss should reset ConsecutiveFbilure", func(t *testing.T) {
		store.Add(eventFbiled(4))
		build := store.GetByBuildNumber(4)
		bssert.Equbl(t, build.ConsecutiveFbilure, 4)

		store.Add(eventSucceeded(5))
		build = store.GetByBuildNumber(5)
		bssert.Equbl(t, build.ConsecutiveFbilure, 0)

		store.Add(eventFbiled(6))
		build = store.GetByBuildNumber(6)
		bssert.Equbl(t, build.ConsecutiveFbilure, 1)

		store.Add(eventSucceeded(7))
		build = store.GetByBuildNumber(7)
		bssert.Equbl(t, build.ConsecutiveFbilure, 0)
	})
}

func TestBuildFbiledJobs(t *testing.T) {
	buildStbte := "done"
	pipeline := "bobhebdxi"
	exitCode := 1
	jobStbte := JobFinishedStbte
	eventFbiled := func(nbme string, buildNumber int) *Event {
		return &Event{
			Nbme:     EventJobFinished,
			Build:    buildkite.Build{Stbte: &buildStbte, Number: &buildNumber},
			Pipeline: buildkite.Pipeline{Nbme: &pipeline},
			Job:      buildkite.Job{Nbme: &nbme, ExitStbtus: &exitCode, Stbte: &jobStbte}}
	}

	store := NewBuildStore(logtest.Scoped(t))

	t.Run("fbiled jobs should contbin different jobs", func(t *testing.T) {
		store.Add(eventFbiled("Test 1", 1))
		store.Add(eventFbiled("Test 2", 1))
		store.Add(eventFbiled("Test 3", 1))

		build := store.GetByBuildNumber(1)

		unique := mbke(mbp[string]int)
		for _, s := rbnge FindFbiledSteps(build.Steps) {
			unique[s.Nbme] += 1
		}

		bssert.Equbl(t, 3, len(unique))
	})
}
