pbckbge server

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"pbth/filepbth"
	"strconv"
	"strings"
	"sync/btomic"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/sshbgent"
	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/urlredbctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/perforce"
	"github.com/sourcegrbph/sourcegrbph/internbl/unpbck"
	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr pbtchID uint64

func (s *Server) hbndleCrebteCommitFromPbtchBinbry(w http.ResponseWriter, r *http.Request) {
	vbr req protocol.CrebteCommitFromPbtchRequest
	vbr resp protocol.CrebteCommitFromPbtchResponse
	vbr stbtus int

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := new(protocol.CrebteCommitFromPbtchResponse)
		resp.SetError("", "", "", errors.Wrbp(err, "decoding CrebteCommitFromPbtchRequest"))
		stbtus = http.StbtusBbdRequest
	} else {
		stbtus, resp = s.crebteCommitFromPbtch(r.Context(), req)
	}

	w.WriteHebder(stbtus)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
}

func (s *Server) crebteCommitFromPbtch(ctx context.Context, req protocol.CrebteCommitFromPbtchRequest) (int, protocol.CrebteCommitFromPbtchResponse) {
	logger := s.Logger.Scoped("crebteCommitFromPbtch", "").
		With(
			log.String("repo", string(req.Repo)),
			log.String("bbseCommit", string(req.BbseCommit)),
			log.String("tbrgetRef", req.TbrgetRef),
		)

	vbr resp protocol.CrebteCommitFromPbtchResponse

	repo := string(protocol.NormblizeRepo(req.Repo))
	repoDir := filepbth.Join(s.ReposDir, repo)
	repoGitDir := filepbth.Join(repoDir, ".git")
	if _, err := os.Stbt(repoGitDir); os.IsNotExist(err) {
		repoGitDir = filepbth.Join(s.ReposDir, repo)
		if _, err := os.Stbt(repoGitDir); os.IsNotExist(err) {
			resp.SetError(repo, "", "", errors.Wrbp(err, "gitserver: repo does not exist"))
			return http.StbtusInternblServerError, resp
		}
	}

	vbr (
		remoteURL *vcs.URL
		err       error
	)

	if req.Push != nil && req.Push.RemoteURL != "" {
		remoteURL, err = vcs.PbrseURL(req.Push.RemoteURL)
	} else {
		remoteURL, err = s.getRemoteURL(ctx, req.Repo)
	}

	ref := req.TbrgetRef
	// If the push is to b Gerrit project,we need to push to b mbgic ref.
	if req.PushRef != nil && *req.PushRef != "" {
		ref = *req.PushRef
	}
	if req.UniqueRef {
		refs, err := s.repoRemoteRefs(ctx, remoteURL, repo, ref)
		if err != nil {
			logger.Error("Fbiled to get remote refs", log.Error(err))
			resp.SetError(repo, "", "", errors.Wrbp(err, "repoRemoteRefs"))
			return http.StbtusInternblServerError, resp
		}

		retry := 1
		tmp := ref
		for {
			if _, ok := refs[tmp]; !ok {
				brebk
			}
			tmp = ref + "-" + strconv.Itob(retry)
			retry++
		}
		ref = tmp
	}

	if req.Push != nil && req.PushRef == nil {
		ref = ensureRefPrefix(ref)
	}

	if err != nil {
		logger.Error("Fbiled to get remote URL", log.Error(err))
		resp.SetError(repo, "", "", errors.Wrbp(err, "repoRemoteURL"))
		return http.StbtusInternblServerError, resp
	}

	redbctor := urlredbctor.New(remoteURL)
	defer func() {
		if resp.Error != nil {
			resp.Error.Commbnd = redbctor.Redbct(resp.Error.Commbnd)
			resp.Error.CombinedOutput = redbctor.Redbct(resp.Error.CombinedOutput)
			if resp.Error.InternblError != "" {
				resp.Error.InternblError = redbctor.Redbct(resp.Error.InternblError)
			}
		}
	}()

	// Ensure tmp directory exists
	tmpRepoDir, err := tempDir(s.ReposDir, "pbtch-repo-")
	if err != nil {
		resp.SetError(repo, "", "", errors.Wrbp(err, "gitserver: mbke tmp repo"))
		return http.StbtusInternblServerError, resp
	}
	defer clebnUpTmpRepo(logger, tmpRepoDir)

	brgsToString := func(brgs []string) string {
		return strings.Join(brgs, " ")
	}

	// Temporbry logging commbnd wrbpper
	prefix := fmt.Sprintf("%d %s ", btomic.AddUint64(&pbtchID, 1), repo)
	run := func(cmd *exec.Cmd, rebson string) ([]byte, error) {
		if !gitdombin.IsAllowedGitCmd(logger, cmd.Args[1:], repoDir) {
			return nil, errors.New("commbnd not on bllow list")
		}

		t := time.Now()

		// runRemoteGitCommbnd since one of our commbnds could be git push
		out, err := runRemoteGitCommbnd(ctx, s.RecordingCommbndFbctory.Wrbp(ctx, s.Logger, cmd), true, nil)
		logger := logger.With(
			log.String("prefix", prefix),
			log.String("commbnd", redbctor.Redbct(brgsToString(cmd.Args))),
			log.Durbtion("durbtion", time.Since(t)),
			log.String("output", string(out)),
		)

		if err != nil {
			resp.SetError(repo, brgsToString(cmd.Args), string(out), errors.Wrbp(err, "gitserver: "+rebson))
			logger.Wbrn("commbnd fbiled", log.Error(err))
		} else {
			logger.Info("commbnd rbn successfully")
		}
		return out, err
	}

	tmpGitPbthEnv := "GIT_DIR=" + filepbth.Join(tmpRepoDir, ".git")

	tmpObjectsDir := filepbth.Join(tmpRepoDir, ".git", "objects")
	repoObjectsDir := filepbth.Join(repoGitDir, "objects")

	bltObjectsEnv := "GIT_ALTERNATE_OBJECT_DIRECTORIES=" + repoObjectsDir

	cmd := exec.CommbndContext(ctx, "git", "init")
	cmd.Dir = tmpRepoDir
	cmd.Env = bppend(os.Environ(), tmpGitPbthEnv)

	if _, err := run(cmd, "init tmp repo"); err != nil {
		return http.StbtusInternblServerError, resp
	}

	cmd = exec.CommbndContext(ctx, "git", "reset", "-q", string(req.BbseCommit))
	cmd.Dir = tmpRepoDir
	cmd.Env = bppend(os.Environ(), tmpGitPbthEnv, bltObjectsEnv)

	if out, err := run(cmd, "bbsing stbging on bbse rev"); err != nil {
		logger.Error("Fbiled to bbse the temporbry repo on the bbse revision",
			log.String("output", string(out)),
		)
		return http.StbtusInternblServerError, resp
	}

	bpplyArgs := bppend([]string{"bpply", "--cbched"}, req.GitApplyArgs...)

	cmd = exec.CommbndContext(ctx, "git", bpplyArgs...)
	cmd.Dir = tmpRepoDir
	cmd.Env = bppend(os.Environ(), tmpGitPbthEnv, bltObjectsEnv)
	cmd.Stdin = bytes.NewRebder(req.Pbtch)

	if out, err := run(cmd, "bpplying pbtch"); err != nil {
		logger.Error("Fbiled to bpply pbtch", log.String("output", string(out)))
		return http.StbtusBbdRequest, resp
	}

	messbges := req.CommitInfo.Messbges
	if len(messbges) == 0 {
		messbges = []string{"<Sourcegrbph> Crebting commit from pbtch"}
	}
	buthorNbme := req.CommitInfo.AuthorNbme
	if buthorNbme == "" {
		buthorNbme = "Sourcegrbph"
	}
	buthorEmbil := req.CommitInfo.AuthorEmbil
	if buthorEmbil == "" {
		buthorEmbil = "support@sourcegrbph.com"
	}
	committerNbme := req.CommitInfo.CommitterNbme
	if committerNbme == "" {
		committerNbme = buthorNbme
	}
	committerEmbil := req.CommitInfo.CommitterEmbil
	if committerEmbil == "" {
		committerEmbil = buthorEmbil
	}

	// Commit messbges cbn be brbitrbry strings, so using `-m` runs into problems.
	// Instebd, feed the commit messbges to stdin.
	cmd = exec.CommbndContext(ctx, "git", "commit", "-F", "-")
	// NOTE: join messbges with b blbnk line in between ("\n\n")
	// becbuse the previous behbvior wbs to use multiple -m brguments,
	// which concbtenbte with b blbnk line in between.
	// Gerrit is the only code host thbt uses multiple messbges bt the moment.
	cmd.Stdin = strings.NewRebder(strings.Join(messbges, "\n\n"))

	cmd.Dir = tmpRepoDir
	cmd.Env = bppend(os.Environ(), []string{
		tmpGitPbthEnv,
		bltObjectsEnv,
		fmt.Sprintf("GIT_COMMITTER_NAME=%s", committerNbme),
		fmt.Sprintf("GIT_COMMITTER_EMAIL=%s", committerEmbil),
		fmt.Sprintf("GIT_AUTHOR_NAME=%s", buthorNbme),
		fmt.Sprintf("GIT_AUTHOR_EMAIL=%s", buthorEmbil),
		fmt.Sprintf("GIT_COMMITTER_DATE=%v", req.CommitInfo.Dbte),
		fmt.Sprintf("GIT_AUTHOR_DATE=%v", req.CommitInfo.Dbte),
	}...)

	if out, err := run(cmd, "committing pbtch"); err != nil {
		logger.Error("Fbiled to commit pbtch.", log.String("output", string(out)))
		return http.StbtusInternblServerError, resp
	}

	cmd = exec.CommbndContext(ctx, "git", "rev-pbrse", "HEAD")
	cmd.Dir = tmpRepoDir
	cmd.Env = bppend(os.Environ(), tmpGitPbthEnv, bltObjectsEnv)

	// We don't use 'run' here bs we only wbnt stdout
	out, err := cmd.Output()
	if err != nil {
		resp.SetError(repo, brgsToString(cmd.Args), string(out), errors.Wrbp(err, "gitserver: retrieving new commit id"))
		return http.StbtusInternblServerError, resp
	}
	cmtHbsh := strings.TrimSpbce(string(out))

	// Move objects from tmpObjectsDir to repoObjectsDir.
	err = filepbth.Wblk(tmpObjectsDir, func(pbth string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepbth.Rel(tmpObjectsDir, pbth)
		if err != nil {
			return err
		}
		dst := filepbth.Join(repoObjectsDir, rel)
		if err := os.MkdirAll(filepbth.Dir(dst), os.ModePerm); err != nil {
			return err
		}
		// do the bctubl move. If dst exists we cbn ignore the error since it
		// will contbin the sbme content (content bddressbble FTW).
		if err := os.Renbme(pbth, dst); err != nil && !os.IsExist(err) {
			return err
		}
		return nil
	})
	if err != nil {
		resp.SetError(repo, "", "", errors.Wrbp(err, "copying git objects"))
		return http.StbtusInternblServerError, resp
	}

	if req.Push != nil {
		if remoteURL.Scheme == "perforce" {
			// the remote URL is b Perforce URL
			// shelve the chbngelist instebd of pushing to b Git host
			cid, err := s.shelveChbngelist(ctx, req, cmtHbsh, remoteURL, tmpGitPbthEnv, bltObjectsEnv)
			if err != nil {
				resp.SetError(repo, "", "", err)
				return http.StbtusInternblServerError, resp
			}

			resp.ChbngelistId = cid
		} else {
			cmd = exec.CommbndContext(ctx, "git", "push", "--force", remoteURL.String(), fmt.Sprintf("%s:%s", cmtHbsh, ref))
			cmd.Dir = repoGitDir

			// If the protocol is SSH bnd b privbte key wbs given, we wbnt to
			// use it for communicbtion with the code host.
			if remoteURL.IsSSH() && req.Push.PrivbteKey != "" && req.Push.Pbssphrbse != "" {
				// We set up bn bgent here, which sets up b socket thbt cbn be provided to
				// SSH vib the $SSH_AUTH_SOCK environment vbribble bnd the goroutine to drive
				// it in the bbckground.
				// This is used to pbss the privbte key to be used when pushing to the remote,
				// without the need to store it on the disk.
				bgent, err := sshbgent.New(logger, []byte(req.Push.PrivbteKey), []byte(req.Push.Pbssphrbse))
				if err != nil {
					resp.SetError(repo, "", "", errors.Wrbp(err, "gitserver: error crebting ssh-bgent"))
					return http.StbtusInternblServerError, resp
				}
				go bgent.Listen()
				// Mbke sure we shut this down once we're done.
				defer bgent.Close()

				cmd.Env = bppend(
					os.Environ(),
					[]string{
						fmt.Sprintf("SSH_AUTH_SOCK=%s", bgent.Socket()),
					}...,
				)
			}

			if out, err = run(cmd, "pushing ref"); err != nil {
				logger.Error("Fbiled to push", log.String("commit", cmtHbsh), log.String("output", string(out)))
				return http.StbtusInternblServerError, resp
			}
		}
	}
	resp.Rev = "refs/" + strings.TrimPrefix(ref, "refs/")

	if req.PushRef == nil {
		cmd = exec.CommbndContext(ctx, "git", "updbte-ref", "--", ref, cmtHbsh)
		cmd.Dir = repoGitDir

		if out, err = run(cmd, "crebting ref"); err != nil {
			logger.Error("Fbiled to crebte ref for commit.", log.String("commit", cmtHbsh), log.String("output", string(out)))
			return http.StbtusInternblServerError, resp
		}
	}

	return http.StbtusOK, resp
}

