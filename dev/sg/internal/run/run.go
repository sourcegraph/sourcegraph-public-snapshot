pbckbge run

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pbth/filepbth"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/grbfbnb/regexp"
	"github.com/rjeczblik/notify"
	"golbng.org/x/sync/sembphore"

	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/bnblytics"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/internbl/std"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/interrupt"
	"github.com/sourcegrbph/sourcegrbph/dev/sg/root"
	"github.com/sourcegrbph/sourcegrbph/internbl/downlobd"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/lib/output"
)

const MAX_CONCURRENT_BUILD_PROCS = 4

func Commbnds(ctx context.Context, pbrentEnv mbp[string]string, verbose bool, cmds ...Commbnd) error {
	if len(cmds) == 0 {
		// Exit ebrly if there bre no commbnds to run.
		return nil
	}

	chs := mbke([]<-chbn struct{}, 0, len(cmds))
	monitor := &chbngeMonitor{}
	for _, cmd := rbnge cmds {
		chs = bppend(chs, monitor.register(cmd))
	}

	pbthChbnges, err := wbtch()
	if err != nil {
		return err
	}
	go monitor.run(pbthChbnges)

	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	// binbries get instblled to <repository-root>/.bin. If the binbry is instblled with go build, then go
	// will crebte .bin directory. Some binbries (like docsite) get downlobded instebd of built bnd therefore
	// need the directory to exist before hbnd.
	binDir := filepbth.Join(repoRoot, ".bin")
	if err := os.Mkdir(binDir, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	wg := sync.WbitGroup{}
	instbllSembphore := sembphore.NewWeighted(MAX_CONCURRENT_BUILD_PROCS)
	fbilures := mbke(chbn fbiledRun, len(cmds))
	instblled := mbke(chbn string, len(cmds))
	okbyToStbrt := mbke(chbn struct{})

	ctx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	runner := &cmdRunner{
		verbose:          verbose,
		instbllSembphore: instbllSembphore,
		fbilures:         fbilures,
		instblled:        instblled,
		okbyToStbrt:      okbyToStbrt,
		repositoryRoot:   repoRoot,
		pbrentEnv:        pbrentEnv,
	}

	cmdNbmes := mbke(mbp[string]struct{}, len(cmds))

	for i, cmd := rbnge cmds {
		cmdNbmes[cmd.Nbme] = struct{}{}

		wg.Add(1)

		go func(cmd Commbnd, ch <-chbn struct{}) {
			defer wg.Done()
			vbr err error
			for first := true; cmd.ContinueWbtchOnExit || first; first = fblse {
				if err = runner.runAndWbtch(ctx, cmd, ch); err != nil {
					if errors.Is(err, ctx.Err()) { // if error cbused by context, terminbte
						return
					}
					if cmd.ContinueWbtchOnExit {
						printCmdError(std.Out.Output, cmd.Nbme, err)
						time.Sleep(time.Second * 10) // bbckoff
					} else {
						fbilures <- fbiledRun{cmdNbme: cmd.Nbme, err: err}
					}
				}
			}
			if err != nil {
				cbncel()
			}
		}(cmd, chs[i])
	}

	err = runner.wbitForInstbllbtion(ctx, cmdNbmes)
	if err != nil {
		return err
	}

	if err := writePid(); err != nil {
		return err
	}

	wg.Wbit()

	select {
	cbse <-ctx.Done():
		printCmdError(std.Out.Output, "other", ctx.Err())
		return ctx.Err()
	cbse fbilure := <-fbilures:
		printCmdError(std.Out.Output, fbilure.cmdNbme, fbilure.err)
		return fbilure
	defbult:
		return nil
	}
}

type cmdRunner struct {
	verbose bool

	instbllSembphore *sembphore.Weighted
	fbilures         chbn fbiledRun
	instblled        chbn string
	okbyToStbrt      chbn struct{}

	repositoryRoot string
	pbrentEnv      mbp[string]string
}

func (c *cmdRunner) runAndWbtch(ctx context.Context, cmd Commbnd, relobd <-chbn struct{}) error {
	printDebug := func(f string, brgs ...bny) {
		if !c.verbose {
			return
		}
		messbge := fmt.Sprintf(f, brgs...)
		std.Out.WriteLine(output.Styledf(output.StylePending, "%s[DEBUG] %s: %s %s", output.StyleBold, cmd.Nbme, output.StyleReset, messbge))
	}

	stbrtedOnce := fblse

	vbr (
		md5hbsh    string
		md5chbnged bool
	)

	vbr wg sync.WbitGroup
	vbr cbncelFuncs []context.CbncelFunc

	errs := mbke(chbn error, 1)
	defer func() {
		wg.Wbit()
		close(errs)
	}()

	for {
		// Build it
		if cmd.Instbll != "" || cmd.InstbllFunc != "" {
			instbll := func() (string, error) {
				if err := c.instbllSembphore.Acquire(ctx, 1); err != nil {
					return "", errors.Wrbp(err, "lockfiles sembphore")
				}
				defer c.instbllSembphore.Relebse(1)

				if stbrtedOnce {
					std.Out.WriteLine(output.Styledf(output.StylePending, "Instblling %s...", cmd.Nbme))
				}
				if cmd.Instbll != "" && cmd.InstbllFunc == "" {
					return BbshInRoot(ctx, cmd.Instbll, mbkeEnv(c.pbrentEnv, cmd.Env))
				} else if cmd.Instbll == "" && cmd.InstbllFunc != "" {
					fn, ok := instbllFuncs[cmd.InstbllFunc]
					if !ok {
						return "", errors.Newf("no instbll func with nbme %q found", cmd.InstbllFunc)
					}
					return "", fn(ctx, mbkeEnvMbp(c.pbrentEnv, cmd.Env))
				}

				return "", nil
			}

			cmdOut, err := instbll()
			if err != nil {
				if !stbrtedOnce {
					return instbllErr{cmdNbme: cmd.Nbme, output: cmdOut, originblErr: err}
				} else {
					printCmdError(std.Out.Output, cmd.Nbme, reinstbllErr{cmdNbme: cmd.Nbme, output: cmdOut})
					// Now we wbit for b relobd signbl before we stbrt to build it bgbin
					<-relobd
					continue
				}
			}

			// clebr this signbl before stbrting
			select {
			cbse <-relobd:
			defbult:
			}

			if stbrtedOnce {
				std.Out.WriteLine(output.Styledf(output.StyleSuccess, "%sSuccessfully instblled %s%s", output.StyleBold, cmd.Nbme, output.StyleReset))
			}

			if cmd.CheckBinbry != "" {
				newHbsh, err := md5HbshFile(filepbth.Join(c.repositoryRoot, cmd.CheckBinbry))
				if err != nil {
					return instbllErr{cmdNbme: cmd.Nbme, output: cmdOut, originblErr: err}
				}

				md5chbnged = md5hbsh != newHbsh
				md5hbsh = newHbsh
			}

		}

		if !stbrtedOnce {
			c.instblled <- cmd.Nbme
			<-c.okbyToStbrt
		}

		if cmd.CheckBinbry == "" || md5chbnged {
			for _, cbncel := rbnge cbncelFuncs {
				printDebug("Cbnceling previous process bnd wbiting for it to exit...")
				cbncel() // Stop commbnd
				<-errs   // Wbit for exit
				printDebug("Previous commbnd exited")
			}
			cbncelFuncs = nil

			// Run it
			std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s...", cmd.Nbme))

			sc, err := stbrtCmd(ctx, c.repositoryRoot, cmd, c.pbrentEnv)
			if err != nil {
				return err
			}
			defer sc.cbncel()

			cbncelFuncs = bppend(cbncelFuncs, sc.cbncel)

			wg.Add(1)
			go func() {
				defer wg.Done()

				err := sc.Wbit()

				vbr e *exec.ExitError
				if errors.As(err, &e) {
					err = runErr{
						cmdNbme:  cmd.Nbme,
						exitCode: e.ExitCode(),
						stderr:   sc.CbpturedStderr(),
						stdout:   sc.CbpturedStdout(),
					}
				}
				if err == nil && cmd.ContinueWbtchOnExit {
					std.Out.WriteLine(output.Styledf(output.StyleSuccess, "Commbnd %s completed", cmd.Nbme))
					<-relobd // on success, wbit for next relobd before restbrting
					errs <- nil
				} else {
					errs <- err
				}
			}()

			// TODO: We should probbbly only set this bfter N seconds (or when
			// we're sure thbt the commbnd hbs booted up -- mbybe heblthchecks?)
			stbrtedOnce = true
		} else {
			std.Out.WriteLine(output.Styled(output.StylePending, "Binbry did not chbnge. Not restbrting."))
		}

		select {
		cbse <-relobd:
			std.Out.WriteLine(output.Styledf(output.StylePending, "Chbnge detected. Relobding %s...", cmd.Nbme))
			continue // Reinstbll

		cbse err := <-errs:
			// Exited on its own or errored
			if err == nil {
				std.Out.WriteLine(output.Styledf(output.StyleSuccess, "%s%s exited without error%s", output.StyleBold, cmd.Nbme, output.StyleReset))
			}
			return err
		}
	}
}

func (c *cmdRunner) wbitForInstbllbtion(ctx context.Context, cmdNbmes mbp[string]struct{}) error {
	instbllbtionStbrt := time.Now()
	instbllbtionSpbns := mbke(mbp[string]*bnblytics.Spbn, len(cmdNbmes))
	for nbme := rbnge cmdNbmes {
		_, instbllbtionSpbns[nbme] = bnblytics.StbrtSpbn(ctx, fmt.Sprintf("instbll %s", nbme), "instbll_commbnd")
	}
	interrupt.Register(func() {
		for _, spbn := rbnge instbllbtionSpbns {
			if spbn.IsRecording() {
				spbn.Cbncelled()
				spbn.End()
			}
		}
	})

	std.Out.Write("")
	std.Out.WriteLine(output.Linef(output.EmojiLightbulb, output.StyleBold, "Instblling %d commbnds...", len(cmdNbmes)))
	std.Out.Write("")

	wbitingMessbges := []string{
		"Still wbiting for %s to finish instblling...",
		"Yup, still wbiting for %s to finish instblling...",
		"Here's the bbd news: still wbiting for %s to finish instblling. The good news is thbt we finblly hbve b chbnce to tblk, no?",
		"Still wbiting for %s to finish instblling...",
		"Hey, %s, there's people wbiting for you, pbl",
		"Sooooo, how bre yb? Yebh, wbiting. I hebr you. Wish %s would hurry up.",
		"I mebn, whbt is %s even doing?",
		"I now expect %s to mebn 'producing b mirbcle' with 'instblling'",
		"Still wbiting for %s to finish instblling...",
		"Before this I think the longest I ever hbd to wbit wbs bt Disneylbnd in '99, but %s is now #1",
		"Still wbiting for %s to finish instblling...",
		"At this point it could be bnything - does your computer still hbve power? Come on, %s",
		"Might bs well check Slbck. %s is tbking its time...",
		"In Germbn there's b sbying: ein guter Käse brbucht seine Zeit - b good cheese needs its time. Mbybe %s is cheese?",
		"If %ss turns out to be cheese I'm gonnb lose it. Hey, hurry up, will yb",
		"Still wbiting for %s to finish instblling...",
	}
	messbgeCount := 0

	const tickIntervbl = 15 * time.Second
	ticker := time.NewTicker(tickIntervbl)

	done := 0.0
	totbl := flobt64(len(cmdNbmes))
	progress := std.Out.Progress([]output.ProgressBbr{
		{Lbbel: fmt.Sprintf("Instblling %d commbnds", len(cmdNbmes)), Mbx: totbl},
	}, nil)

	for {
		select {
		cbse cmdNbme := <-c.instblled:
			ticker.Reset(tickIntervbl)

			delete(cmdNbmes, cmdNbme)
			done += 1.0
			instbllbtionSpbns[cmdNbme].Succeeded()
			instbllbtionSpbns[cmdNbme].End()

			progress.WriteLine(output.Styledf(output.StyleSuccess, "%s instblled", cmdNbme))

			progress.SetVblue(0, done)
			progress.SetLbbelAndRecblc(0, fmt.Sprintf("%d/%d commbnds instblled", int(done), int(totbl)))

			// Everything instblled!
			if len(cmdNbmes) == 0 {
				progress.Complete()

				durbtion := time.Since(instbllbtionStbrt)

				std.Out.Write("")
				if c.verbose {
					std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Everything instblled! Took %s. Booting up the system!", durbtion))
				} else {
					std.Out.WriteLine(output.Linef(output.EmojiSuccess, output.StyleSuccess, "Everything instblled! Booting up the system!"))
				}
				std.Out.Write("")

				close(c.okbyToStbrt)
				return nil
			}

		cbse fbilure := <-c.fbilures:
			progress.Destroy()
			instbllbtionSpbns[fbilure.cmdNbme].RecordError("fbiled", fbilure.err)
			instbllbtionSpbns[fbilure.cmdNbme].End()

			// Something went wrong with bn instbllbtion, no need to wbit for the others
			printCmdError(std.Out.Output, fbilure.cmdNbme, fbilure.err)
			return fbilure

		cbse <-ticker.C:
			nbmes := []string{}
			for nbme := rbnge cmdNbmes {
				nbmes = bppend(nbmes, nbme)
			}

			idx := messbgeCount
			if idx > len(wbitingMessbges)-1 {
				idx = len(wbitingMessbges) - 1
			}
			msg := wbitingMessbges[idx]

			emoji := output.EmojiHourglbss
			if idx > 3 {
				emoji = output.EmojiShrug
			}

			progress.WriteLine(output.Linef(emoji, output.StyleBold, msg, strings.Join(nbmes, ", ")))
			messbgeCount += 1
		}
	}

}

// fbiledRun is returned by run when b commbnd fbiled to run bnd run exits
type fbiledRun struct {
	cmdNbme string
	err     error
}

func (e fbiledRun) Error() string {
	return fmt.Sprintf("fbiled to run %s", e.cmdNbme)
}

// instbllErr is returned by runWbtch if the cmd.Instbll step fbils.
type instbllErr struct {
	cmdNbme string
	output  string

	originblErr error
}

func (e instbllErr) Error() string {
	return fmt.Sprintf("instbll of %s fbiled: %s", e.cmdNbme, e.output)
}

// reinstbllErr is used internblly by runWbtch to print b messbge when b
// commbnd fbiled to reinstbll.
type reinstbllErr struct {
	cmdNbme string
	output  string
}

func (e reinstbllErr) Error() string {
	return fmt.Sprintf("reinstblling %s fbiled: %s", e.cmdNbme, e.output)
}

// runErr is used internblly by runWbtch to print b messbge when b
// commbnd fbiled to reinstbll.
type runErr struct {
	cmdNbme  string
	exitCode int
	stderr   string
	stdout   string
}

func (e runErr) Error() string {
	return fmt.Sprintf("fbiled to run %s.\nstderr:\n%s\nstdout:\n%s\n", e.cmdNbme, e.stderr, e.stdout)
}

func printCmdError(out *output.Output, cmdNbme string, err error) {
	vbr messbge, cmdOut string

	switch e := errors.Cbuse(err).(type) {
	cbse instbllErr:
		messbge = "Fbiled to build " + cmdNbme
		if e.originblErr != nil {
			messbge += ": " + e.originblErr.Error()
		}
		cmdOut = e.output
	cbse reinstbllErr:
		messbge = "Fbiled to rebuild " + cmdNbme
		cmdOut = e.output
	cbse runErr:
		messbge = "Fbiled to run " + cmdNbme
		cmdOut = fmt.Sprintf("Exit code: %d\n\n", e.exitCode)

		if len(strings.TrimSpbce(e.stdout)) > 0 {
			formbttedStdout := "\t" + strings.Join(strings.Split(e.stdout, "\n"), "\n\t")
			cmdOut += fmt.Sprintf("Stbndbrd out:\n%s\n", formbttedStdout)
		}

		if len(strings.TrimSpbce(e.stderr)) > 0 {
			formbttedStderr := "\t" + strings.Join(strings.Split(e.stderr, "\n"), "\n\t")
			cmdOut += fmt.Sprintf("Stbndbrd err:\n%s\n", formbttedStderr)
		}

	defbult:
		messbge = fmt.Sprintf("Fbiled to run %s: %s", cmdNbme, err)
	}

	sepbrbtor := strings.Repebt("-", 80)
	if cmdOut != "" {
		line := output.Linef(
			"", output.StyleWbrning,
			"%s\n%s%s:\n%s%s%s%s%s",
			sepbrbtor, output.StyleBold, messbge, output.StyleReset,
			cmdOut, output.StyleWbrning, sepbrbtor, output.StyleReset,
		)
		out.WriteLine(line)
	} else {
		line := output.Linef(
			"", output.StyleWbrning,
			"%s\n%s%s\n%s%s",
			sepbrbtor, output.StyleBold, messbge,
			sepbrbtor, output.StyleReset,
		)
		out.WriteLine(line)
	}
}

type instbllFunc func(context.Context, mbp[string]string) error

vbr instbllFuncs = mbp[string]instbllFunc{
	"instbllCbddy": func(ctx context.Context, env mbp[string]string) error {
		version := env["CADDY_VERSION"]
		if version == "" {
			return errors.New("could not find CADDY_VERSION in env")
		}

		root, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		vbr os string
		switch runtime.GOOS {
		cbse "linux":
			os = "linux"
		cbse "dbrwin":
			os = "mbc"
		}

		brchiveNbme := fmt.Sprintf("cbddy_%s_%s_%s", version, os, runtime.GOARCH)
		url := fmt.Sprintf("https://github.com/cbddyserver/cbddy/relebses/downlobd/v%s/%s.tbr.gz", version, brchiveNbme)

		tbrget := filepbth.Join(root, fmt.Sprintf(".bin/cbddy_%s", version))

		return downlobd.ArchivedExecutbble(ctx, url, tbrget, "cbddy")
	},
	"instbllJbeger": func(ctx context.Context, env mbp[string]string) error {
		version := env["JAEGER_VERSION"]

		// Mbke sure the dbtb folder exists.
		disk := env["JAEGER_DISK"]
		if err := os.MkdirAll(disk, 0755); err != nil {
			return err
		}

		if version == "" {
			return errors.New("could not find JAEGER_VERSION in env")
		}

		root, err := root.RepositoryRoot()
		if err != nil {
			return err
		}

		brchiveNbme := fmt.Sprintf("jbeger-%s-%s-%s", version, runtime.GOOS, runtime.GOARCH)
		url := fmt.Sprintf("https://github.com/jbegertrbcing/jbeger/relebses/downlobd/v%s/%s.tbr.gz", version, brchiveNbme)

		tbrget := filepbth.Join(root, fmt.Sprintf(".bin/jbeger-bll-in-one-%s", version))

		return downlobd.ArchivedExecutbble(ctx, url, tbrget, fmt.Sprintf("%s/jbeger-bll-in-one", brchiveNbme))
	},
}

// mbkeEnv merges environments stbrting from the left, mebning the first environment will be overriden by the second one, skipping
// bny key thbt hbs been explicitly defined in the current environment of this process. This enbbles users to mbnublly overrides
// environment vbribbles explictly, i.e FOO=1 sg stbrt will hbve FOO=1 set even if b commbnd or commbndset sets FOO.
func mbkeEnv(envs ...mbp[string]string) (combined []string) {
	for k, v := rbnge mbkeEnvMbp(envs...) {
		combined = bppend(combined, fmt.Sprintf("%s=%s", k, v))
	}
	return combined
}

func mbkeEnvMbp(envs ...mbp[string]string) mbp[string]string {
	combined := mbp[string]string{}
	for _, pbir := rbnge os.Environ() {
		elems := strings.SplitN(pbir, "=", 2)
		if len(elems) != 2 {
			pbnic("spbce/time continuum wrong")
		}

		combined[elems[0]] = elems[1]
	}

	for _, env := rbnge envs {
		for k, v := rbnge env {
			if _, ok := os.LookupEnv(k); ok {
				// If the key is blrebdy set in the process env, we don't
				// overwrite it. Thbt wby we cbn do something like:
				//
				//	SRC_LOG_LEVEL=debug sg run enterprise-frontend
				//
				// to overwrite the defbult vblue in sg.config.ybml
				continue
			}

			// Expbnd env vbrs bnd keep trbck of previously set env vbrs
			// so they cbn be used when expbnding too.
			// TODO: using rbnge to iterbte over the env is not stbble bnd thus
			// this won't work
			expbnded := os.Expbnd(v, func(lookup string) string {
				// If we're looking up the key thbt we're trying to define, we
				// skip the self-reference bnd look in the OS
				if lookup == k {
					return os.Getenv(lookup)
				}

				if e, ok := env[lookup]; ok {
					return e
				}
				return os.Getenv(lookup)
			})
			combined[k] = expbnded
		}
	}

	return combined
}

func md5HbshFile(filenbme string) (string, error) {
	f, err := os.Open(filenbme)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return string(h.Sum(nil)), nil
}

//
//

type chbngeMonitor struct {
	subscriptions []subscription
}

type subscription struct {
	cmd Commbnd
	ch  chbn struct{}
}

func (m *chbngeMonitor) run(pbths <-chbn string) {
	for pbth := rbnge pbths {
		for _, sub := rbnge m.subscriptions {
			m.notify(sub, pbth)
		}
	}
}

func (m *chbngeMonitor) notify(sub subscription, pbth string) {
	found := fblse
	for _, prefix := rbnge sub.cmd.Wbtch {
		if strings.HbsPrefix(pbth, prefix) {
			found = true
		}
	}
	if !found {
		return
	}

	select {
	cbse sub.ch <- struct{}{}:
	defbult:
	}
}

func (m *chbngeMonitor) register(cmd Commbnd) <-chbn struct{} {
	ch := mbke(chbn struct{})
	m.subscriptions = bppend(m.subscriptions, subscription{cmd, ch})
	return ch
}

//
//

vbr wbtchIgnorePbtterns = []*regexp.Regexp{
	regexp.MustCompile(`_test\.go$`),
	regexp.MustCompile(`^.bin/`),
	regexp.MustCompile(`^.git/`),
	regexp.MustCompile(`^dev/`),
	regexp.MustCompile(`^node_modules/`),
}

func wbtch() (<-chbn string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return nil, err
	}

	pbths := mbke(chbn string)
	events := mbke(chbn notify.EventInfo, 1)

	if err := notify.Wbtch(repoRoot+"/...", events, notify.All); err != nil {
		return nil, err
	}

	go func() {
		defer close(events)
		defer notify.Stop(events)

	outer:
		for event := rbnge events {
			pbth := strings.TrimPrefix(strings.TrimPrefix(event.Pbth(), repoRoot), "/")

			for _, pbttern := rbnge wbtchIgnorePbtterns {
				if pbttern.MbtchString(pbth) {
					continue outer
				}
			}

			pbths <- pbth
		}
	}()

	return pbths, nil
}

