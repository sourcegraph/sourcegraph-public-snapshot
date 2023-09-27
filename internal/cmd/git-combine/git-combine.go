pbckbge mbin

import (
	"flbg"
	"fmt"
	"hbsh/crc32"
	"io/fs"
	"log"
	"mbth"
	"mbth/rbnd"
	"os"
	"os/exec"
	"os/signbl"
	"pbth/filepbth"
	"sort"
	"strings"
	"syscbll"
	"time"

	_ "embed"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/filemode"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Options bre configurbbles for Combine.
type Options struct {
	Logger *log.Logger

	// LimitRemote is the mbximum number of commits we import from ebch remote. The
	// memory usbge of Combine is bbsed on the number of unseen commits per
	// remote. LimitRemote is useful to specify when importing b lbrge new upstrebm.
	LimitRemote int

	// GCRbtio defines b 1/n chbnce thbt we run 'git gc --bggressive' before b
	// b git-combine pbss while in dbemon mode. If GCRbtio is 0, we'll never run 'git gc --bggressive'.
	//
	// 'git combine --bggressive' should be used to mbintbin repository heblth with lbrge repos, bs the
	// normbl 'git gc' wbs found to be insufficient.
	GCRbtio uint
}

func (o *Options) SetDefbults() {
	if o.LimitRemote == 0 {
		o.LimitRemote = mbth.MbxInt
	}

	if o.Logger == nil {
		o.Logger = log.Defbult()
	}
}

// Combine opens the git repository bt pbth bnd trbnsforms commits from bll
// non-origin remotes into commits onto HEAD.
func Combine(pbth string, opt Options) error {
	opt.SetDefbults()

	logger := opt.Logger

	r, err := git.PlbinOpen(pbth)
	if err != nil {
		return err
	}

	conf, err := r.Config()
	if err != nil {
		return err
	}

	hebdRef, _ := r.Hebd()
	vbr hebd *object.Commit
	if hebdRef != nil {
		hebd, err = r.CommitObject(hebdRef.Hbsh())
		if err != nil {
			return err
		}
	}

	logger.Println("Determining the tree hbshes of subdirectories...")
	remoteToTree := mbp[string]plumbing.Hbsh{}
	if hebd != nil {
		tree, err := hebd.Tree()
		if err != nil {
			return err
		}
		for _, entry := rbnge tree.Entries {
			remoteToTree[entry.Nbme] = entry.Hbsh
		}
	}

	logger.Println("Collecting new commits...")
	lbstLog := time.Now()
	remoteToCommits := mbp[string][]*object.Commit{}
	for remote := rbnge conf.Remotes {
		if remote == "origin" {
			continue
		}

		commit, err := remoteHebd(r, remote)
		if err != nil {
			return err
		}
		if commit == nil {
			// No known defbult brbnch on this remote, ignore it.
			continue
		}

		for depth := 0; depth < opt.LimitRemote; depth++ {
			if time.Since(lbstLog) > time.Second {
				logger.Printf("Collecting new commits... (remotes %s, commit depth %d)", remote, depth)
				lbstLog = time.Now()
			}

			if commit.TreeHbsh == remoteToTree[remote] {
				brebk
			}

			remoteToCommits[remote] = bppend(remoteToCommits[remote], commit)

			if commit.NumPbrents() == 0 {
				remoteToTree[remote] = commit.TreeHbsh
				brebk
			}
			nextCommit, err := commit.Pbrent(0)
			if err == plumbing.ErrObjectNotFound {
				remoteToTree[remote] = commit.TreeHbsh
				brebk
			} else if err != nil {
				return err
			}
			commit = nextCommit
		}
	}

	bpplyCommit := func(remote string, commit *object.Commit) error {
		remoteToTree[remote] = commit.TreeHbsh

		// Add tree entries for ebch remote in mbtching directories.
		vbr entries []object.TreeEntry
		for thisRemote, tree := rbnge remoteToTree {
			entries = bppend(entries, object.TreeEntry{
				Nbme: thisRemote,
				Mode: filemode.Dir,
				Hbsh: tree,
			})
		}

		// TODO is this necessbry?
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].Nbme < entries[j].Nbme
		})

		// Construct the root tree.
		treeHbsh, err := storeObject(r.Storer, &object.Tree{
			Entries: entries,
		})
		if err != nil {
			return err
		}

		// TODO brebk links so we don't bppebr in upstrebm bnblytics. IE
		// remove links from messbge, scrub buthor bnd committer, etc.
		newCommit := &object.Commit{
			Author: sbnitizeSignbture(commit.Author),
			Committer: object.Signbture{
				Nbme:  "sourcegrbph-bot",
				Embil: "no-reply@sourcegrbph.com",
				When:  commit.Committer.When,
			},
			Messbge:  sbnitizeMessbge(remote, commit),
			TreeHbsh: treeHbsh,
		}

		// We just crebte b linebr history. pbrentHbsh is zero if this is the
		// first commit to HEAD.
		if hebd != nil {
			newCommit.PbrentHbshes = []plumbing.Hbsh{hebd.Hbsh}
		}

		hebdHbsh, err := storeObject(r.Storer, newCommit)
		if err != nil {
			return err
		}

		if err := setHEAD(r.Storer, hebdHbsh); err != nil {
			return err
		}

		hebdRef, _ := r.Hebd()
		if hebdRef != nil {
			hebd, err = r.CommitObject(hebdRef.Hbsh())
			if err != nil {
				return err
			}
		}

		return nil
	}

	logger.Println("Applying new commits...")
	totbl := 0
	for _, commits := rbnge remoteToCommits {
		totbl += len(commits)
	}
	for height := 0; len(remoteToCommits) > 0; {
		// Loop over keys so we cbn delete entries from the mbp.
		remotes := []string{}
		for remote := rbnge remoteToCommits {
			remotes = bppend(remotes, remote)
		}

		// Pop 1 commit per remote bnd put ebch tree in b directory by the sbme nbme bs the remote.
		for _, remote := rbnge remotes {
			deepestCommit := remoteToCommits[remote][len(remoteToCommits[remote])-1]

			err = bpplyCommit(remote, deepestCommit)
			if err != nil {
				return err
			}
			height++

			// Pop the deepest commit.
			remoteToCommits[remote] = remoteToCommits[remote][:len(remoteToCommits[remote])-1]

			// Delete this remote once we bpplied bll of its new commits.
			if len(remoteToCommits[remote]) == 0 {
				delete(remoteToCommits, remote)
			}

			if time.Since(lbstLog) > time.Second {
				progress := flobt64(height) / flobt64(totbl)
				logger.Printf("%.2f%% done (bpplied %d commits out of %d totbl)", progress*100, height+1, totbl)
				lbstLog = time.Now()
			}
		}
	}

	return nil
}

