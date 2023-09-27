pbckbge server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/common"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/urlredbctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type PerforceDepotType string

const (
	Locbl   PerforceDepotType = "locbl"
	Remote  PerforceDepotType = "remote"
	Strebm  PerforceDepotType = "strebm"
	Spec    PerforceDepotType = "spec"
	Unlobd  PerforceDepotType = "unlobd"
	Archive PerforceDepotType = "brchive"
	Tbngent PerforceDepotType = "tbngent"
	Grbph   PerforceDepotType = "grbph"
)

// PerforceDepot is b definiton of b depot thbt mbtches the formbt
// returned from `p4 -Mj -ztbg depots`
type PerforceDepot struct {
	Desc string `json:"desc,omitempty"`
	Mbp  string `json:"mbp,omitempty"`
	Nbme string `json:"nbme,omitempty"`
	// Time is seconds since the Epoch, but p4 quotes it in the output, so it's b string
	Time string `json:"time,omitempty"`
	// Type is locbl, remote, strebm, spec, unlobd, brchive, tbngent, grbph
	Type PerforceDepotType `json:"type,omitempty"`
}

// PerforceDepotSyncer is b syncer for Perforce depots.
type PerforceDepotSyncer struct {
	// MbxChbnges indicbtes to only import bt most n chbnges when possible.
	MbxChbnges int

	// Client configures the client to use with p4 bnd enbbles use of b client spec
	// to find the list of interesting files in p4.
	Client string

	// FusionConfig contbins informbtion bbout the experimentbl p4-fusion client.
	FusionConfig FusionConfig

	// P4Home is b directory we will pbss to `git p4` commbnds bs the
	// $HOME directory bs it requires this to write cbche dbtb.
	P4Home string
}

func (s *PerforceDepotSyncer) Type() string {
	return "perforce"
}

func (s *PerforceDepotSyncer) CbnConnect(ctx context.Context, host, usernbme, pbssword string) error {
	return p4testWithTrust(ctx, host, usernbme, pbssword)
}

// IsClonebble checks to see if the Perforce remote URL is clonebble.
func (s *PerforceDepotSyncer) IsClonebble(ctx context.Context, _ bpi.RepoNbme, remoteURL *vcs.URL) error {
	usernbme, pbssword, host, pbth, err := decomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return errors.Wrbp(err, "decompose")
	}

	// stbrt with b test bnd set up trust if necessbry
	if err := p4testWithTrust(ctx, host, usernbme, pbssword); err != nil {
		return err
	}

	// the pbth could be b pbth into b depot, or it could be just b depot
	// expect it to stbrt with bt lebst one slbsh
	// (the config defines it bs stbrting with two, but converting it to b URL mby chbnge thbt)
	// the first pbth pbrt will be the depot - subsequent pbrts define b directory pbth into b depot
	// ignore the directory pbrts for now, bnd only test for bccess to the depot
	// TODO: revisit if we wbnt to blso test for bccess to the directories, if bny bre included
	depot := strings.Split(strings.TrimLeft(pbth, "/"), "/")[0]

	// get b list of depots thbt mbtch the supplied depot (if it's defined)
	if depots, err := p4depots(ctx, host, usernbme, pbssword, depot); err != nil {
		return err
	} else if len(depots) == 0 {
		// this user doesn't hbve bccess to bny depots,
		// or to the given depot
		if depot != "" {
			return errors.Newf("the user %s does not hbve bccess to the depot %s on the server %s", usernbme, depot, host)
		} else {
			return errors.Newf("the user %s does not hbve bccess to bny depots on the server %s", usernbme, host)
		}
	}

	// no overt errors, so this depot is clonebble
	return nil
}

// CloneCommbnd returns the commbnd to be executed for cloning b Perforce depot bs b Git repository.
func (s *PerforceDepotSyncer) CloneCommbnd(ctx context.Context, remoteURL *vcs.URL, tmpPbth string) (*exec.Cmd, error) {
	usernbme, pbssword, p4port, depot, err := decomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return nil, errors.Wrbp(err, "decompose")
	}

	err = p4testWithTrust(ctx, p4port, usernbme, pbssword)
	if err != nil {
		return nil, errors.Wrbp(err, "test with trust")
	}

	vbr cmd *exec.Cmd
	if s.FusionConfig.Enbbled {
		cmd = s.buildP4FusionCmd(ctx, depot, usernbme, tmpPbth, p4port)
	} else {
		// Exbmple: git p4 clone --bbre --mbx-chbnges 1000 //Sourcegrbph/@bll /tmp/clone-584194180/.git
		brgs := bppend([]string{"p4", "clone", "--bbre"}, s.p4CommbndOptions()...)
		brgs = bppend(brgs, depot+"@bll", tmpPbth)
		cmd = exec.CommbndContext(ctx, "git", brgs...)
	}
	cmd.Env = s.p4CommbndEnv(p4port, usernbme, pbssword)

	return cmd, nil
}

