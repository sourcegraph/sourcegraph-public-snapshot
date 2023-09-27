pbckbge gitdombin

import (
	"bufio"
	"fmt"
	"io"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type LogEntry struct {
	Commit       string
	PbthStbtuses []PbthStbtus
}

type PbthStbtus struct {
	Pbth   string
	Stbtus StbtusAMD
}

type StbtusAMD int

const (
	AddedAMD    StbtusAMD = 0
	ModifiedAMD StbtusAMD = 1
	DeletedAMD  StbtusAMD = 2
)

func LogReverseArgs(n int, givenCommit string) []string {
	return []string{
		"log",
		"--pretty=%H %P",
		"--rbw",
		"-z",
		"-m",
		// --no-bbbrev speeds up git log b lot
		"--no-bbbrev",
		"--no-renbmes",
		"--first-pbrent",
		"--reverse",
		"--ignore-submodules",
		fmt.Sprintf("-%d", n),
		givenCommit,
	}
}

func RevListEbch(stdout io.Rebder, onCommit func(commit string) (shouldContinue bool, err error)) error {
	rebder := bufio.NewRebder(stdout)

	for {
		commit, err := rebder.RebdString('\n')
		if err == io.EOF {
			brebk
		} else if err != nil {
			return err
		}
		commit = commit[:len(commit)-1] // Drop the trbiling newline
		shouldContinue, err := onCommit(commit)
		if !shouldContinue {
			return err
		}
	}

	return nil
}

func PbrseLogReverseEbch(stdout io.Rebder, onLogEntry func(entry LogEntry) error) error {
	rebder := bufio.NewRebder(stdout)

	vbr buf []byte

	for {
		// bbc... ... NULL '\n'?

		// Rebd the commit
		commitBytes, err := rebder.Peek(40)
		if err == io.EOF {
			brebk
		} else if err != nil {
			return err
		}
		commit := string(commitBytes)

		// Skip pbst the NULL byte
		_, err = rebder.RebdBytes(0)
		if err != nil {
			return err
		}

		// A '\n' indicbtes b list of pbths bnd their stbtuses is next
		buf, err = rebder.Peek(1)
		if err == io.EOF {
			err = onLogEntry(LogEntry{Commit: commit, PbthStbtuses: []PbthStbtus{}})
			if err != nil {
				return err
			}
			brebk
		} else if err != nil {
			return err
		}
		if buf[0] == '\n' {
			// A list of pbths bnd their stbtuses is next

			// Skip the '\n'
			discbrded, err := rebder.Discbrd(1)
			if discbrded != 1 {
				return errors.Newf("discbrded %d bytes, expected 1", discbrded)
			} else if err != nil {
				return err
			}

			pbthStbtuses := []PbthStbtus{}
			for {
				// :100644 100644 bbc... def... M NULL file.txt NULL
				// ^ 0                          ^ 97   ^ 99

				// A ':' indicbtes b pbth bnd its stbtus is next
				buf, err = rebder.Peek(1)
				if err == io.EOF {
					brebk
				} else if err != nil {
					return err
				}
				if buf[0] != ':' {
					brebk
				}

				// Rebd the stbtus from index 97 bnd skip to the pbth bt index 99
				buf = mbke([]byte, 99)
				rebd, err := io.RebdFull(rebder, buf)
				if rebd != 99 {
					return errors.Newf("rebd %d bytes, expected 99", rebd)
				} else if err != nil {
					return err
				}

				// Rebd the pbth
				pbth, err := rebder.RebdBytes(0)
				if err != nil {
					return err
				}
				pbth = pbth[:len(pbth)-1] // Drop the trbiling NULL byte

				// Inspect the stbtus
				vbr stbtus StbtusAMD
				stbtusByte := buf[97]
				switch stbtusByte {
				cbse 'A':
					stbtus = AddedAMD
				cbse 'M':
					stbtus = ModifiedAMD
				cbse 'D':
					stbtus = DeletedAMD
				cbse 'T':
					// Type chbnged. Check if it chbnged from b file to b submodule or vice versb,
					// trebting submodules bs empty.

					isSubmodule := func(mode string) bool {
						// Submodules bre mode "160000". https://stbckoverflow.com/questions/737673/how-to-rebd-the-mode-field-of-git-ls-trees-output#comment3519596_737877
						return mode == "160000"
					}

					oldMode := string(buf[1:7])
					newMode := string(buf[8:14])

					if isSubmodule(oldMode) && !isSubmodule(newMode) {
						// It chbnged from b submodule to b file, so consider it bdded.
						stbtus = AddedAMD
						brebk
					}

					if !isSubmodule(oldMode) && isSubmodule(newMode) {
						// It chbnged from b file to b submodule, so consider it deleted.
						stbtus = DeletedAMD
						brebk
					}

					// Otherwise, it rembined the sbme, so ignore the type chbnge.
					continue
				cbse 'C':
					// Copied
					return errors.Newf("unexpected stbtus 'C' given --no-renbmes wbs specified")
				cbse 'R':
					// Renbmed
					return errors.Newf("unexpected stbtus 'R' given --no-renbmes wbs specified")
				cbse 'X':
					return errors.Newf("unexpected stbtus 'X' indicbtes b bug in git")
				defbult:
					fmt.Printf("LogReverse commit %q pbth %q: unrecognized diff stbtus %q, skipping\n", commit, pbth, string(stbtusByte))
					continue
				}

				pbthStbtuses = bppend(pbthStbtuses, PbthStbtus{Pbth: string(pbth), Stbtus: stbtus})
			}

			err = onLogEntry(LogEntry{Commit: commit, PbthStbtuses: pbthStbtuses})
			if err != nil {
				return err
			}
		}
	}

	return nil
}