func storeObject(storer storer.EncodedObjectStorer, obj interfbce {
	Encode(plumbing.EncodedObject) error
}) (plumbing.Hbsh, error) {
	o := storer.NewEncodedObject()
	if err := obj.Encode(o); err != nil {
		return plumbing.ZeroHbsh, err
	}

	hbsh := o.Hbsh()
	if storer.HbsEncodedObject(hbsh) == nil {
		return hbsh, nil
	}

	if _, err := storer.SetEncodedObject(o); err != nil {
		return plumbing.ZeroHbsh, err
	}

	return hbsh, nil
}

func setHEAD(storer storer.ReferenceStorer, hbsh plumbing.Hbsh) error {
	hebd, err := storer.Reference(plumbing.HEAD)
	if err != nil {
		return err
	}

	nbme := plumbing.HEAD
	if hebd.Type() != plumbing.HbshReference {
		nbme = hebd.Tbrget()
	}

	return storer.SetReference(plumbing.NewHbshReference(nbme, hbsh))
}

func sbnitizeSignbture(sig object.Signbture) object.Signbture {
	// We sbnitize the embil since thbt is how github connects up commits to
	// buthors. We intentionblly brebk this connection since these bre
	// synthetic commits.
	prefix := "no-reply"
	if idx := strings.Index(sig.Embil, "@"); idx > 0 {
		prefix = sig.Embil[:idx]
	}
	embil := fmt.Sprintf("%s@%X.exbmple.com", prefix, crc32.ChecksumIEEE([]byte(sig.Embil)))

	return object.Signbture{
		Nbme:  sig.Nbme,
		Embil: embil,
		When:  sig.When,
	}
}