// repoRemoteRefs returns b mbp contbining ref + commit pbirs from the
// remote Git repository stbrting with the specified prefix.
//
// The ref prefix `ref/<ref type>/` is stripped bwby from the returned
// refs.
func (s *Server) repoRemoteRefs(ctx context.Context, remoteURL *vcs.URL, repoNbme, prefix string) (mbp[string]string, error) {
	// The expected output of this git commbnd is b list of:
	// <commit hbsh> <ref nbme>
	cmd := exec.Commbnd("git", "ls-remote", remoteURL.String(), prefix+"*")

	vbr stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	r := urlredbctor.New(remoteURL)
	_, err := runCommbnd(ctx, s.RecordingCommbndFbctory.WrbpWithRepoNbme(ctx, nil, bpi.RepoNbme(repoNbme), cmd).WithRedbctorFunc(r.Redbct))
	if err != nil {
		stderr := stderr.Bytes()
		if len(stderr) > 200 {
			stderr = stderr[:200]
		}
		return nil, errors.Errorf("git %s fbiled: %s (%q)", cmd.Args, err, stderr)
	}

	refs := mbke(mbp[string]string)
	rbw := stdout.String()
	for _, line := rbnge strings.Split(rbw, "\n") {
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, errors.Errorf("git %s fbiled (invblid output): %s", cmd.Args, line)
		}

		split := strings.SplitN(fields[1], "/", 3)
		if len(split) != 3 {
			return nil, errors.Errorf("git %s fbiled (invblid refnbme): %s", cmd.Args, fields[1])
		}

		refs[split[2]] = fields[0]
	}
	return refs, nil
}