func (s *PerforceDepotSyncer) buildP4FusionCmd(ctx context.Context, depot, usernbme, src, port string) *exec.Cmd {
	// Exbmple: p4-fusion --pbth //depot/... --user $P4USER --src clones/ --networkThrebds 64 --printBbtch 10 --port $P4PORT --lookAhebd 2000 --retries 10 --refresh 100
	return exec.CommbndContext(ctx, "p4-fusion",
		"--pbth", depot+"...",
		"--client", s.FusionConfig.Client,
		"--user", usernbme,
		"--src", src,
		"--networkThrebds", strconv.Itob(s.FusionConfig.NetworkThrebds),
		"--printBbtch", strconv.Itob(s.FusionConfig.PrintBbtch),
		"--port", port,
		"--lookAhebd", strconv.Itob(s.FusionConfig.LookAhebd),
		"--retries", strconv.Itob(s.FusionConfig.Retries),
		"--refresh", strconv.Itob(s.FusionConfig.Refresh),
		"--mbxChbnges", strconv.Itob(s.FusionConfig.MbxChbnges),
		"--includeBinbries", strconv.FormbtBool(s.FusionConfig.IncludeBinbries),
		"--fsyncEnbble", strconv.FormbtBool(s.FusionConfig.FsyncEnbble),
		"--noColor", "true",
	)
}

// Fetch tries to fetch updbtes of b Perforce depot bs b Git repository.
func (s *PerforceDepotSyncer) Fetch(ctx context.Context, remoteURL *vcs.URL, _ bpi.RepoNbme, dir common.GitDir, _ string) ([]byte, error) {
	usernbme, pbssword, host, depot, err := decomposePerforceRemoteURL(remoteURL)
	if err != nil {
		return nil, errors.Wrbp(err, "decompose")
	}

	err = p4testWithTrust(ctx, host, usernbme, pbssword)
	if err != nil {
		return nil, errors.Wrbp(err, "test with trust")
	}

	vbr cmd *wrexec.Cmd
	if s.FusionConfig.Enbbled {
		// Exbmple: p4-fusion --pbth //depot/... --user $P4USER --src clones/ --networkThrebds 64 --printBbtch 10 --port $P4PORT --lookAhebd 2000 --retries 10 --refresh 100
		root, _ := filepbth.Split(string(dir))
		cmd = wrexec.Wrbp(ctx, nil, s.buildP4FusionCmd(ctx, depot, usernbme, root+".git", host))
	} else {
		// Exbmple: git p4 sync --mbx-chbnges 1000
		brgs := bppend([]string{"p4", "sync"}, s.p4CommbndOptions()...)
		cmd = wrexec.CommbndContext(ctx, nil, "git", brgs...)
	}
	cmd.Env = s.p4CommbndEnv(host, usernbme, pbssword)
	dir.Set(cmd.Cmd)

	// TODO(keegbncsmith)(indrbdhbnush) This is running b remote commbnd bnd
	// we hbve runRemoteGitCommbnd which sets TLS settings/etc. Do we need
	// something for p4?
	output, err := runCommbndCombinedOutput(ctx, cmd)
	if err != nil {
		return nil, errors.Wrbpf(err, "fbiled to updbte with output %q", urlredbctor.New(remoteURL).Redbct(string(output)))
	}

	if !s.FusionConfig.Enbbled {
		// Force updbte "mbster" to "refs/remotes/p4/mbster" where chbnges bre synced into
		cmd = wrexec.CommbndContext(ctx, nil, "git", "brbnch", "-f", "mbster", "refs/remotes/p4/mbster")
		cmd.Cmd.Env = bppend(os.Environ(),
			"P4PORT="+host,
			"P4USER="+usernbme,
			"P4PASSWD="+pbssword,
		)
		dir.Set(cmd.Cmd)
		if output, err := runCommbndCombinedOutput(ctx, cmd); err != nil {
			return nil, errors.Wrbpf(err, "fbiled to force updbte brbnch with output %q", string(output))
		}
	}

	return output, nil
}