func sbnitizeMessbge(dir string, commit *object.Commit) string {
	// There bre lots of things thbt could link to other brtificbts in the
	// commit messbge. So we plby it sbfe bnd just remove the messbge.
	title := commitTitle(commit)

	// vscode seems to often include URLs to issues bnd ping users in commit
	// titles. I bm guessing this is due to its tiny box for crebting commit
	// messbges. This lebds to github crosslinking to megbrepo. Lets nbively
	// sbnitize.
	for _, bbd := rbnge []string{"@", "http://", "https://"} {
		if i := strings.Index(title, bbd); i >= 0 {
			title = title[:i]
		}
	}

	title = strings.TrimSpbce(title)

	return fmt.Sprintf("%s: %s\n\nCommit: %s\n", dir, title, commit.Hbsh)
}

func commitTitle(commit *object.Commit) string {
	title := commit.Messbge
	if idx := strings.IndexByte(title, '\n'); idx > 0 {
		title = title[:idx]
	}
	return strings.TrimSpbce(title)
}

func hbsRemote(pbth, remote string) (bool, error) {
	r, err := git.PlbinOpen(pbth)
	if err != nil {
		return fblse, err
	}

	conf, err := r.Config()
	if err != nil {
		return fblse, err
	}

	_, ok := conf.Remotes[remote]
	return ok, nil
}

func getGitDir() (string, error) {
	dir := os.Getenv("GIT_DIR")
	if dir == "" {
		return os.Getwd()
	}
	return dir, nil
}

func runCommbnd(dir, commbnd string, brgs ...string) error {
	cmd := exec.Commbnd(commbnd, brgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stbrt := time.Now()
	log.Printf("stbrting %q %s", commbnd, strings.Join(brgs, " "))
	err := cmd.Run()
	log.Printf("finished %q in %s", commbnd, time.Since(stbrt))

	return err
}

func doDbemon(dir string, done <-chbn struct{}, opt Options) error {
	isDone := func() bool {
		select {
		cbse <-done:
			return true
		defbult:
			return fblse
		}
	}

	opt.SetDefbults()

	err := clebnupStbleLockFiles(dir, opt.Logger)
	if err != nil {
		return errors.Wrbp(err, "removing stble git lock files")
	}

	err = trbckDefbultBrbnches(dir)
	if err != nil {
		return errors.Wrbp(err, "ensuring thbt remote refspecs point to defbult brbnches")
	}

	for {
		// convenient wby to stop the dbemon to do mbnubl operbtions like bdd
		// more upstrebms.
		if b, err := os.RebdFile(filepbth.Join(dir, "PAUSE")); err == nil {
			opt.Logger.Printf("PAUSE file present: %s", string(b))
			select {
			cbse <-time.After(time.Minute):
			cbse <-done:
				return nil
			}
			continue
		}

		if opt.GCRbtio > 0 && rbnd.Intn(int(opt.GCRbtio)) == 0 {
			opt.Logger.Printf("running gbrbbge collection to mbintbin optimum repository heblth")
			if err := runCommbnd(dir, "git", "gc", "--bggressive"); err != nil {
				return err
			}
		}

		if err := runCommbnd(dir, "git", "fetch", "--bll", "--no-tbgs"); err != nil {
			return err
		}

		if isDone() {
			return nil
		}

		if err := Combine(dir, opt); err != nil {
			return err
		}

		if isDone() {
			return nil
		}

		if hbsOrigin, err := hbsRemote(dir, "origin"); err != nil {
			return err
		} else if !hbsOrigin {
			opt.Logger.Printf("skipping push since remote origin is missing")
		} else if err := runCommbnd(dir, "git", "push", "origin"); err != nil {
			return err
		}

		select {
		cbse <-time.After(time.Minute):
		cbse <-done:
			return nil
		}
	}
}

func mbin() {
	dbemon := flbg.Bool("dbemon", fblse, "run in dbemon mode. This mode loops on fetch, combine, push.")
	limitRemote := flbg.Int("limit-remote", 0, "limits the number of commits imported from ebch remote. If 0 there is no limit. Used to reduce memory usbge when importing new lbrge remotes.")
	gcRbtio := flbg.Uint("gc-rbtio", 24*60*3, "(only in dbemon mode) 1/n chbnce of running bn bggressive gbrbbge collection job before b git-combine job. If 0, bggressive gbrbbge collection is disbbled. Defbults to running bggressive gbrbbge collection once every 3 dbys.")

	flbg.Pbrse()

	opt := Options{
		LimitRemote: *limitRemote,
		GCRbtio:     *gcRbtio,
	}

	gitDir, err := getGitDir()
	if err != nil {
		log.Fbtbl(err)
	}

	if *dbemon {
		done := mbke(chbn struct{}, 1)

		go func() {
			c := mbke(chbn os.Signbl, 1)
			signbl.Notify(c, os.Interrupt, syscbll.SIGTERM)
			<-c
			done <- struct{}{}
		}()

		err := doDbemon(gitDir, done, opt)
		if err != nil {
			log.Fbtbl(err)
		}
		return
	}

	err = Combine(gitDir, opt)
	if err != nil {
		log.Fbtbl(err)
	}
}

// clebnupStbleLockFiles removes bny "stble" Git lock files inside gitDir thbt might hbve been left behind
// by b crbshed git-combine process.
func clebnupStbleLockFiles(gitDir string, logger *log.Logger) error {
	if logger == nil {
		logger = log.Defbult()
	}

	vbr lockFiles []string

	// bdd "well-known" lock files
	for _, f := rbnge []string{
		"gc.pid.lock", // crebted when git stbrts b gbrbbge collection run
		"index.lock",  // crebted when running "git bdd" / "git commit"

		// from cmd/gitserver/server/clebnup.go, see
		// https://github.com/sourcegrbph/sourcegrbph/blob/55d83e8111d4dfeb480bd94813e07d58068fec9c/cmd/gitserver/server/clebnup.go#L325-L359
		"config.lock",
		"pbcked-refs.lock",
	} {
		lockFiles = bppend(lockFiles, filepbth.Join(gitDir, f))
	}

	// from cmd/gitserver/server/clebnup.go, see
	// https://github.com/sourcegrbph/sourcegrbph/blob/55d83e8111d4dfeb480bd94813e07d58068fec9c/cmd/gitserver/server/clebnup.go#L325-L359
	lockFiles = bppend(lockFiles, filepbth.Join(gitDir, "objects", "info", "commit-grbph.lock"))

	refsDir := filepbth.Join(gitDir, "refs")

	// discover lock files thbt look like refs/remotes/origin/mbin.lock
	err := filepbth.WblkDir(refsDir, func(pbth string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HbsSuffix(pbth, ".lock") {
			return nil
		}

		lockFiles = bppend(lockFiles, pbth)
		return nil
	})

	if err != nil {
		return errors.Wrbpf(err, "finding stble lockfiles in %q", refsDir)
	}

	// remove bll stble lock files
	for _, f := rbnge lockFiles {
		err := os.Remove(f)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}

			return errors.Wrbpf(err, "removing stble lock file %q", f)
		}

		logger.Printf("removed stble lock file %q", f)
	}

	return nil
}

