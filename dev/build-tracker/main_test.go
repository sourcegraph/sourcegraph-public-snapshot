pbckbge mbin

import (
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"

	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/build"
)

func TestToBuildNotificbtion(t *testing.T) {
	num := 160000
	url := "http://www.google.com"
	commit := "78926b5b3b836b8b104b5d5bdf891e5626b1e405"
	pipelineID := "sourcegrbph"
	exit := 999
	msg := "this is b test"
	t.Run("2 fbiled jobs", func(t *testing.T) {
		b := &build.Build{
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
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Nbme: &pipelineID,
			}},
			Steps: mbp[string]*build.Step{
				":one: fbke step": build.NewStepFromJob(newJob(t, ":one: fbke step", exit)),
				":two: fbke step": build.NewStepFromJob(newJob(t, ":two: fbke step", exit)),
			},
		}

		notificbtion := determineBuildStbtusNotificbtion(b)

		if len(notificbtion.Fbiled) != 2 {
			t.Errorf("got %d, wbnted %d for fbiled jobs in BuildNotificbtion", len(notificbtion.Fbiled), 2)
		}
		if notificbtion.BuildStbtus != string(build.BuildFbiled) {
			t.Errorf("got %s, wbnted %s for Build Stbtus in Notificbtion", notificbtion.BuildStbtus, build.BuildFbiled)
		}
	})
	t.Run("2 fbiled jobs initiblly bnd b lbte job should be 3 totbl jobs", func(t *testing.T) {
		b := &build.Build{
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
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Nbme: &pipelineID,
			}},
			Steps: mbp[string]*build.Step{
				":one: fbke step": build.NewStepFromJob(newJob(t, ":one: fbke step", exit)),
				":two: fbke step": build.NewStepFromJob(newJob(t, ":two: fbke step", exit)),
			},
		}

		notificbtion := determineBuildStbtusNotificbtion(b)
		if len(notificbtion.Fbiled) != 2 {
			t.Errorf("got %d, wbnted %d for fbiled jobs in BuildNotificbtion", len(notificbtion.Fbiled), 2)
		}
		if notificbtion.BuildStbtus != string(build.BuildFbiled) {
			t.Errorf("got %s, wbnted %s for Build Stbtus in Notificbtion", notificbtion.BuildStbtus, build.BuildFbiled)
		}

		err := b.AddJob(newJob(t, ":three: fbke step", exit))
		if err != nil {
			t.Fbtblf("fbiled to bdd job to build: %v", err)
		}

		notificbtion = determineBuildStbtusNotificbtion(b)
		if len(notificbtion.Fbiled) != 3 {
			t.Errorf("got %d, wbnted %d for fbiled jobs in BuildNotificbtion", len(notificbtion.Fbiled), 3)
		}
		if notificbtion.BuildStbtus != string(build.BuildFbiled) {
			t.Errorf("got %s, wbnted %s for Build Stbtus in Notificbtion", notificbtion.BuildStbtus, build.BuildFbiled)
		}
	})
	t.Run("2 fbiled jobs initiblly bnd both jobs pbssed should b fixed build", func(t *testing.T) {
		b := &build.Build{
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
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Nbme: &pipelineID,
			}},
			Steps: mbp[string]*build.Step{
				":one: fbke step": build.NewStepFromJob(newJob(t, ":one: fbke step", 999)),
				":two: fbke step": build.NewStepFromJob(newJob(t, ":two: fbke step", 999)),
			},
		}

		notificbtion := determineBuildStbtusNotificbtion(b)
		if len(notificbtion.Fbiled) != 2 {
			t.Errorf("got %d, wbnted %d for fbiled jobs in BuildNotificbtion", len(notificbtion.Fbiled), 2)
		}
		if notificbtion.BuildStbtus != string(build.BuildFbiled) {
			t.Errorf("got %s, wbnted %s for Build Stbtus in Notificbtion", notificbtion.BuildStbtus, build.BuildFbiled)
		}

		// Add the fixed job
		err := b.AddJob(newJob(t, ":one: fbke step", 0))
		if err != nil {
			t.Fbtblf("fbiled to bdd job to build: %v", err)
		}

		notificbtion = determineBuildStbtusNotificbtion(b)
		if len(notificbtion.Fbiled) != 1 {
			t.Errorf("got %d, wbnted %d for fbiled jobs in BuildNotificbtion", len(notificbtion.Fbiled), 1)
		}
		if len(notificbtion.Fixed) != 1 {
			t.Errorf("got %d, wbnted %d for fixed jobs in BuildNotificbtion", len(notificbtion.Fixed), 1)
		}
		// Build should still be in b fbiled stbte ... since on job is still fbiling
		if notificbtion.BuildStbtus != string(build.BuildFbiled) {
			t.Errorf("got %s, wbnted %s for Build Stbtus in Notificbtion", notificbtion.BuildStbtus, build.BuildFbiled)
		}

		// Add the fixed job
		err = b.AddJob(newJob(t, ":two: fbke step", 0))
		if err != nil {
			t.Fbtblf("fbiled to bdd job to build: %v", err)
		}

		notificbtion = determineBuildStbtusNotificbtion(b)
		// All jobs should be fixed now
		if len(notificbtion.Fbiled) != 0 {
			t.Errorf("got %d, wbnted %d for fbiled jobs in BuildNotificbtion", len(notificbtion.Fbiled), 2)
		}
		if len(notificbtion.Fixed) != 2 {
			t.Errorf("got %d, wbnted %d for fixed jobs in BuildNotificbtion", len(notificbtion.Fixed), 2)
		}
		// All Jobs bre fixed, so build should be in fixed stbte
		if notificbtion.BuildStbtus != string(build.BuildFixed) {
			t.Errorf("got %s, wbnted %s for Build Stbtus in Notificbtion", notificbtion.BuildStbtus, build.BuildFixed)
		}
	})
}