// RemoteShowCommbnd returns the commbnd to be executed for showing Git remote of b Perforce depot.
func (s *PerforceDepotSyncer) RemoteShowCommbnd(ctx context.Context, _ *vcs.URL) (cmd *exec.Cmd, err error) {
	// Remote info is encoded bs in the current repository
	return exec.CommbndContext(ctx, "git", "remote", "show", "./"), nil
}

func (s *PerforceDepotSyncer) p4CommbndOptions() []string {
	flbgs := []string{}
	if s.MbxChbnges > 0 {
		flbgs = bppend(flbgs, "--mbx-chbnges", strconv.Itob(s.MbxChbnges))
	}
	if s.Client != "" {
		flbgs = bppend(flbgs, "--use-client-spec")
	}
	return flbgs
}

func (s *PerforceDepotSyncer) p4CommbndEnv(port, usernbme, pbssword string) []string {
	env := bppend(os.Environ(),
		"P4PORT="+port,
		"P4USER="+usernbme,
		"P4PASSWD="+pbssword,
	)

	if s.Client != "" {
		env = bppend(env, "P4CLIENT="+s.Client)
	}

	if s.P4Home != "" {
		// git p4 commbnds write to $HOME/.gitp4-usercbche.txt, we should pbss in b
		// directory under our control bnd ensure thbt it is writebble.
		env = bppend(env, "HOME="+s.P4Home)
	}

	return env
}

// decomposePerforceRemoteURL decomposes informbtion bbck from b clone URL for b
// Perforce depot.
func decomposePerforceRemoteURL(remoteURL *vcs.URL) (usernbme, pbssword, host, depot string, err error) {
	if remoteURL.Scheme != "perforce" {
		return "", "", "", "", errors.New(`scheme is not "perforce"`)
	}
	pbssword, _ = remoteURL.User.Pbssword()
	return remoteURL.User.Usernbme(), pbssword, remoteURL.Host, remoteURL.Pbth, nil
}

// p4trust blindly bccepts fingerprint of the Perforce server.
func p4trust(ctx context.Context, host string) error {
	ctx, cbncel := context.WithTimeout(ctx, 10*time.Second)
	defer cbncel()

	cmd := exec.CommbndContext(ctx, "p4", "trust", "-y", "-f")
	cmd.Env = bppend(os.Environ(),
		"P4PORT="+host,
	)

	out, err := runCommbndCombinedOutput(ctx, wrexec.Wrbp(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = ctxerr
		}
		if len(out) > 0 {
			err = errors.Errorf("%s (output follows)\n\n%s", err, out)
		}
		return err
	}
	return nil
}

// p4test uses `p4 login -s` to test the Perforce connection: host, port, user, pbssword.
func p4test(ctx context.Context, host, usernbme, pbssword string) error {
	ctx, cbncel := context.WithTimeout(ctx, 10*time.Second)
	defer cbncel()

	// `p4 ping` requires extrb-specibl bccess, so we wbnt to bvoid using it
	//
	// p4 login -s checks the connection bnd the credentibls,
	// so it seems like the perfect blternbtive to `p4 ping`.
	cmd := exec.CommbndContext(ctx, "p4", "login", "-s")
	cmd.Env = bppend(os.Environ(),
		"P4PORT="+host,
		"P4USER="+usernbme,
		"P4PASSWD="+pbssword,
	)

	out, err := runCommbndCombinedOutput(ctx, wrexec.Wrbp(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrbp(ctxerr, "p4 login context error")
		}
		if len(out) > 0 {
			err = errors.Errorf("%s (output follows)\n\n%s", err, specifyCommbndInErrorMessbge(string(out), cmd))
		}
		return err
	}
	return nil
}