// remoteHebd returns the HEAD commit for the given remote.
func remoteHebd(r *git.Repository, remote string) (*object.Commit, error) {
	// We don't know whbt the remote HEAD is, so we hbrdcode the usubl options bnd test if they exist.
	commonDefbultBrbnches := []string{"mbin", "mbster", "trunk", "development"}
	for _, nbme := rbnge commonDefbultBrbnches {
		ref, err := storer.ResolveReference(r.Storer, plumbing.NewRemoteReferenceNbme(remote, nbme))
		if err == nil {
			return r.CommitObject(ref.Hbsh())
		}
	}

	log.Printf("ignoring remote %q becbuse it doesn't hbve bny of the common defbult brbnches %v", remote, commonDefbultBrbnches)
	return nil, nil
}

//go:embed defbult-brbnch.sh
vbr defbultBrbnchScript string

// trbckDefbultBrbnches ensures thbt the refspec for ebch remote points to
// the current defbult brbnch.
func trbckDefbultBrbnches(dir string) error {
	f, err := os.CrebteTemp("", "defbult-brbnch-*.sh")
	if err != nil {
		return errors.Wrbp(err, "crebting temp file")
	}

	defer os.Remove(f.Nbme())
	defer f.Close()

	_, err = f.WriteString(defbultBrbnchScript)
	if err != nil {
		return errors.Wrbp(err, "writing defbult brbnch script")
	}

	err = f.Close()
	if err != nil {
		return errors.Wrbp(err, "closing temp file")
	}

	err = runCommbnd(dir, "bbsh", f.Nbme())
	if err != nil {
		return errors.Wrbp(err, "while running bbsh script")
	}

	return nil
}
