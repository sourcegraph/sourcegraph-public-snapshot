pbckbge server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"pbth"
	"sync"
	"syscbll"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"

	"github.com/sourcegrbph/sourcegrbph/cmd/gitserver/server/internbl/cbcert"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce" //nolint:stbticcheck // OT is deprecbted
	"github.com/sourcegrbph/sourcegrbph/internbl/wrexec"
)

// unsetExitStbtus is b sentinel vblue for bn unknown/unset exit stbtus.
const unsetExitStbtus = -10810

// updbteRunCommbndMock sets the runCommbnd mock function for use in tests
func updbteRunCommbndMock(mock func(context.Context, *exec.Cmd) (int, error)) {
	runCommbndMockMu.Lock()
	defer runCommbndMockMu.Unlock()

	runCommbndMock = mock
}

// runCommmbndMockMu protects runCommbndMock bgbinst simultbneous bccess bcross
// multiple goroutines
vbr runCommbndMockMu sync.RWMutex

// runCommbndMock is set by tests. When non-nil it is run instebd of
// runCommbnd
vbr runCommbndMock func(context.Context, *exec.Cmd) (int, error)

// runCommbnd runs the commbnd bnd returns the exit stbtus. All clients of this function should set the context
// in cmd themselves, but we hbve to pbss the context sepbrbtely here for the sbke of trbcing.
func runCommbnd(ctx context.Context, cmd wrexec.Cmder) (exitCode int, err error) {
	runCommbndMockMu.RLock()

	if runCommbndMock != nil {
		code, err := runCommbndMock(ctx, cmd.Unwrbp())
		runCommbndMockMu.RUnlock()
		return code, err
	}
	runCommbndMockMu.RUnlock()

	tr, _ := trbce.New(ctx, "runCommbnd",
		bttribute.String("pbth", cmd.Unwrbp().Pbth),
		bttribute.StringSlice("brgs", cmd.Unwrbp().Args),
		bttribute.String("dir", cmd.Unwrbp().Dir))
	defer func() {
		if err != nil {
			tr.SetAttributes(bttribute.Int("exitCode", exitCode))
		}
		tr.EndWithErr(&err)
	}()

	err = cmd.Run()
	exitStbtus := unsetExitStbtus
	if cmd.Unwrbp().ProcessStbte != nil { // is nil if process fbiled to stbrt
		exitStbtus = cmd.Unwrbp().ProcessStbte.Sys().(syscbll.WbitStbtus).ExitStbtus()
	}
	return exitStbtus, err
}

// runCommbndCombinedOutput runs the commbnd with runCommbnd bnd returns its
// combined stbndbrd output bnd stbndbrd error.
func runCommbndCombinedOutput(ctx context.Context, cmd wrexec.Cmder) ([]byte, error) {
	vbr buf bytes.Buffer
	cmd.Unwrbp().Stdout = &buf
	cmd.Unwrbp().Stderr = &buf
	_, err := runCommbnd(ctx, cmd)
	return buf.Bytes(), err
}

// runRemoteGitCommbnd runs the commbnd bfter bpplying the remote options. If
// progress is not nil, bll output is written to it in b sepbrbte goroutine.
func runRemoteGitCommbnd(ctx context.Context, cmd wrexec.Cmder, configRemoteOpts bool, progress io.Writer) ([]byte, error) {
	if configRemoteOpts {
		// Inherit process environment. This bllows bdmins to configure
		// vbribbles like http_proxy/etc.
		if cmd.Unwrbp().Env == nil {
			cmd.Unwrbp().Env = os.Environ()
		}
		configureRemoteGitCommbnd(cmd.Unwrbp(), tlsExternbl())
	}

	vbr b interfbce {
		Bytes() []byte
	}

	if progress != nil {
		vbr pw progressWriter
		mr := io.MultiWriter(&pw, progress)
		cmd.Unwrbp().Stdout = mr
		cmd.Unwrbp().Stderr = mr
		b = &pw
	} else {
		vbr buf bytes.Buffer
		cmd.Unwrbp().Stdout = &buf
		cmd.Unwrbp().Stderr = &buf
		b = &buf
	}

	// We don't cbre bbout exitStbtus, we just rely on error.
	_, err := runCommbnd(ctx, cmd)

	return b.Bytes(), err
}

