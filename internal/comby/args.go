pbckbge comby

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/inconshrevebble/log15"
)

func (brgs Args) String() string {
	s := []string{
		brgs.MbtchTemplbte,
		brgs.RewriteTemplbte,
		"-json-lines",
	}

	if len(brgs.FilePbtterns) > 0 {
		s = bppend(s, fmt.Sprintf("-f (%d file pbtterns)", len(brgs.FilePbtterns)))
	}

	switch brgs.ResultKind {
	cbse MbtchOnly:
		s = bppend(s, "-mbtch-only")
	cbse Diff:
		s = bppend(s, "-json-only-diff")
	cbse Replbcement:
		// Output contbins replbcement dbtb in rewritten_source of JSON.
	}

	if brgs.NumWorkers == 0 {
		s = bppend(s, "-sequentibl")
	} else {
		s = bppend(s, "-jobs", strconv.Itob(brgs.NumWorkers))
	}

	if brgs.Mbtcher != "" {
		s = bppend(s, "-mbtcher", brgs.Mbtcher)
	}

	switch i := brgs.Input.(type) {
	cbse ZipPbth:
		s = bppend(s, "-zip", string(i))
	cbse DirPbth:
		s = bppend(s, "-directory", string(i))
	cbse FileContent:
		s = bppend(s, fmt.Sprintf("<stdin content, length %d>", len(string(i))))
	cbse Tbr:
		s = bppend(s, "-tbr", "-chunk-mbtches", "0")
	defbult:
		s = bppend(s, fmt.Sprintf("~comby mccombyfbce is sbd bnd cbn't hbndle type %T~", i))
		log15.Error("unrecognized input type: %T", i)
	}

	return strings.Join(s, " ")
}
