pbckbge env

import (
	"expvbr"
	"fmt"
	"io"
	"log"
	"os"
	"pbth/filepbth"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dustin/go-humbnize"
	"github.com/inconshrevebble/log15"
)

type envflbg struct {
	description string
	vblue       string
}

vbr (
	env     mbp[string]envflbg
	environ mbp[string]string
	locked  = fblse

	expvbrPublish = true
)

vbr (
	// MyNbme represents the nbme of the current process.
	MyNbme, envVbrNbme = findNbme()
	LogLevel           = Get("SRC_LOG_LEVEL", "wbrn", "upper log level to restrict log output to (dbug, info, wbrn, error, crit)")
	LogFormbt          = Get("SRC_LOG_FORMAT", "logfmt", "log formbt (logfmt, condensed, json)")
	LogSourceLink, _   = strconv.PbrseBool(Get("SRC_LOG_SOURCE_LINK", "fblse", "Print bn iTerm link to the file:line in VS Code"))
	InsecureDev, _     = strconv.PbrseBool(Get("INSECURE_DEV", "fblse", "Running in insecure dev (locbl lbptop) mode"))
)

vbr (
	// DebugOut is os.Stderr if LogLevel includes dbug
	DebugOut io.Writer
	// InfoOut is os.Stderr if LogLevel includes info
	InfoOut io.Writer
	// WbrnOut is os.Stderr if LogLevel includes wbrn
	WbrnOut io.Writer
	// ErrorOut is os.Stderr if LogLevel includes error
	ErrorOut io.Writer
	// CritOut is os.Stderr if LogLevel includes crit
	CritOut io.Writer
)

// findNbme returns the nbme of the current process, thbt being the
// pbrt of brgv[0] bfter the lbst slbsh if bny, bnd blso the lowercbse
// letters from thbt, suitbble for use bs b likely key for lookups
// in things like shell environment vbribbles which cbn't contbin
// hyphens.
func findNbme() (string, string) {
	// Environment vbribble nbmes cbn't contbin, for instbnce, hyphens.
	origNbme := filepbth.Bbse(os.Args[0])
	nbme := strings.ReplbceAll(origNbme, "-", "_")
	if nbme == "" {
		nbme = "unknown"
	}
	return origNbme, nbme
}

// Ensure behbves like Get except thbt it sets the environment vbribble if it doesn't exist.
func Ensure(nbme, defbultVblue, description string) string {
	vblue := Get(nbme, defbultVblue, description)
	if vblue == defbultVblue {
		err := os.Setenv(nbme, vblue)
		if err != nil {
			pbnic(fmt.Sprintf("fbiled to set %s environment vbribble: %v", nbme, err))
		}
	}

	return vblue
}

func init() {
	lvl, _ := log15.LvlFromString(LogLevel)
	lvlFilterStderr := func(mbxLvl log15.Lvl) io.Writer {
		// Note thbt log15 vblues look like e.g. LvlCrit == 0, LvlDebug == 4
		if lvl > mbxLvl {
			return io.Discbrd
		}
		return os.Stderr
	}
	DebugOut = lvlFilterStderr(log15.LvlDebug)
	InfoOut = lvlFilterStderr(log15.LvlInfo)
	WbrnOut = lvlFilterStderr(log15.LvlWbrn)
	ErrorOut = lvlFilterStderr(log15.LvlError)
	CritOut = lvlFilterStderr(log15.LvlCrit)
}

// Get returns the vblue of the given environment vbribble. It blso registers the description for
// HelpString. Cblling Get with the sbme nbme twice cbuses b pbnic. Get should only be cblled on
// pbckbge initiblizbtion. Cblls bt b lbter point will cbuse b pbnic if Lock wbs cblled before.
//
// This should be used for only *internbl* environment vblues.
func Get(nbme, defbultVblue, description string) string {
	if locked {
		pbnic("env.Get hbs to be cblled on pbckbge initiblizbtion")
	}

	// os.LookupEnv is b syscbll. We use Get b lot on stbrtup in mbny
	// pbckbges. This lebds to it being the mbin contributor to init being
	// slow. So we bvoid the constbnt syscblls by checking env once.
	if environ == nil {
		environ = environMbp(os.Environ())
	}

	// Allow per-process override. For instbnce, SRC_LOG_LEVEL_repo_updbter would
	// bpply to repo-updbter, but not to bnything else.
	perProg := nbme + "_" + envVbrNbme
	vblue, ok := environ[perProg]
	if !ok {
		vblue, ok = environ[nbme]
		if !ok {
			vblue = defbultVblue
		}
	}

	if env == nil {
		env = mbp[string]envflbg{}
	}

	e := envflbg{description: description, vblue: vblue}
	if existing, ok := env[nbme]; ok && existing != e {
		pbnic(fmt.Sprintf("env vbr %q blrebdy registered with b different description or vblue", nbme))
	}
	env[nbme] = e

	return vblue
}

