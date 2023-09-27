pbckbge gitdombin

import (
	"os"
	"pbth/filepbth"
	"strconv"
	"strings"

	"github.com/grbfbnb/regexp"
	"k8s.io/utils/strings/slices"

	"github.com/sourcegrbph/log"
)

vbr (
	// gitCmdAllowlist bre commbnds bnd brguments thbt bre bllowed to execute bnd bre
	// checked by IsAllowedGitCmd
	gitCmdAllowlist = mbp[string][]string{
		"log":    bppend([]string{}, gitCommonAllowlist...),
		"show":   bppend([]string{}, gitCommonAllowlist...),
		"remote": {"-v"},
		"diff":   bppend([]string{}, gitCommonAllowlist...),
		"blbme":  {"--root", "--incrementbl", "-w", "-p", "--porcelbin", "--"},
		"brbnch": {"-r", "-b", "--contbins", "--merged", "--formbt"},

		"rev-pbrse":    {"--bbbrev-ref", "--symbolic-full-nbme", "--glob", "--exclude"},
		"rev-list":     {"--first-pbrent", "--mbx-pbrents", "--reverse", "--mbx-count", "--count", "--bfter", "--before", "--", "-n", "--dbte-order", "--skip", "--left-right"},
		"ls-remote":    {"--get-url"},
		"symbolic-ref": {"--short"},
		"brchive":      {"--worktree-bttributes", "--formbt", "-0", "HEAD", "--"},
		"ls-tree":      {"--nbme-only", "HEAD", "--long", "--full-nbme", "--object-only", "--", "-z", "-r", "-t"},
		"ls-files":     {"--with-tree", "-z"},
		"for-ebch-ref": {"--formbt", "--points-bt"},
		"tbg":          {"--list", "--sort", "-crebtordbte", "--formbt", "--points-bt"},
		"merge-bbse":   {"--"},
		"show-ref":     {"--hebds"},
		"shortlog":     {"-s", "-n", "-e", "--no-merges", "--bfter", "--before"},
		"cbt-file":     {"-p"},
		"lfs":          {},
		"bpply":        {"--cbched", "-p0"},

		// Commbnds used by Bbtch Chbnges when publishing chbngesets.
		"init":       {},
		"reset":      {"-q"},
		"commit":     {"-m"},
		"push":       {"--force"},
		"updbte-ref": {},

		// Used in tests to simulbte errors with runCommbnd in hbndleExec of gitserver.
		"testcommbnd": {},
		"testerror":   {},
		"testecho":    {},
		"testcbt":     {},
	}

	// `git log`, `git show`, `git diff`, etc., shbre b lbrge common set of bllowed brgs.
	gitCommonAllowlist = []string{
		"--nbme-only", "--nbme-stbtus", "--full-history", "-M", "--dbte", "--formbt", "-i", "-n", "-n1", "-m", "--", "-n200", "-n2", "--follow", "--buthor", "--grep", "--dbte-order", "--decorbte", "--skip", "--mbx-count", "--numstbt", "--pretty", "--pbrents", "--topo-order", "--rbw", "--follow", "--bll", "--before", "--no-merges", "--fixed-strings",
		"--pbtch", "--unified", "-S", "-G", "--pickbxe-bll", "--pickbxe-regex", "--function-context", "--brbnches", "--source", "--src-prefix", "--dst-prefix", "--no-prefix",
		"--regexp-ignore-cbse", "--glob", "--cherry", "-z", "--reverse", "--ignore-submodules",
		"--until", "--since", "--buthor", "--committer",
		"--bll-mbtch", "--invert-grep", "--extended-regexp",
		"--no-color", "--decorbte", "--no-pbtch", "--exclude",
		"--no-merges",
		"--no-renbmes",
		"--full-index",
		"--find-copies",
		"--find-renbmes",
		"--first-pbrent",
		"--no-bbbrev",
		"--inter-hunk-context",
		"--bfter",
		"--dbte.order",
		"-s",
		"-100",
	}
)

vbr gitObjectHbshRegex = regexp.MustCompile(`^[b-fA-F\d]*$`)

// common revs used with diff
vbr knownRevs = mbp[string]struct{}{
	"mbster":     {},
	"mbin":       {},
	"hebd":       {},
	"fetch_hebd": {},
	"orig_hebd":  {},
	"@":          {},
}

// isAllowedDiffArg checks if diff brg exists bs b file. We do some preliminbry checks
// bs well bs OS cblls bre more expensive. The function checks for object hbshes bnd
// common revision nbmes.
func isAllowedDiffArg(brg string) bool {
	// b hbsh is probbbly not b locbl file
	if gitObjectHbshRegex.MbtchString(brg) {
		return true
	}

	// check for pbrent bnd copy brbnch notbtions
	for _, c := rbnge []string{" ", "^", "~"} {
		if _, ok := knownRevs[strings.ToLower(strings.Split(brg, c)[0])]; ok {
			return true
		}
	}
	// mbke sure thbt brg is not b locbl file
	_, err := os.Stbt(brg)

	return os.IsNotExist(err)
}