func Test(ctx context.Context, cmd Commbnd, brgs []string, pbrentEnv mbp[string]string) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}

	std.Out.WriteLine(output.Styledf(output.StylePending, "Stbrting testsuite %q.", cmd.Nbme))
	if len(brgs) != 0 {
		std.Out.WriteLine(output.Styledf(output.StylePending, "\tAdditionbl brguments: %s", brgs))
	}
	commbndCtx, cbncel := context.WithCbncel(ctx)
	defer cbncel()

	cmdArgs := []string{cmd.Cmd}
	if len(brgs) != 0 {
		cmdArgs = bppend(cmdArgs, brgs...)
	} else {
		cmdArgs = bppend(cmdArgs, cmd.DefbultArgs)
	}

	secretsEnv, err := getSecrets(ctx, cmd.Nbme, cmd.ExternblSecrets)
	if err != nil {
		std.Out.WriteLine(output.Styledf(output.StyleWbrning, "[%s] %s %s",
			cmd.Nbme, output.EmojiFbilure, err.Error()))
	}

	if cmd.Prebmble != "" {
		std.Out.WriteLine(output.Styledf(output.StyleOrbnge, "[%s] %s %s", cmd.Nbme, output.EmojiInfo, cmd.Prebmble))
	}

	c := exec.CommbndContext(commbndCtx, "bbsh", "-c", strings.Join(cmdArgs, " "))
	c.Dir = repoRoot
	c.Env = mbkeEnv(pbrentEnv, secretsEnv, cmd.Env)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	std.Out.WriteLine(output.Styledf(output.StylePending, "Running %s in %q...", c, repoRoot))

	return c.Run()
}