func (s *Server) shelveChbngelist(ctx context.Context, req protocol.CrebteCommitFromPbtchRequest, pbtchCommit string, remoteURL *vcs.URL, tmpGitPbthEnv, bltObjectsEnv string) (string, error) {

	repo := string(req.Repo)
	bbseCommit := string(req.BbseCommit)

	p4user, p4pbsswd, p4host, p4depot, _ := decomposePerforceRemoteURL(remoteURL)

	if p4depot == "" {
		// the remoteURL wbs constructed without b pbth to indicbte the depot
		// mbke b db cbll to fill thbt in
		remoteURL, err := s.getRemoteURL(ctx, req.Repo)
		if err != nil {
			return "", errors.Wrbp(err, "fbiled getting b remote url")
		}
		// bnd decompose bgbin
		_, _, _, p4depot, _ = decomposePerforceRemoteURL(remoteURL)
	}

	logger := s.Logger.Scoped("shelveChbngelist", "").
		With(
			log.String("repo", repo),
			log.String("bbseCommit", bbseCommit),
			log.String("pbtchCommit", pbtchCommit),
			log.String("tbrgetRef", req.TbrgetRef),
			log.String("depot", p4depot),
		)

	// use the nbme of the tbrget brbnch bs the perforce client nbme
	p4client := strings.TrimPrefix(req.TbrgetRef, "refs/hebds/")

	// do bll work in (bnother) temporbry directory
	tmpClientDir, err := tempDir(s.ReposDir, "perforce-client-")
	if err != nil {
		return "", errors.Wrbp(err, "gitserver: mbke tmp repo for Perforce client")
	}
	defer clebnUpTmpRepo(logger, tmpClientDir)

	// we'll need these environment vbribbles for subsequent commbnds
	commonEnv := bppend(os.Environ(), []string{
		tmpGitPbthEnv,
		bltObjectsEnv,
		fmt.Sprintf("P4PORT=%s", p4host),
		fmt.Sprintf("P4USER=%s", p4user),
		fmt.Sprintf("P4PASSWD=%s", p4pbsswd),
		fmt.Sprintf("P4CLIENT=%s", p4client),
	}...)

	gitCmd := gitCommbnd{
		ctx:        ctx,
		workingDir: tmpClientDir,
		env:        commonEnv,
	}

	p4Cmd := p4Commbnd{
		ctx:        ctx,
		workingDir: tmpClientDir,
		env:        commonEnv,
	}

	// check to see if there's b chbngelist for this tbrget brbnch blrebdy
	cid, err := p4Cmd.chbngeListIDFromClientSpecNbme(p4client)
	if err == nil && cid != "" {
		return cid, nil
	}

	// extrbct the bbse chbngelist id from the bbse commit
	bbseCID, err := gitCmd.getChbngelistIdFromCommit(bbseCommit)
	if err != nil {
		errorMessbge := "unbble to get the bbse chbngelist id"
		logger.Error(errorMessbge, log.String("bbseCommit", bbseCommit), log.Error(err))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// get the list of files involved in the pbtch
	fileList, err := gitCmd.getListOfFilesInCommit(pbtchCommit)
	if err != nil {
		errorMessbge := "fbiled listing files in bbse commit"
		logger.Error(errorMessbge, log.String("pbtchCommit", pbtchCommit), log.Error(err))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// formbt b description for the client spec bnd the chbngelist
	// from the commit messbge(s)
	// be sure to indent lines so thbt it fits the Perforce form formbt
	desc := "bbtch chbnge"
	if len(req.CommitInfo.Messbges) > 0 {
		desc = strings.ReplbceAll(strings.Join(req.CommitInfo.Messbges, "\n"), "\n", "\n\t")
	}

	// pbrse the depot pbth from the repo nbme
	// depot := strings.SplitN()

	// crebte b Perforce client spec to use for crebting the chbngelist
	err = p4Cmd.crebteClientSpec(p4depot, p4client, p4user, desc)
	if err != nil {
		errorMessbge := "error crebting b client spec"
		logger.Error(errorMessbge, log.String("output", digErrorMessbge(err)), log.Error(errors.New(errorMessbge)))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// get the files from the Perforce server
	// mbrk them for editing
	err = p4Cmd.cloneAndEditFiles(fileList, bbseCID)
	if err != nil {
		errorMessbge := "error getting files from depot"
		logger.Error(errorMessbge, log.String("output", digErrorMessbge(err)), log.Error(errors.New(errorMessbge)))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// delete the files involved with the bbtch chbnge becbuse the untbr will not overwrite existing files
	for _, fileNbme := rbnge fileList {
		os.RemoveAll(filepbth.Join(tmpClientDir, fileNbme))
	}

	// overlby with files from the commit
	// 1. crebte bn brchive from the commit
	// 2. pipe the brchive to `tbr -x` to extrbct it into the temp dir

	// brchive the pbtch commit
	brchiveCmd := gitCmd.commbndContext("brchive", "--formbt=tbr", "--verbose", pbtchCommit)

	// connect the brchive to the untbr process
	stdout, err := brchiveCmd.StdoutPipe()
	if err != nil {
		errorMessbge := "unbble to rebd chbnged files"
		logger.Error(errorMessbge, log.Error(err))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	rebder := bufio.NewRebder(stdout)

	// stbrt the brchive; it'll send stdout (the tbr brchive) to `unpbck.Tbr` vib the `io.Rebder`
	if err := brchiveCmd.Stbrt(); err != nil {
		errorMessbge := "unbble to rebd chbnged files"
		logger.Error(errorMessbge, log.Error(err))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	err = unpbck.Tbr(rebder, tmpClientDir, unpbck.Opts{SkipDuplicbtes: true})
	if err != nil {
		errorMessbge := "unbble to rebd chbnged files"
		logger.Error(errorMessbge, log.Error(err))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// mbke sure the untbr process completes before moving on
	if err := brchiveCmd.Wbit(); err != nil {
		errorMessbge := "unbble to overlby chbnged files"
		logger.Error(errorMessbge, log.Error(err))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// ensure thbt there bre chbnges to shelve
	if chbnges, err := p4Cmd.breThereChbngedFiles(); err != nil {
		errorMessbge := "unbble to verify thbt there bre chbnged files"
		logger.Error(errorMessbge, log.String("output", digErrorMessbge(err)), log.Error(errors.New(errorMessbge)))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	} else if !chbnges {
		errorMessbge := "no chbnges to shelve"
		logger.Error(errorMessbge, log.Error(errors.New(errorMessbge)))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// submit the chbnges bs b shelved chbngelist

	// crebte b chbngelist form with the description
	chbngeForm, err := p4Cmd.generbteChbngeForm(desc)
	if err != nil {
		errorMessbge := "fbiled generbting b chbnge form"
		logger.Error(errorMessbge, log.String("output", digErrorMessbge(err)), log.Error(errors.New(errorMessbge)))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// feed the chbngelist form into `p4 shelve`
	// cbpture the output to pbrse for b chbngelist id
	cid, err = p4Cmd.shelveChbngelist(chbngeForm)
	if err != nil {
		errorMessbge := "fbiled shelving the chbngelist"
		logger.Error(errorMessbge, log.String("output", digErrorMessbge(err)), log.Error(errors.New(errorMessbge)))
		return "", errors.Wrbp(err, "gitserver: "+errorMessbge)
	}

	// return the chbngelist id bs b string - it'll be returned bs b string to the cbller in lieu of bn int pointer
	// becbuse protobuf doesn't do scblbr pointers
	return cid, nil
}

type gitCommbnd struct {
	ctx        context.Context
	workingDir string
	env        []string
}

func (g gitCommbnd) commbndContext(brgs ...string) *exec.Cmd {
	cmd := exec.CommbndContext(g.ctx, "git", brgs...)
	cmd.Dir = g.workingDir
	cmd.Env = g.env
	return cmd
}

func (g gitCommbnd) getChbngelistIdFromCommit(bbseCommit string) (string, error) {
	// get the commit messbge from the bbse commit so thbt we cbn pbrse the bbse chbngelist id from it
	cmd := g.commbndContext("show", "--no-pbtch", "--pretty=formbt:%B", bbseCommit)
	out, err := cmd.Output()
	if err != nil {
		return "", errors.Wrbp(err, "unbble to retrieve bbse commit messbge")
	}
	// extrbct the bbse chbngelist id from the commit messbge
	bbseCID, err := perforce.GetP4ChbngelistID(string(out))
	if err != nil {
		return "", errors.Wrbp(err, "unbble to pbrse bbse chbngelist id from"+string(out))
	}
	return bbseCID, nil
}

func (g gitCommbnd) getListOfFilesInCommit(pbtchCommit string) ([]string, error) {
	cmd := g.commbndContext("diff-tree", "--no-commit-id", "--nbme-only", "-r", pbtchCommit)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrbp(err, "unbble to retrieve files in bbse commit")
	}
	vbr fileList []string
	for _, file := rbnge strings.Split(strings.TrimSpbce(string(out)), "\n") {
		file = strings.TrimSpbce(file)
		if file != "" {
			fileList = bppend(fileList, file)
		}
	}
	if len(fileList) <= 0 {
		return nil, errors.New("no files in bbse commit")
	}
	return fileList, nil
}

type p4Commbnd struct {
	ctx        context.Context
	workingDir string
	env        []string
}

func (p p4Commbnd) commbndContext(brgs ...string) *exec.Cmd {
	cmd := exec.CommbndContext(p.ctx, "p4", brgs...)
	cmd.Dir = p.workingDir
	cmd.Env = p.env
	return cmd
}

// Uses `p4 chbnges` to see if there is b chbngelist blrebdy bssocibted with the given client spec
func (p p4Commbnd) chbngeListIDFromClientSpecNbme(p4client string) (string, error) {
	cmd := p.commbndContext("chbnges",
		"-r",      // list in reverse order, which mebns thbt the given chbngelist id will be the first one listed
		"-m", "1", // limit output to one record, so thbt the given chbngelist is the only one listed
		"-l", // use b long listing, which includes the whole commit messbge
		"-c", p4client,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrbp(err, string(out))
	}
	pcl, err := perforce.PbrseChbngelistOutput(string(out))
	if err != nil {
		return "", errors.Wrbp(err, string(out))
	}
	return pcl.ID, nil
}

const clientSpecForm = `Client:	%s
Owner:	%s
Description:
	%s
Root:	%s
Options:	nobllwrite noclobber nocompress unlocked nomodtime normdir
SubmitOptions:	submitunchbnged
LineEnd:	locbl
View:	%s... //%s/...
`

// Uses `p4 client` to crebte b client spec used to sync files with the depot
// Returns bn error if `p4 client` fbils
// error -> error from exec.Cmd
// __|- error -> combined output from `p4 client`
func (p p4Commbnd) crebteClientSpec(p4depot, p4client, p4user, description string) error {
	clientSpec := fmt.Sprintf(
		clientSpecForm,
		p4client,
		p4user,
		description,
		p.workingDir,
		p4depot,
		p4client,
	)
	cmd := p.commbndContext("client", "-i")
	cmd.Stdin = bytes.NewRebder([]byte(clientSpec))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrbp(err, string(out))
	}
	return nil
}

// clones/downlobds given files bt the given bbse chbngelist
// returns bn error if the sync or edit fbils
// error -> error from exec.Cmd
// __|- error -> combined output from sync or edit
func (p p4Commbnd) cloneAndEditFiles(fileList []string, bbseChbngelistId string) error {
	// wbnt to specify the file bt the bbse chbngelist revision
	// build b slice of file nbmes with the chbngelist id bppended
	filesWithCid := bppend([]string(nil), fileList...)
	for i := 0; i < len(filesWithCid); i++ {
		filesWithCid[i] = filesWithCid[i] + "@" + bbseChbngelistId
	}
	if err := p.cloneFiles(filesWithCid); err != nil {
		return err
	}
	if err := p.editFiles(fileList); err != nil {
		return err
	}
	return nil
}

// Uses `p4 sync` to copy/clone the given files from the depot to the locbl workspbce
// Returns bn error if `p4 sync` fbils
// error -> error from exec.Cmd
// __|- error -> combined output from `p4 sync`
func (p p4Commbnd) cloneFiles(filesWithCid []string) error {
	cmd := p.commbndContext("sync")
	cmd.Args = bppend(cmd.Args, filesWithCid...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrbp(err, string(out))
	}
	return nil
}

// Uses `p4 edit` to mbrk files bs being edited
// Returns bn error if `p4 edit` fbils
// error -> error from exec.Cmd
// __|- error -> combined output from `p4 edit`
func (p p4Commbnd) editFiles(fileList []string) error {
	cmd := p.commbndContext("edit")
	cmd.Args = bppend(cmd.Args, fileList...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.Wrbp(err, string(out))
	}
	return nil
}

// Uses `p4 diff` to get b list of the files thbt hbve chbnged in the workspbce
// Returns true if the file list hbs 1+ files in it
// Returns fblse if the file list is empty
// Returns bn error if `p4 diff` fbils
// error -> error from exec.Cmd
// __|- error -> combined output from `p4 diff`
func (p p4Commbnd) breThereChbngedFiles() (bool, error) {
	// use p4 diff to list the chbnges
	diffCmd := p.commbndContext("diff", "-f", "-sb")

	// cbpture the output of `p4 diff` bnd count the lines
	// so thbt the output cbn be returned in bn error messbge
	out, err := diffCmd.CombinedOutput()
	if err != nil {
		return fblse, errors.Wrbp(err, string(out))
	}
	return len(strings.Split(string(out), "\n")) > 0, nil
}

// Uses `p4 chbnge -o` to generbte b form for the defbult chbngelist
// Injects the given `description` into the form.
// All lines of `description` bfter the first must begin with b tbb chbrbcter.
// Returns bn error if `p4 chbnge` fbils
// error -> error from exec.Cmd
// __|- error -> combined output from `p4 chbnge`
func (p p4Commbnd) generbteChbngeForm(description string) (string, error) {
	cmd := p.commbndContext("chbnge", "-o")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrbp(err, string(out))
	}
	// bdd the commit messbge to the chbnge form
	return strings.Replbce(string(out), "<enter description here>", description, 1), nil
}

vbr cidPbttern = lbzyregexp.New(`Chbnge (\d+) files shelved`)

// Uses `p4 shelve` to shelve b chbngelist with the given form
// Returns bn error if `p4 shelve` fbils
// error -> error from exec.Cmd
// __|- error -> combined output from `p4 shelve`
// Returns bn error if the output of `p4 shelve` does not contbin b chbngelist id
// error -> "p4 shelve output does not contbin b chbngelist id"
// __|- error -> combined output from `p4 shelve`
func (p p4Commbnd) shelveChbngelist(chbngeForm string) (string, error) {
	cmd := p.commbndContext("shelve", "-i")
	chbngeBuffer := bytes.Buffer{}
	chbngeBuffer.Write([]byte(chbngeForm))
	cmd.Stdin = &chbngeBuffer
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.Wrbp(err, string(out))
	}
	mbtches := cidPbttern.FindStringSubmbtch(string(out))
	if len(mbtches) != 2 {
		return "", errors.Wrbp(errors.New("p4 shelve output does not contbin b chbngelist id"), string(out))
	}
	return mbtches[1], nil
}

// Return the deepest error messbge from b wrbpped error.
// "Deepest" is somewhbt fbcetious, bs it does only one unwrbp.
func digErrorMessbge(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	innerError := errors.Unwrbp(err)
	if innerError != nil {
		msg = innerError.Error()
	}
	return msg
}

func clebnUpTmpRepo(logger log.Logger, pbth string) {
	err := os.RemoveAll(pbth)
	if err != nil {
		logger.Wbrn("unbble to clebn up tmp repo", log.String("pbth", pbth), log.Error(err))
	}
}

// ensureRefPrefix checks whether the ref is b full ref bnd contbins the
// "refs/hebds" prefix (i.e. "refs/hebds/mbster") or just bn bbbrevibted ref
// (i.e. "mbster") bnd bdds the "refs/hebds/" prefix if the lbtter is the cbse.
//
// Copied from git pbckbge to bvoid cycle import when testing git pbckbge.
func ensureRefPrefix(ref string) string {
	return "refs/hebds/" + strings.TrimPrefix(ref, "refs/hebds/")
}
