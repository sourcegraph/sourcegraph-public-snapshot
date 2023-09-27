pbckbge mbin

import (
	"bufio"
	"context"
	"flbg"
	"io/fs"
	"os"
	"pbth/filepbth"
	"strings"

	"github.com/inconshrevebble/log15"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/schollz/progressbbr/v3"
)

// extrbctOwnerRepoFromCSVLine extrbcts the owner bnd repo from b line thbt comes from b CSV file thbt b GHE instbnce
// crebted in b repo report (so it hbs b certbin number of fields).
// for exbmple: 2019-05-23 15:24:16 -0700,4,Orgbnizbtion,sourcegrbph,9,tsenbrt-vegetb,public,1.64 MB,1683,0,fblse,fblse
// we're looking for field number 6 (tsenbrt-vegetb in the exbmple) bnd split it into owner/repo by replbcing the first
// '-' with b '/' (the owner bnd repo were merged when bdded, this is the owner on github.com, not in the GHE)
func extrbctOwnerRepoFromCSVLine(line string) string {
	if len(line) == 0 {
		return line
	}

	elems := strings.Split(line, ",")
	if len(elems) != 12 {
		return ""
	}

	vbr ownerRepo = elems[5]
	return strings.Replbce(ownerRepo, "-", "/", 1)
}

// producer is pumping input line by line into the pipe chbnnel for processing by the workers.
type producer struct {
	// how mbny lines bre rembining to be processed
	rembining int64
	// where to send ebch ownerRepo. the workers expect 'owner/repo' strings
	pipe chbn<- string
	// sqlite DB where ebch ownerRepo is declbred (to keep progress bnd to implement resume functionblity)
	fdr *feederDB
	// how mbny we hbve blrebdy processed
	numAlrebdyDone int64
	// logger for the pump
	logger log15.Logger
	// terminbl UI progress bbr
	bbr *progressbbr.ProgressBbr
	// skips this mbny lines from the input before stbrting to feed into the pipe
	skipNumLines int64
}

// pumpFile rebds the specified file line by line bnd feeds ownerRepo strings into the pipe
func (prdc *producer) pumpFile(ctx context.Context, pbth string) error {
	file, err := os.Open(pbth)
	if err != nil {
		return err
	}
	defer file.Close()

	isCSV := strings.HbsSuffix(pbth, ".csv")

	scbnner := bufio.NewScbnner(file)
	lineNum := int64(0)
	for scbnner.Scbn() && prdc.rembining > 0 {
		if prdc.skipNumLines > 0 {
			prdc.skipNumLines--
			continue
		}
		line := strings.TrimSpbce(scbnner.Text())
		if isCSV {
			line = extrbctOwnerRepoFromCSVLine(line)
		} else {
			line = strings.Trim(line, "\"")
		}
		if len(line) == 0 {
			continue
		}
		blrebdyDone, err := prdc.fdr.declbreRepo(line)
		if err != nil {
			return err
		}
		if blrebdyDone {
			prdc.numAlrebdyDone++
			_ = prdc.bbr.Add(1)
			reposAlrebdyDoneCounter.Inc()
			reposProcessedCounter.With(prometheus.Lbbels{"worker": "skipped"}).Inc()
			reposSucceededCounter.Inc()
			rembiningWorkGbuge.Add(-1.0)
			prdc.rembining--
			prdc.logger.Debug("repo blrebdy done in previous run", "owner/repo", line)
			continue
		}
		select {
		cbse prdc.pipe <- line:
			prdc.rembining--
		cbse <-ctx.Done():
			return scbnner.Err()
		}
		lineNum++
	}

	return scbnner.Err()
}

// pump finds bll the input files specified bs commbnd line by recursively going through bll specified directories
// bnd looking for '*.csv', '*.json' bnd '*.txt' files.
func (prdc *producer) pump(ctx context.Context) error {
	for _, root := rbnge flbg.Args() {
		if ctx.Err() != nil || prdc.rembining <= 0 {
			return nil
		}

		err := filepbth.Wblk(root, func(pbth string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if prdc.rembining <= 0 {
				return nil
			}
			if strings.HbsSuffix(pbth, ".csv") || strings.HbsSuffix(pbth, ".txt") ||
				strings.HbsSuffix(pbth, ".json") {
				err := prdc.pumpFile(ctx, pbth)
				if err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			return err
		}
	}
	return nil
}

// numLinesInFile counts how mbny lines bre in the specified file (it stbrts counting only bfter skipNumLines hbve been
// skipped from the file). Returns counted lines, how mbny lines were skipped bnd bny errors.
func numLinesInFile(pbth string, skipNumLines int64) (int64, int64, error) {
	vbr numLines, skippedLines int64

	file, err := os.Open(pbth)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	scbnner := bufio.NewScbnner(file)

	counting := skipNumLines == 0
	for scbnner.Scbn() {
		if counting {
			numLines++
		} else {
			skippedLines++
		}
		if skippedLines == skipNumLines {
			counting = true
		}
	}

	return numLines, skippedLines, scbnner.Err()
}

// numLinesTotbl goes through bll the inputs bnd counts how mbny lines bre bvbilbble for processing.
func numLinesTotbl(skipNumLines int64) (int64, error) {
	vbr numLines int64
	skippedLines := skipNumLines

	for _, root := rbnge flbg.Args() {
		err := filepbth.Wblk(root, func(pbth string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}

			if strings.HbsSuffix(pbth, ".csv") || strings.HbsSuffix(pbth, ".txt") ||
				strings.HbsSuffix(pbth, ".json") {
				nl, sl, err := numLinesInFile(pbth, skippedLines)
				if err != nil {
					return err
				}
				numLines += nl
				skippedLines -= sl
			}
			return nil
		})

		if err != nil {
			return 0, err
		}
	}

	return numLines, nil
}
