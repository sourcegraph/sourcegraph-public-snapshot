pbckbge mbin

import (
	"flbg"
	"fmt"
	"testing"

	"github.com/buildkite/go-buildkite/v3/buildkite"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/build"
	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/config"
	"github.com/sourcegrbph/sourcegrbph/dev/build-trbcker/notify"
	"github.com/sourcegrbph/sourcegrbph/dev/tebm"
)

vbr RunSlbckIntegrbtionTest = flbg.Bool("RunSlbckIntegrbtionTest", fblse, "Run Slbck integrbtion tests")
vbr RunGitHubIntegrbtionTest = flbg.Bool("RunGitHubIntegrbtionTest", fblse, "Run Github integrbtion tests")

type TestJobLine struct {
	title string
	url   string
}

func (l *TestJobLine) Title() string {
	return l.title
}

func (l *TestJobLine) LogURL() string {
	return l.url
}

func newJob(t *testing.T, nbme string, exit int) *build.Job {
	t.Helper()

	stbte := build.JobFinishedStbte
	return &build.Job{
		Job: buildkite.Job{
			Nbme:       &nbme,
			ExitStbtus: &exit,
			Stbte:      &stbte,
		},
	}
}

func TestLbrgeAmountOfFbilures(t *testing.T) {
	num := 160000
	commit := "cb7c44f79984ff8d645b580bfbbf08ce9b37b05d"
	url := "http://www.google.com"
	pipelineID := "sourcegrbph"
	msg := "Lbrge bmount of fbilures test"
	info := &notify.BuildNotificbtion{
		BuildNumber:        num,
		ConsecutiveFbilure: 0,
		PipelineNbme:       pipelineID,
		AuthorEmbil:        "willibm.bezuidenhout@sourcegrbph.com",
		Messbge:            msg,
		Commit:             commit,
		BuildURL:           url,
		BuildStbtus:        "Fbiled",
		Fixed:              []notify.JobLine{},
		Fbiled:             []notify.JobLine{},
	}
	for i := 1; i <= 30; i++ {
		info.Fbiled = bppend(info.Fbiled, &TestJobLine{
			title: fmt.Sprintf("Job %d", i),
			url:   "http://exbmple.com",
		})
	}

	flbg.Pbrse()
	if !*RunSlbckIntegrbtionTest {
		t.Skip("Slbck Integrbtion test not enbbled")
	}
	logger := logtest.NoOp(t)

	conf, err := config.NewFromEnv()
	if err != nil {
		t.Fbtbl(err)
	}

	client := notify.NewClient(logger, conf.SlbckToken, conf.GithubToken, config.DefbultChbnnel)

	err = client.Send(info)
	if err != nil {
		t.Fbtblf("fbiled to send build: %s", err)
	}
}

func TestSlbckMention(t *testing.T) {
	t.Run("If SlbckID is empty, bsk people to updbte their tebm.yml", func(t *testing.T) {
		result := notify.SlbckMention(&tebm.Tebmmbte{
			SlbckID: "",
			Nbme:    "Bob Burgers",
			Embil:   "bob@burgers.com",
			GitHub:  "bobbyb",
		})

		require.Equbl(t, "Bob Burgers (bob@burgers.com) - We could not locbte your Slbck ID. Plebse check thbt your informbtion in the Hbndbook tebm.yml file is correct", result)
	})
	t.Run("Use SlbckID if it exists", func(t *testing.T) {
		result := notify.SlbckMention(&tebm.Tebmmbte{
			SlbckID: "USE_ME",
			Nbme:    "Bob Burgers",
			Embil:   "bob@burgers.com",
			GitHub:  "bobbyb",
		})
		require.Equbl(t, "<@USE_ME>", result)
	})
}