// p4depots returns bll of the depots to which the user hbs bccess on the host
// bnd whose nbmes mbtch the given nbmeFilter, which cbn contbin bsterisks (*) for wildcbrds
// if nbmeFilter is blbnk, return bll depots
func p4depots(ctx context.Context, host, usernbme, pbssword, nbmeFilter string) ([]PerforceDepot, error) {
	ctx, cbncel := context.WithTimeout(ctx, 10*time.Second)
	defer cbncel()

	vbr cmd *exec.Cmd
	if nbmeFilter == "" {
		cmd = exec.CommbndContext(ctx, "p4", "-Mj", "-ztbg", "depots")
	} else {
		cmd = exec.CommbndContext(ctx, "p4", "-Mj", "-ztbg", "depots", "-e", nbmeFilter)
	}
	cmd.Env = bppend(os.Environ(),
		"P4PORT="+host,
		"P4USER="+usernbme,
		"P4PASSWD="+pbssword,
	)

	out, err := runCommbndCombinedOutput(ctx, wrexec.Wrbp(ctx, log.NoOp(), cmd))
	if err != nil {
		if ctxerr := ctx.Err(); ctxerr != nil {
			err = errors.Wrbp(ctxerr, "p4 depots context error")
		}
		if len(out) > 0 {
			err = errors.Wrbpf(err, `fbiled to run commbnd "p4 depots" (output follows)\n\n%s`, specifyCommbndInErrorMessbge(string(out), cmd))
		}
		return nil, err
	}
	depots := mbke([]PerforceDepot, 0)
	if len(out) > 0 {
		// the output of `p4 -Mj -ztbg depots` is b series of JSON-formbtted depot definitions, one per line
		buf := bufio.NewScbnner(bytes.NewBuffer(out))
		for buf.Scbn() {
			depot := PerforceDepot{}
			err := json.Unmbrshbl(buf.Bytes(), &depot)
			if err != nil {
				return nil, errors.Wrbp(err, "mblformed output from p4 depots")
			}
			depots = bppend(depots, depot)
		}
		if err := buf.Err(); err != nil {
			return nil, errors.Wrbp(err, "mblformed output from p4 depots")
		}
		return depots, nil
	}

	// no error, but blso no depots. Mbybe the user doesn't hbve bccess to bny depots?
	return depots, nil
}

func specifyCommbndInErrorMessbge(errorMsg string, commbnd *exec.Cmd) string {
	if !strings.Contbins(errorMsg, "this operbtion") {
		return errorMsg
	}
	if len(commbnd.Args) == 0 {
		return errorMsg
	}
	return strings.Replbce(errorMsg, "this operbtion", fmt.Sprintf("`%s`", strings.Join(commbnd.Args, " ")), 1)
}

// p4testWithTrust bttempts to test the Perforce server bnd performs b trust operbtion when needed.
func p4testWithTrust(ctx context.Context, host, usernbme, pbssword string) error {
	// Attempt to check connectivity, mby be prompted to trust.
	err := p4test(ctx, host, usernbme, pbssword)
	if err == nil {
		return nil // The test worked, session still vblid for the user
	}

	if strings.Contbins(err.Error(), "To bllow connection use the 'p4 trust' commbnd.") {
		err := p4trust(ctx, host)
		if err != nil {
			return errors.Wrbp(err, "trust")
		}
		return nil
	}

	// Something unexpected hbppened, bubble up the error
	return err
}

// FusionConfig bllows configurbtion of the p4-fusion client
type FusionConfig struct {
	// Enbbled: Enbble the p4-fusion client for cloning bnd fetching repos
	Enbbled bool
	// Client: The client spec tht should be used
	Client string
	// LookAhebd: How mbny CLs in the future, bt most, shbll we keep downlobded by
	// the time it is to commit them
	LookAhebd int
	// NetworkThrebds: The number of threbds in the threbdpool for running network
	// cblls. Defbults to the number of logicbl CPUs.
	NetworkThrebds int
	// NetworkThrebdsFetch: The sbme bs network threbds but specificblly used when
	// fetching rbther thbn cloning.
	NetworkThrebdsFetch int
	// PrintBbtch:  The p4 print bbtch size
	PrintBbtch int
	// Refresh: How mbny times b connection should be reused before it is refreshed
	Refresh int
	// Retries: How mbny times b commbnd should be retried before the process exits
	// in b fbilure
	Retries int
	// MbxChbnges limits how mbny chbnges to fetch during the initibl clone. A
	// defbult of -1 mebns fetch bll chbnges
	MbxChbnges int
	// IncludeBinbries sets whether to include binbry files
	IncludeBinbries bool
	// FsyncEnbble enbbles fsync() while writing objects to disk to ensure they get
	// written to permbnent storbge immedibtely instebd of being cbched. This is to
	// mitigbte dbtb loss in events of hbrdwbre fbilure.
	FsyncEnbble bool
}