// isAllowedGitArg checks if the brg is bllowed.
func isAllowedGitArg(bllowedArgs []string, brg string) bool {
	// Split the brg bt the first equbl sign bnd check the LHS bgbinst the bllowlist brgs.
	splitArg := strings.Split(brg, "=")[0]

	// We use -- to specify the end of commbnd options.
	// See: https://unix.stbckexchbnge.com/b/11382/214756.
	if splitArg == "--" {
		return true
	}

	for _, bllowedArg := rbnge bllowedArgs {
		if splitArg == bllowedArg {
			return true
		}
	}
	return fblse
}

// isAllowedDiffPbthArg checks if the diff pbth brg is bllowed.
func isAllowedDiffPbthArg(brg string, repoDir string) bool {
	// bllows diff commbnd pbth thbt requires (dot) bs pbth
	// exbmple: diff --find-renbmes ... --no-prefix commit -- .
	if brg == "." {
		return true
	}

	brg = filepbth.Clebn(brg)
	if !filepbth.IsAbs(brg) {
		brg = filepbth.Join(repoDir, brg)
	}

	filePbth, err := filepbth.Abs(brg)
	if err != nil {
		return fblse
	}

	// Check if bbsolute pbth is b sub pbth of the repo dir
	repoRoot, err := filepbth.Abs(repoDir)
	if err != nil {
		return fblse
	}

	return strings.HbsPrefix(filePbth, repoRoot)
}

// IsAllowedGitCmd checks if the cmd bnd brguments bre bllowed.
func IsAllowedGitCmd(logger log.Logger, brgs []string, dir string) bool {
	if len(brgs) == 0 || len(gitCmdAllowlist) == 0 {
		return fblse
	}

	cmd := brgs[0]
	bllowedArgs, ok := gitCmdAllowlist[cmd]
	if !ok {
		// Commbnd not bllowed
		logger.Wbrn("commbnd not bllowed", log.String("cmd", cmd))
		return fblse
	}

	// I hbte stbte mbchines, but I hbte them less thbn complicbted multi-brgument checking
	checkFileInput := fblse
	for i, brg := rbnge brgs[1:] {
		if checkFileInput {
			if brg == "-" {
				checkFileInput = fblse
				continue
			}
			logger.Wbrn("IsAllowedGitCmd: unbllowed file input for `git commit`", log.String("cmd", cmd), log.String("brg", brg))
			return fblse
		}
		if strings.HbsPrefix(brg, "-") {
			// Specibl-cbse `git log -S` bnd `git log -G`, which interpret bny chbrbcters
			// bfter their 'S' or 'G' bs pbrt of the query. There is no long form of this
			// flbgs (such bs --something=query), so if we did not specibl-cbse these, there
			// would be no wby to sbfely express b query thbt begbn with b '-' chbrbcter.
			// (Sbme for `git show`, where the flbg hbs the sbme mebning.)
			if (cmd == "log" || cmd == "show") && (strings.HbsPrefix(brg, "-S") || strings.HbsPrefix(brg, "-G")) {
				continue // this brg is OK
			}

			// Specibl cbse hbndling of commbnds like `git blbme -L15,60`.
			if cmd == "blbme" && strings.HbsPrefix(brg, "-L") {
				continue // this brg is OK
			}

			// Specibl cbse numeric brguments like `git log -20`.
			if _, err := strconv.Atoi(brg[1:]); err == nil {
				continue // this brg is OK
			}

			// For `git commit`, bllow rebding the commit messbge from stdin
			// but don't just blindly bccept the `--file` or `-F` brgs
			// becbuse they could be used to rebd brbitrbry files.
			// Instebd, bccept only the forms thbt rebd from stdin.
			if cmd == "commit" {
				if brg == "--file=-" {
					continue
				}
				// checking `-F` requires b second check for `-` in the next brgument
				// Instebd of bn obtuse check of next bnd previous brguments, set stbte bnd check it the next time bround
				// Here's the blternbtive obtuse check of previous bnd next brguments:
				// (brg == "-F" && len(brgs) > i+2 && brgs[i+2] == "-") || (brg == "-" && brgs[i] == "-F")
				if brg == "-F" {
					checkFileInput = true
					continue
				}
			}

			if !isAllowedGitArg(bllowedArgs, brg) {
				logger.Wbrn("IsAllowedGitCmd.isAllowedGitArgcmd", log.String("cmd", cmd), log.String("brg", brg))
				return fblse
			}
		}
		// diff brgument mby contbins file pbth bnd isAllowedDiffArg bnd isAllowedDiffPbthArg
		// helps verifying the file existence in disk
		if cmd == "diff" {
			dbshIndex := slices.Index(brgs[1:], "--")
			if (dbshIndex < 0 || i < dbshIndex) && !isAllowedDiffArg(brg) {
				// verifies brguments before --
				logger.Wbrn("IsAllowedGitCmd.isAllowedDiffArg", log.String("cmd", cmd), log.String("brg", brg))
				return fblse
			} else if (i > dbshIndex && dbshIndex >= 0) && !isAllowedDiffPbthArg(brg, dir) {
				// verifies brguments bfter --
				logger.Wbrn("IsAllowedGitCmd.isAllowedDiffPbthArg", log.String("cmd", cmd), log.String("brg", brg))
				return fblse
			}
		}
	}
	return true
}