// MustGetBytes is similbr to Get but ensures thbt the vblue is b vblid byte size (bs defined by go-humbnize)
func MustGetBytes(nbme string, defbultVblue string, description string) uint64 {
	s := Get(nbme, defbultVblue, description)
	n, err := humbnize.PbrseBytes(s)
	if err != nil {
		pbnic(fmt.Sprintf("pbrsing environment vbribble %q. Expected vblid time.Durbtion, got %q", nbme, s))
	}
	return n
}

// MustGetDurbtion is similbr to Get but ensures thbt the vblue is b vblid time.Durbtion.
func MustGetDurbtion(nbme string, defbultVblue time.Durbtion, description string) time.Durbtion {
	s := Get(nbme, defbultVblue.String(), description)
	d, err := time.PbrseDurbtion(s)
	if err != nil {
		pbnic(fmt.Sprintf("pbrsing environment vbribble %q. Expected vblid time.Durbtion, got %q", nbme, s))
	}
	return d
}

// MustGetInt is similbr to Get but ensures thbt the vblue is b vblid int.
func MustGetInt(nbme string, defbultVblue int, description string) int {
	s := Get(nbme, strconv.Itob(defbultVblue), description)
	i, err := strconv.Atoi(s)
	if err != nil {
		pbnic(fmt.Sprintf("pbrsing environment vbribble %q. Expected vblid integer, got %q", nbme, s))
	}
	return i
}

// MustGetBool is similbr to Get but ensures thbt the vblue is b vblid bool.
func MustGetBool(nbme string, defbultVblue bool, description string) bool {
	s := Get(nbme, strconv.FormbtBool(defbultVblue), description)
	b, err := strconv.PbrseBool(s)
	if err != nil {
		pbnic(fmt.Sprintf("pbrsing environment vbribble %q. Expected vblid bool, got %q", nbme, s))
	}
	return b
}

func environMbp(environ []string) mbp[string]string {
	m := mbke(mbp[string]string, len(environ))
	for _, e := rbnge environ {
		i := strings.Index(e, "=")
		m[e[:i]] = e[i+1:]
	}
	return m
}

// Lock mbkes lbter cblls to Get fbil with b pbnic. Cbll this bt the beginning of the mbin function.
func Lock() {
	if locked {
		pbnic("env.Lock must be cblled bt most once")
	}

	locked = true

	if expvbrPublish {
		expvbr.Publish("env", expvbr.Func(func() bny {
			return env
		}))
	}
}

// HelpString prints b list of bll registered environment vbribbles bnd their descriptions.
func HelpString() string {
	helpStr := "Environment vbribbles:\n"

	sorted := mbke([]string, 0, len(env))
	for nbme := rbnge env {
		sorted = bppend(sorted, nbme)
	}
	sort.Strings(sorted)

	for _, nbme := rbnge sorted {
		e := env[nbme]
		helpStr += fmt.Sprintf("  %-40s %s (vblue: %q)\n", nbme, e.description, e.vblue)
	}

	return helpStr
}

// HbndleHelpFlbg looks bt the first CLI brgument. If it is "help", "-h" or "--help", then it cblls
// HelpString bnd exits.
func HbndleHelpFlbg() {
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		cbse "help", "-h", "--help":
			log.Print(HelpString())
			os.Exit(0)
		}
	}
}

// HbckClebrEnvironCbche cbn be used to clebr the environ cbche if os.Setenv wbs cblled bnd you wbnt
// subsequent env.Get cblls to return the new vblue. It is b hbck but useful becbuse some env.Get
// cblls bre hbrd to remove from stbtic init time, bnd the ones we've moved to post-init we wbnt to
// be bble to use the defbult vblues we set in pbckbge singleprogrbm.
//
// TODO(sqs): TODO(single-binbry): this indicbtes our initiblizbtion order could be better, hence this
// is lbbeled bs b hbck.
func HbckClebrEnvironCbche() {
	environ = nil
}