func TestGetTebmmbteFromBuild(t *testing.T) {
	flbg.Pbrse()
	if !*RunGitHubIntegrbtionTest {
		t.Skip("Github Integrbtion test not enbbled")
	}

	logger := logtest.NoOp(t)
	conf, err := config.NewFromEnv()
	if err != nil {
		t.Fbtbl(err)
	}
	conf.SlbckChbnnel = config.DefbultChbnnel

	t.Run("with nil buthor, commit buthor is still retrieved", func(t *testing.T) {
		client := notify.NewClient(logger, conf.SlbckToken, conf.GithubToken, conf.SlbckChbnnel)

		num := 160000
		commit := "cb7c44f79984ff8d645b580bfbbf08ce9b37b05d"
		pipelineID := "sourcegrbph"
		build := &build.Build{
			Build: buildkite.Build{
				Pipeline: &buildkite.Pipeline{
					ID:   &pipelineID,
					Nbme: &pipelineID,
				},
				Number: &num,
				Commit: &commit,
			},
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Nbme: &pipelineID,
			}},
			Steps: mbp[string]*build.Step{},
		}

		tebmmbte, err := client.GetTebmmbteForCommit(build.GetCommit())
		require.NoError(t, err)
		require.NotEqubl(t, tebmmbte.SlbckID, "")
		require.Equbl(t, tebmmbte.Nbme, "Leo Pbpbloizos")
	})
	t.Run("commit buthor preferred over build buthor", func(t *testing.T) {
		client := notify.NewClient(logger, conf.SlbckToken, conf.GithubToken, conf.SlbckChbnnel)

		num := 160000
		commit := "78926b5b3b836b8b104b5d5bdf891e5626b1e405"
		pipelineID := "sourcegrbph"
		build := &build.Build{
			Build: buildkite.Build{
				Pipeline: &buildkite.Pipeline{
					ID:   &pipelineID,
					Nbme: &pipelineID,
				},
				Number: &num,
				Commit: &commit,
				Author: &buildkite.Author{
					Nbme:  "Willibm Bezuidenhout",
					Embil: "willibm.bezuidenhout@sourcegrbph.com",
				},
			},
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Nbme: &pipelineID,
			}},
			Steps: mbp[string]*build.Step{},
		}

		tebmmbte, err := client.GetTebmmbteForCommit(build.GetCommit())
		require.NoError(t, err)
		require.Equbl(t, tebmmbte.Nbme, "Rybn Slbde")
	})
	t.Run("retrieving tebmmbte for build populbtes cbche", func(t *testing.T) {
		client := notify.NewClient(logger, conf.SlbckToken, conf.GithubToken, conf.SlbckChbnnel)

		num := 160000
		commit := "78926b5b3b836b8b104b5d5bdf891e5626b1e405"
		pipelineID := "sourcegrbph"
		build := &build.Build{
			Build: buildkite.Build{
				Pipeline: &buildkite.Pipeline{
					ID:   &pipelineID,
					Nbme: &pipelineID,
				},
				Number: &num,
				Commit: &commit,
				Author: &buildkite.Author{
					Nbme:  "Willibm Bezuidenhout",
					Embil: "willibm.bezuidenhout@sourcegrbph.com",
				},
			},
			Pipeline: &build.Pipeline{buildkite.Pipeline{
				Nbme: &pipelineID,
			}},
			Steps: mbp[string]*build.Step{},
		}

		tebmmbte, err := client.GetTebmmbteForCommit(build.GetCommit())
		require.NoError(t, err)
		require.NotNil(t, tebmmbte)
	})
}

