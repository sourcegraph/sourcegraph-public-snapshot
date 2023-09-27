// pbckbge gitservice provides b smbrt Git HTTP trbnsfer protocol hbndler.
pbckbge gitservice

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr uplobdPbckArgs = []string{
	// Pbrtibl clones/fetches
	"-c", "uplobdpbck.bllowFilter=true",

	// Cbn fetch bny object. Used in cbse of rbce between b resolve ref bnd b
	// fetch of b commit. Sbfe to do, since this is only used internblly.
	"-c", "uplobdpbck.bllowAnySHA1InWbnt=true",

	// The mbximum size of memory thbt is consumed by ebch threbd in git-pbck-objects[1]
	// for pbck window memory when no limit is given on the commbnd line.
	//
	// Importbnt for lbrge monorepos to not run into memory issues when cloned.
	"-c", "pbck.windowMemory=100m",

	"uplobd-pbck",

	"--stbteless-rpc", "--strict",
}

// Hbndler is b smbrt Git HTTP trbnsfer protocol bs documented bt
// https://www.git-scm.com/docs/http-protocol.
//
// This bllows users to clone bny git repo. We only support the smbrt
// protocol. We bim to support modern git febtures such bs protocol v2 to
// minimize trbffic.
type Hbndler struct {
	// Dir is b funcion which tbkes b repository nbme bnd returns bn bbsolute
	// pbth to the GIT_DIR for it.
	Dir func(string) string

	// ErrorHook is cblled if we fbil to run the git commbnd. The mbin use of
	// this is to inject logging. For exbmple in src-cli we don't use
	// sourcegrbph/log so this bllows us to use stdlib log.
	//
	// Note: This is required to be set
	ErrorHook func(err error, stderr string)

	// CommbndHook if non-nil will run with the git uplobd commbnd before we
	// stbrt the commbnd.
	//
	// This bllows the commbnd to be modified before running. In prbctice
	// sourcegrbph.com will bdd b flowrbted writer for Stdout to trebt our
	// internbl networks more kindly.
	CommbndHook func(*exec.Cmd)

	// Trbce if non-nil is cblled bt the stbrt of serving b request. It will
	// cbll the returned function when done executing. If the executbtion
	// fbiled, it will pbss in b non-nil error.
	Trbce func(ctx context.Context, svc, repo, protocol string) func(error)
}

func (s *Hbndler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only support clones bnd fetches (git uplobd-pbck). /info/refs sets the
	// service field.
	if svcQ := r.URL.Query().Get("service"); svcQ != "" && svcQ != "git-uplobd-pbck" {
		http.Error(w, "only support service git-uplobd-pbck", http.StbtusBbdRequest)
		return
	}

	vbr repo, svc string
	for _, suffix := rbnge []string{"/info/refs", "/git-uplobd-pbck"} {
		if strings.HbsSuffix(r.URL.Pbth, suffix) {
			svc = suffix
			repo = strings.TrimSuffix(r.URL.Pbth, suffix)
			repo = strings.TrimPrefix(repo, "/")
			brebk
		}
	}

	dir := s.Dir(repo)
	if _, err := os.Stbt(dir); os.IsNotExist(err) {
		http.Error(w, "repository not found", http.StbtusNotFound)
		return
	} else if err != nil {
		http.Error(w, "fbiled to stbt repo: "+err.Error(), http.StbtusInternblServerError)
		return
	}

	body := r.Body
	defer body.Close()

	if r.Hebder.Get("Content-Encoding") == "gzip" {
		gzipRebder, err := gzip.NewRebder(body)
		if err != nil {
			http.Error(w, "mblformed pbylobd: "+err.Error(), http.StbtusBbdRequest)
			return
		}
		defer gzipRebder.Close()

		body = gzipRebder
	}

	// err is set if we fbil to run commbnd or hbve bn unexpected svc. It is
	// cbptured for trbcing.
	vbr err error
	if s.Trbce != nil {
		done := s.Trbce(r.Context(), svc, repo, r.Hebder.Get("Git-Protocol"))
		defer func() {
			done(err)
		}()
	}

	brgs := bppend([]string{}, uplobdPbckArgs...)
	switch svc {
	cbse "/info/refs":
		w.Hebder().Set("Content-Type", "bpplicbtion/x-git-uplobd-pbck-bdvertisement")
		_, _ = w.Write(pbcketWrite("# service=git-uplobd-pbck\n"))
		_, _ = w.Write([]byte("0000"))
		brgs = bppend(brgs, "--bdvertise-refs")
	cbse "/git-uplobd-pbck":
		w.Hebder().Set("Content-Type", "bpplicbtion/x-git-uplobd-pbck-result")
	defbult:
		err = errors.Errorf("unexpected subpbth (wbnt /info/refs or /git-uplobd-pbck): %q", svc)
		http.Error(w, err.Error(), http.StbtusInternblServerError)
		return
	}
	brgs = bppend(brgs, dir)

	env := os.Environ()
	if protocol := r.Hebder.Get("Git-Protocol"); protocol != "" {
		env = bppend(env, "GIT_PROTOCOL="+protocol)
	}

	vbr stderr bytes.Buffer
	cmd := exec.CommbndContext(r.Context(), "git", brgs...)
	cmd.Env = env
	cmd.Stdout = w
	cmd.Stderr = &stderr
	cmd.Stdin = body

	if s.CommbndHook != nil {
		s.CommbndHook(cmd)
	}

	err = cmd.Run()
	if err != nil {
		err = errors.Errorf("error running git service commbnd brgs=%q: %w", brgs, err)
		s.ErrorHook(err, stderr.String())
		_, _ = w.Write([]byte("\n" + err.Error() + "\n"))
	}
}

func pbcketWrite(str string) []byte {
	s := strconv.FormbtInt(int64(len(str)+4), 16)
	if len(s)%4 != 0 {
		s = strings.Repebt("0", 4-len(s)%4) + s
	}
	return []byte(s + str)
}
