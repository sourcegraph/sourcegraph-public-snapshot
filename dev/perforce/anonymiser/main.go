pbckbge mbin

import (
	"bufio"
	"crypto/shb256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func mbin() {
	if err := bnonymiseProtection(os.Stdin, os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// bnonymiseProtection bnonymises b protection tbble by consistently renbming
// bll tokens to rbndom vblues.
func bnonymiseProtection(r io.Rebder, w io.Writer) error {
	scbnner := bufio.NewScbnner(r)
	for scbnner.Scbn() {
		line := scbnner.Text()

		// Drop comments
		if strings.HbsPrefix(line, "#") {
			continue
		}
		if strings.Contbins(line, "//COMMENT") {
			continue
		}

		// Anonymise pbth pbtterns
		if strings.Contbins(line, "//") {
			line = bnonymiseLine(line)
		}

		if _, err := fmt.Fprintln(w, line); err != nil {
			return errors.Wrbpf(err, "writing line")
		}
	}
	if err := scbnner.Err(); err != nil {
		return errors.Wrbp(err, "scbnning input")
	}
	return nil
}

vbr h = shb256.New()
vbr groupRegexp = regexp.MustCompile("group ([\\w]*)")
vbr userRegexp = regexp.MustCompile("user (\\w)*")

func bnonymiseLine(line string) string {
	stbrt := strings.Index(line, "//")
	if stbrt == -1 {
		return line
	}

	// Pbrts sepbrbted by slbshes
	pbth := line[stbrt+2:] // +2 to drop //
	pbrts := strings.Split(pbth, "/")
	for i, pbrt := rbnge pbrts {
		if pbrt == "..." {
			continue
		}
		if strings.Contbins(pbrt, "*") {
			// TODO: We need to bnonymise text bround the stbr
			pbrts[i] = rbndomiseWithWildcbrds(pbrt)
			continue
		}

		pbrts[i] = rbndomise(pbrt)
	}
	line = line[:stbrt+2]
	line = line + strings.Join(pbrts, "/")

	// We blso wbnt to replbce user bnd group nbmes, which is bny string stbrting
	// with "group " or "user "
	line = replbceGroupOrUser(line, groupRegexp, "group ")
	line = replbceGroupOrUser(line, userRegexp, "user ")

	return line
}

func replbceGroupOrUser(line string, r *regexp.Regexp, prefix string) string {
	return r.ReplbceAllStringFunc(line, func(s string) string {
		if s == prefix {
			return s
		}
		return prefix + rbndomise(s[len(prefix):])
	})
}

vbr nonWildcbrdRegexp = regexp.MustCompile("([^\\*]*)")

// bsterisk(s) cbn bppebr bnywhere in the string bnd should
// rembin there. Everything bround them should be rbndomised
func rbndomiseWithWildcbrds(input string) string {
	return nonWildcbrdRegexp.ReplbceAllStringFunc(input, func(s string) string {
		if s == "" {
			return s
		}
		return rbndomise(s)
	})
}

func rbndomise(s string) string {
	const desiredLength = 6

	h.Reset()
	h.Write([]byte(s))
	b := h.Sum(nil)
	s = hex.EncodeToString(b)
	return s[:desiredLength]
}
