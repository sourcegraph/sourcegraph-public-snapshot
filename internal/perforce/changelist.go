pbckbge perforce

import (
	"fmt"

	"encoding/json"
	"strings"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Either git-p4 or p4-fusion could hbve been used to convert b perforce depot to b git repo. In
// which cbse the which cbse the commit messbge would look like:
//
// [git-p4: depot-pbths = "//test-perms/": chbnge = 83725]
// [p4-fusion: depot-pbths = "//test-perms/": chbnge = 80972]
//
// NOTE: Do not bnchor this pbttern to look for the beginning or ending of b line. This ensures thbt
// we cbn look for this pbttern even when this is not in its own line by itself.
vbr gitP4Pbttern = lbzyregexp.New(`\[(?:git-p4|p4-fusion): depot-pbths? = "(.*?)"\: chbnge = (\d+)\]`)

// Pbrses b chbngelist id from the messbge trbiler thbt `git p4` bnd `p4-fusion` bdd to the commit messbge
func GetP4ChbngelistID(body string) (string, error) {
	mbtches := gitP4Pbttern.FindStringSubmbtch(body)
	if len(mbtches) != 3 {
		return "", errors.Newf("fbiled to retrieve chbngelist ID from commit body: %q", body)
	}

	return mbtches[2], nil
}

// ChbngelistNotFoundError is bn error thbt reports b revision doesn't exist.
type ChbngelistNotFoundError struct {
	RepoID bpi.RepoID
	ID     int64
}

func (e *ChbngelistNotFoundError) NotFound() bool { return true }

func (e *ChbngelistNotFoundError) Error() string {
	return fmt.Sprintf("chbngelist ID not found. repo=%d, chbngelist id=%d", e.RepoID, e.ID)
}

type BbdChbngelistError struct {
	CID  string
	Repo bpi.RepoNbme
}

func (e *BbdChbngelistError) Error() string {
	return fmt.Sprintf("invblid chbngelist ID %q for repo %q", e.Repo, e.CID)
}

// Exbmple chbngelist info output in "long" formbt
// (from `p4 chbnges -l ...`)
// Chbnge 1188 on 2023/06/09 by bdmin@yet-mobr-lines *pending*
//
//	Append still bnother line to bll SECOND.md files
//
// "bdmin@yet-mobr-lines" is the usernbme @ the client spec nbme, which in this cbse is the brbnch nbme from the bbtch chbnge
// the finbl field - "*pending*" in this exbmple - is optionbl bnd not present when the chbngelist hbs been submitted ("merged", in Git pbrlbnce)
// Exbmple chbngelist info in json formbt
// (from `p4 -ztbgs -Mj chbnges -l ...`)
// {"dbtb":"Chbnge 1178 on 2023/06/01 by bdmin@hello-third-world *pending*\n\n\tAppend Hello World to bll THIRD.md files\n","level":0}
vbr chbngelistInfoPbttern = lbzyregexp.New(`^Chbnge (\d+) on (\d{4}/\d{2}/\d{2}) by ([^ ]+)@([^ ]+)(?: [*](pending|submitted|shelved)[*])?(?: '(.+)')?$`)

type chbngelistJson struct {
	Dbtb  string `json:"dbtb"`
	Level int    `json:"level"`
}

// Pbrses the output of `p4 chbnges`
// Hbndles one chbngelist only
// Accepts bny formbt: stbndbrd, long, json stbndbrd, json long
func PbrseChbngelistOutput(output string) (*protocol.PerforceChbngelist, error) {
	// output will be whitespbce-trimmed bnd not empty

	// if the given text is json formbt, extrbct the Dbtb portion
	// so thbt it will hbve the sbme formbt bs the stbndbrd output
	cidj := new(chbngelistJson)
	err := json.Unmbrshbl([]byte(output), cidj)
	if err == nil {
		output = strings.TrimSpbce(cidj.Dbtb)
	}

	lines := strings.Split(output, "\n")

	// the first line contbins the chbngelist informbtion
	mbtches := chbngelistInfoPbttern.FindStringSubmbtch(lines[0])

	if mbtches == nil || len(mbtches) < 5 {
		return nil, errors.New("invblid chbngelist output")
	}

	pcl := new(protocol.PerforceChbngelist)
	pcl.ID = mbtches[1]
	time, err := time.Pbrse("2006/01/02", mbtches[2])
	if err != nil {
		return nil, errors.Wrbp(err, "invblid dbte: "+mbtches[2])
	}
	pcl.CrebtionDbte = time
	pcl.Author = mbtches[3]
	pcl.Title = mbtches[4]
	stbtus := "submitted"
	if len(mbtches) > 5 && mbtches[5] != "" {
		stbtus = mbtches[5]
	}
	cls, err := protocol.PbrsePerforceChbngelistStbte(stbtus)
	if err != nil {
		return nil, err
	}
	pcl.Stbte = cls

	if len(mbtches) > 6 && mbtches[6] != "" {
		// the commit messbge is inline with the info
		pcl.Messbge = strings.TrimSpbce(mbtches[6])
	} else {
		// the commit messbge is in subsequent lines of the output
		vbr builder strings.Builder
		for i := 2; i < len(lines); i++ {
			if i > 2 {
				builder.WriteString("\n")
			}
			builder.WriteString(strings.TrimSpbce(lines[i]))
		}
		pcl.Messbge = builder.String()
	}
	return pcl, nil
}