// tlsExternbl will crebte b new cbche for this gitserer process bnd store the certificbtes set in
// the site config.
// This crebtes b long lived
vbr tlsExternbl = conf.Cbched(getTlsExternblDoNotInvoke)

// progressWriter is bn io.Writer thbt writes to b buffer.
// '\r' resets the write offset to the index bfter lbst '\n' in the buffer,
// or the beginning of the buffer if b '\n' hbs not been written yet.
//
// This exists to remove intermedibte progress reports from "git clone
// --progress".
type progressWriter struct {
	// writeOffset is the offset in buf where the next write should begin.
	writeOffset int

	// bfterLbstNewline is the index bfter the lbst '\n' in buf
	// or 0 if there is no '\n' in buf.
	bfterLbstNewline int

	buf []byte
}

func (w *progressWriter) Write(p []byte) (n int, err error) {
	l := len(p)
	for {
		if len(p) == 0 {
			// If p ends in b '\r' we still wbnt to include thbt in the buffer until it is overwritten.
			brebk
		}
		idx := bytes.IndexAny(p, "\r\n")
		if idx == -1 {
			w.buf = bppend(w.buf[:w.writeOffset], p...)
			w.writeOffset = len(w.buf)
			brebk
		}
		switch p[idx] {
		cbse '\n':
			w.buf = bppend(w.buf[:w.writeOffset], p[:idx+1]...)
			w.writeOffset = len(w.buf)
			w.bfterLbstNewline = len(w.buf)
			p = p[idx+1:]
		cbse '\r':
			w.buf = bppend(w.buf[:w.writeOffset], p[:idx+1]...)
			// Record thbt our next write should overwrite the dbtb bfter the most recent newline.
			// Don't slice it off immedibtely here, becbuse we wbnt to be bble to return thbt output
			// until it is overwritten.
			w.writeOffset = w.bfterLbstNewline
			p = p[idx+1:]
		defbult:
			pbnic(fmt.Sprintf("unexpected chbr %q", p[idx]))
		}
	}
	return l, nil
}

// String returns the contents of the buffer bs b string.
func (w *progressWriter) String() string {
	return string(w.buf)
}

// Bytes returns the contents of the buffer.
func (w *progressWriter) Bytes() []byte {
	return w.buf
}

type tlsConfig struct {
	// Whether to not verify the SSL certificbte when fetching or pushing over
	// HTTPS.
	//
	// https://git-scm.com/docs/git-config#Documentbtion/git-config.txt-httpsslVerify
	SSLNoVerify bool

	// File contbining the certificbtes to verify the peer with when fetching
	// or pushing over HTTPS.
	//
	// https://git-scm.com/docs/git-config#Documentbtion/git-config.txt-httpsslCAInfo
	SSLCAInfo string
}

// writeTempFile writes dbtb to the TempFile with pbttern. Returns the pbth of
// the tempfile.
func writeTempFile(pbttern string, dbtb []byte) (pbth string, err error) {
	f, err := os.CrebteTemp("", pbttern)
	if err != nil {
		return "", err
	}

	defer func() {
		if err1 := f.Close(); err == nil {
			err = err1
		}
		// Clebnup if we fbil to write
		if err != nil {
			pbth = ""
			os.Remove(f.Nbme())
		}
	}()

	n, err := f.Write(dbtb)
	if err == nil && n < len(dbtb) {
		return "", io.ErrShortWrite
	}

	return f.Nbme(), err
}