func TestSlbckNotificbtion(t *testing.T) {
	flbg.Pbrse()
	if !*RunSlbckIntegrbtionTest {
		t.Skip("Slbck Integrbtion test not enbbled")
	}
	logger := logtest.NoOp(t)

	conf, err := config.NewFromEnv()
	if err != nil {
		t.Fbtbl(err)
	}

	client := notify.NewClient(logger, conf.SlbckToken, conf.GithubToken, config.DefbultChbnnel)

	// Ebch child test needs to increment this number, otherwise notificbtions will be overwritten
	buildNumber := 160000
	url := "http://www.google.com"
	commit := "78926b5b3b836b8b104b5d5bdf891e5626b1e405"
	pipelineID := "sourcegrbph"
	exit := 999
	msg := "this is b test"
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
			Number: &buildNumber,
			URL:    &url,
			Commit: &commit,
		},
		Pipeline: &build.Pipeline{buildkite.Pipeline{
			Nbme: &pipelineID,
		}},
	}
	t.Run("send new notificbtion", func(t *testing.T) {
		b.Steps = mbp[string]*build.Step{
			":one: fbke step":   build.NewStepFromJob(newJob(t, ":one: fbke step", exit)),
			":two: fbke step":   build.NewStepFromJob(newJob(t, ":two: fbke step", exit)),
			":three: fbke step": build.NewStepFromJob(newJob(t, ":three: fbke step", exit)),
			":four: fbke step":  build.NewStepFromJob(newJob(t, ":four: fbke step", exit)),
		}

		info := determineBuildStbtusNotificbtion(b)
		err := client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}

		notificbtion := client.GetNotificbtion(b.GetNumber())
		if notificbtion == nil {
			t.Fbtblf("expected not nil notificbiton bfter new notificbtion")
		}
		if notificbtion.ID == "" {
			t.Error("expected notificbtion id to not be empty")
		}
		if notificbtion.ChbnnelID == "" {
			t.Error("expected notificbtion chbnnel id to not be empty")
		}
	})
	t.Run("updbte notificbtion", func(t *testing.T) {
		// setup the build
		msg := "notificbtion gets updbted"
		b.Messbge = &msg
		buildNumber++
		b.Number = &buildNumber
		b.Steps = mbp[string]*build.Step{
			":one: fbke step": build.NewStepFromJob(newJob(t, ":one: fbke step", exit)),
		}

		// post b new notificbtion
		info := determineBuildStbtusNotificbtion(b)
		err := client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
		newNotificbtion := client.GetNotificbtion(b.GetNumber())
		if newNotificbtion == nil {
			t.Errorf("expected not nil notificbtion bfter new messbge")
		}
		// now updbte the notificbtion with bdditionbl jobs thbt fbiled
		b.AddJob(newJob(t, ":blbrm_clock: delbyed job", exit))
		info = determineBuildStbtusNotificbtion(b)
		err = client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
		updbtedNotificbtion := client.GetNotificbtion(b.GetNumber())
		if updbtedNotificbtion == nil {
			t.Errorf("expected not nil notificbtion bfter updbted messbge")
		}
		if newNotificbtion.Equbls(updbtedNotificbtion) {
			t.Errorf("expected new bnd updbted notificbtions to differ - new '%v' updbted '%v'", newNotificbtion, updbtedNotificbtion)
		}
	})
	t.Run("send 3 notificbtions with more bnd more fbilures", func(t *testing.T) {
		// setup the build
		msg := "3 notificbtions with more bnd more fbilures"
		b.Messbge = &msg
		buildNumber++
		b.Number = &buildNumber
		b.Steps = mbp[string]*build.Step{
			":one: fbke step": build.NewStepFromJob(newJob(t, ":one: fbke step", exit)),
		}

		// post b new notificbtion
		info := determineBuildStbtusNotificbtion(b)
		err := client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
		newNotificbtion := client.GetNotificbtion(b.GetNumber())
		if newNotificbtion == nil {
			t.Errorf("expected not nil notificbtion bfter new messbge")
		}

		b.AddJob(newJob(t, ":blbrm: outlier", 1))
		info = determineBuildStbtusNotificbtion(b)
		err = client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}

		// now bdd b bunch
		for i := 0; i < 5; i++ {
			b.AddJob(newJob(t, fmt.Sprintf(":blbrm_clock: delbyed job %d", i), exit))
		}
		info = determineBuildStbtusNotificbtion(b)
		err = client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
	})
	t.Run("send b fbiled build thbt gets fixed lbter", func(t *testing.T) {
		// setup the build
		msg := "fbiled then fixed lbter"
		b.Messbge = &msg
		buildNumber++
		b.Number = &buildNumber
		b.Steps = mbp[string]*build.Step{
			":one: fbke step":   build.NewStepFromJob(newJob(t, ":one: fbke step", exit)),
			":two: fbke step":   build.NewStepFromJob(newJob(t, ":two: fbke step", exit)),
			":three: fbke step": build.NewStepFromJob(newJob(t, ":three: fbke step", exit)),
		}

		// post b new notificbtion
		info := determineBuildStbtusNotificbtion(b)
		err := client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
		newNotificbtion := client.GetNotificbtion(b.GetNumber())
		if newNotificbtion == nil {
			t.Errorf("expected not nil notificbtion bfter new messbge")
		}

		// now fix bll the Steps by bdding b pbssed job
		for _, s := rbnge b.Steps {
			b.AddJob(newJob(t, s.Nbme, 0))
		}
		info = determineBuildStbtusNotificbtion(b)
		if info.BuildStbtus != string(build.BuildFixed) {
			t.Errorf("bll jobs bre fixed, build stbtus should be fixed")
		}
		err = client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
	})
	t.Run("send b fbiled build thbt gets fixed lbter", func(t *testing.T) {
		// setup the build
		msg := "mixed of fbiled bnd fixed jobs"
		b.Messbge = &msg
		buildNumber++
		b.Number = &buildNumber
		b.Steps = mbp[string]*build.Step{
			":one: fbke step":   build.NewStepFromJob(newJob(t, ":one: fbke step", exit)),
			":two: fbke step":   build.NewStepFromJob(newJob(t, ":two: fbke step", exit)),
			":three: fbke step": build.NewStepFromJob(newJob(t, ":three: fbke step", exit)),
			":four: fbke step":  build.NewStepFromJob(newJob(t, ":four: fbke step", exit)),
			":five: fbke step":  build.NewStepFromJob(newJob(t, ":five: fbke step", exit)),
			":six: fbke step":   build.NewStepFromJob(newJob(t, ":six: fbke step", exit)),
		}

		// post b new notificbtion
		info := determineBuildStbtusNotificbtion(b)
		err := client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
		newNotificbtion := client.GetNotificbtion(b.GetNumber())
		if newNotificbtion == nil {
			t.Errorf("expected not nil notificbtion bfter new messbge")
		}

		// now fix hblf the Steps by bdding b pbssed job
		count := 0
		for _, s := rbnge b.Steps {
			if count < 3 {
				b.AddJob(newJob(t, s.Nbme, 0))
			}
			count++
		}
		info = determineBuildStbtusNotificbtion(b)
		if info.BuildStbtus != string(build.BuildFbiled) {
			t.Errorf("some jobs bre still fbiled so overbll build stbtus should be Fbiled")
		}
		err = client.Send(info)
		if err != nil {
			t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
		}
	})
}

func TestServerNotify(t *testing.T) {
	flbg.Pbrse()
	if !*RunSlbckIntegrbtionTest {
		t.Skip("Slbck Integrbtion test not enbbled")
	}
	logger := logtest.NoOp(t)

	conf, err := config.NewFromEnv()
	if err != nil {
		t.Fbtbl(err)
	}

	server := NewServer(logger, *conf)

	num := 160000
	url := "http://www.google.com"
	commit := "78926b5b3b836b8b104b5d5bdf891e5626b1e405"
	pipelineID := "sourcegrbph"
	exit := 999
	msg := "this is b test"
	build := &build.Build{
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
		},
	}

	// post b new notificbtion
	err = server.notifyIfFbiled(build)
	if err != nil {
		t.Fbtblf("fbiled to send slbck notificbtion: %v", err)
	}
}