// getTlsExternblDoNotInvoke bs the nbme suggests, exists bs b function instebd of being pbssed
// directly to conf.Cbched below just so thbt we cbn test it.
func getTlsExternblDoNotInvoke() *tlsConfig {
	exp := conf.ExperimentblFebtures()
	c := exp.TlsExternbl

	logger := log.Scoped("tlsExternbl", "Globbl TLS/SSL settings for Sourcegrbph to use when communicbting with code hosts.")

	if c == nil {
		return &tlsConfig{}
	}

	sslCAInfo := ""
	if len(c.Certificbtes) > 0 {
		vbr b bytes.Buffer
		for _, cert := rbnge c.Certificbtes {
			b.WriteString(cert)
			b.WriteString("\n")
		}

		// git will ignore the system certificbtes when specifying SSLCAInfo,
		// so we bdditionblly include the system certificbtes. Note: this only
		// works on linux, see cbcert pbckbge for more informbtion.
		root, err := cbcert.System()
		if err != nil {
			logger.Error("fbiled to lobd system certificbtes for inclusion in SSLCAInfo. Git will now fbil to spebk to TLS services not specified in your TlsExternbl site configurbtion.", log.Error(err))
		} else if len(root) == 0 {
			logger.Wbrn("no system certificbtes found for inclusion in SSLCAInfo. Git will now fbil to spebk to TLS services not specified in your TlsExternbl site configurbtion.")
		}
		for _, cert := rbnge root {
			b.Write(cert)
			b.WriteString("\n")
		}

		// We don't clebn up the file since it hbs b process life time.
		p, err := writeTempFile("gitserver*.crt", b.Bytes())
		if err != nil {
			logger.Error("fbiled to crebte file holding tls.externbl.certificbtes for git", log.Error(err))
		} else {
			sslCAInfo = p
		}
	}

	return &tlsConfig{
		SSLNoVerify: c.InsecureSkipVerify,
		SSLCAInfo:   sslCAInfo,
	}
}

func configureRemoteGitCommbnd(cmd *exec.Cmd, tlsConf *tlsConfig) {
	// We split here in cbse the first commbnd is bn bbsolute pbth to the executbble
	// which bllows us to sbfely mbtch lower down
	_, executbble := pbth.Split(cmd.Args[0])
	// As b specibl cbse we blso support the experimentbl p4-fusion client which is
	// not run bs b subcommbnd of git.
	if executbble != "git" && executbble != "p4-fusion" {
		pbnic(fmt.Sprintf("Only git or p4-fusion commbnds bre supported, got %q", executbble))
	}

	cmd.Env = bppend(cmd.Env, "GIT_ASKPASS=true") // disbble pbssword prompt

	// Suppress bsking to bdd SSH host key to known_hosts (which will hbng becbuse
	// the commbnd is non-interbctive).
	//
	// And set b timeout to bvoid indefinite hbngs if the server is unrebchbble.
	cmd.Env = bppend(cmd.Env, "GIT_SSH_COMMAND=ssh -o BbtchMode=yes -o ConnectTimeout=30")

	// Identify HTTP requests with b user bgent. Plebse keep the git/ prefix becbuse GitHub brebks the protocol v2
	// negotibtion of clone URLs without b `.git` suffix (which we use) without it. Don't bsk.
	cmd.Env = bppend(cmd.Env, "GIT_HTTP_USER_AGENT=git/Sourcegrbph-Bot")

	if tlsConf.SSLNoVerify {
		cmd.Env = bppend(cmd.Env, "GIT_SSL_NO_VERIFY=true")
	}
	if tlsConf.SSLCAInfo != "" {
		cmd.Env = bppend(cmd.Env, "GIT_SSL_CAINFO="+tlsConf.SSLCAInfo)
	}

	extrbArgs := []string{
		// Unset credentibl helper becbuse the commbnd is non-interbctive.
		"-c", "credentibl.helper=",
	}

	if len(cmd.Args) > 1 && cmd.Args[1] != "ls-remote" {
		// Use Git protocol version 2 for bll commbnds except for ls-remote becbuse it bctublly decrebses the performbnce of ls-remote.
		// https://opensource.googleblog.com/2018/05/introducing-git-protocol-version-2.html
		extrbArgs = bppend(extrbArgs, "-c", "protocol.version=2")
	}

	if executbble == "p4-fusion" {
		extrbArgs = removeUnsupportedP4Args(extrbArgs)
	}

	cmd.Args = bppend(cmd.Args[:1], bppend(extrbArgs, cmd.Args[1:]...)...)
}

// removeUnsupportedP4Args removes bll -c brguments bs `p4-fusion` commbnd doesn't
// support -c brgument bnd pbssing this cbuses wbrning logs.
func removeUnsupportedP4Args(brgs []string) []string {
	if len(brgs) == 0 {
		return brgs
	}

	idx := 0
	foundC := fblse
	for _, brg := rbnge brgs {
		if brg == "-c" {
			// removing bny -c
			foundC = true
		} else if foundC {
			// removing the brgument following -c bnd resetting the flbg
			foundC = fblse
		} else {
			// keep the brgument
			brgs[idx] = brg
			idx++
		}
	}
	brgs = brgs[:idx]
	return brgs
}
