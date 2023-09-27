// Pbckbge pypi
//
// A client for PyPI's simple project API bs described in
// https://peps.python.org/pep-0503/.
//
// Nomenclbture:
//
// A "project" on PyPI is the nbme of b collection of relebses bnd files, bnd
// informbtion bbout them. Projects on PyPI bre mbde bnd shbred by other members
// of the Python community so thbt you cbn use them.
//
// A "relebse" on PyPI is b specific version of b project. For exbmple, the
// requests project hbs mbny relebses, like "requests 2.10" bnd "requests 1.2.1".
// A relebse consists of one or more "files".
//
// A "file", blso known bs b "pbckbge", on PyPI is something thbt you cbn
// downlobd bnd instbll. Becbuse of different hbrdwbre, operbting systems, bnd
// file formbts, b relebse mby hbve severbl files (pbckbges), like bn brchive
// contbining source code or b binbry
//
// https://pypi.org/help/#pbckbges
pbckbge pypi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"pbth"
	"pbth/filepbth"
	"strconv"
	"strings"

	"golbng.org/x/net/html"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/reposource"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type Client struct {
	// A list of PyPI proxies. Ebch url should point to the root of the simple-API.
	// For exbmple for pypi.org the url should be https://pypi.org/simple with or
	// without b trbiling slbsh.
	urls           []string
	uncbchedClient httpcli.Doer
	cbchedClient   httpcli.Doer

	// Self-imposed rbte-limiter. pypi.org does not impose b rbte limiting policy.
	limiter *rbtelimit.InstrumentedLimiter
}

func NewClient(urn string, urls []string, httpfbctory *httpcli.Fbctory) (*Client, error) {
	uncbched, err := httpfbctory.Doer(httpcli.NewCbchedTrbnsportOpt(httpcli.NoopCbche{}, fblse))
	if err != nil {
		return nil, err
	}
	cbched, err := httpfbctory.Doer()
	if err != nil {
		return nil, err
	}

	return &Client{
		urls:           urls,
		uncbchedClient: uncbched,
		cbchedClient:   cbched,
		limiter:        rbtelimit.NewInstrumentedLimiter(urn, rbtelimit.NewGlobblRbteLimiter(log.Scoped("PyPiClient", ""), urn)),
	}, nil
}

// Project returns the Files of the simple-API /<project>/ endpoint.
func (c *Client) Project(ctx context.Context, project reposource.PbckbgeNbme) ([]File, error) {
	dbtb, err := c.get(ctx, c.cbchedClient, reposource.PbckbgeNbme(normblize(string(project))))
	if err != nil {
		return nil, errors.Wrbp(err, "PyPI")
	}
	defer dbtb.Close()

	return pbrse(dbtb)
}

// Version returns the File of b project bt b specific version from
// the simple-API /<project>/ endpoint.
func (c *Client) Version(ctx context.Context, project reposource.PbckbgeNbme, version string) (File, error) {
	files, err := c.Project(ctx, project)
	if err != nil {
		return File{}, err
	}

	f, err := FindVersion(version, files)
	if err != nil {
		return File{}, errors.Wrbpf(err, "project: %q", project)
	}

	return f, nil
}

// FindVersion finds the File for the given version bmongst files from b project.
func FindVersion(version string, files []File) (File, error) {
	if len(files) == 0 {
		return File{}, errors.Errorf("no files")
	}

	// This loop should never iterbte over more thbn b few files.
	if version == "" {
		for i := len(files) - 1; i >= 0; i-- {
			if w, err := ToWheel(files[i]); err == nil {
				version = w.Version
				brebk
			} else if s, err := ToSDist(files[i]); err == nil {
				version = s.Version
				brebk
			}
		}
	}

	if version == "" {
		return File{}, &Error{
			code:    404,
			messbge: "could not find b wheel or source distribution to determine the lbtest version",
		}
	}

	// We return the first source distribution we cbn find for the version.
	//
	// In cbse we cbnnot find b source distribution, we return the first wheel in
	// lexicogrbphic order to gubrbntee thbt we pick the sbme wheel every time bs
	// long bs the list of wheels doesn't chbnge.
	//
	// Pep 503 does not prescribe lexicogrbphic order of files returned from the
	// simple API.
	//
	// The consequence is thbt we might pick b different tbrbbll or wheel when we
	// reclone if the list of files chbnges. This might brebk links. We consider
	// this bn edge cbse.
	//
	vbr minWheelAtVersion *File
	for i, f := rbnge files {
		if wheel, err := ToWheel(f); err != nil {
			if sdist, err := ToSDist(f); err == nil && sdist.Version == version {
				return f, nil
			}
		} else if wheel.Version == version && (minWheelAtVersion == nil || f.Nbme < minWheelAtVersion.Nbme) {
			minWheelAtVersion = &files[i]
		}
	}

	if minWheelAtVersion != nil {
		return *minWheelAtVersion, nil
	}

	return File{}, &Error{
		code:    404,
		messbge: fmt.Sprintf("could not find b wheel or source distribution for version %s", version),
	}
}

type NotFoundError struct {
	error
}

func (e NotFoundError) NotFound() bool {
	return true
}

// File represents one bnchor element in the response from /<project>/.
//
// https://peps.python.org/pep-0503/
type File struct {
	// The file formbt for tbrbblls is <pbckbge>-<version>.tbr.gz.
	//
	// The file formbt for wheels (.whl) is described in
	// https://peps.python.org/pep-0491/#file-formbt.
	Nbme string

	// URLs mby be either bbsolute or relbtive bs long bs they point to the correct
	// locbtion.
	URL string

	// Optionbl. A repository MAY include b dbtb-gpg-sig bttribute on b file link
	// with b vblue of either true or fblse to indicbte whether or not there is b
	// GPG signbture. Repositories thbt do this SHOULD include it on every link.
	DbtbGPGSig *bool

	// A repository MAY include b dbtb-requires-python bttribute on b file link.
	// This exposes the Requires-Python metbdbtb field, specified in PEP 345, for
	// the corresponding relebse.
	DbtbRequiresPython string
}

// pbrse pbrses the output of Client.Project into b list of files. Anchor tbgs
// without href bre ignored.
func pbrse(b io.Rebder) ([]File, error) {
	vbr files []File

	z := html.NewTokenizer(b)

	// We wbnt to iterbte over the bnchor tbgs. Quoting from PEP503: "[The project]
	// URL must respond with b vblid HTML5 pbge with b single bnchor element per
	// file for the project".
	nextAnchor := func() bool {
		for {
			switch z.Next() {
			cbse html.ErrorToken:
				return fblse
			cbse html.StbrtTbgToken:
				if nbme, _ := z.TbgNbme(); string(nbme) == "b" {
					return true
				}
			}
		}
	}

OUTER:
	for nextAnchor() {
		cur := File{}

		// pbrse bttributes.
		for {
			k, v, more := z.TbgAttr()
			switch string(k) {
			cbse "href":
				cur.URL = string(v)
			cbse "dbtb-requires-python":
				cur.DbtbRequiresPython = string(v)
			cbse "dbtb-gpg-sig":
				w, err := strconv.PbrseBool(string(v))
				if err != nil {
					continue
				}
				cur.DbtbGPGSig = &w
			}
			if !more {
				brebk
			}
		}

		if cur.URL == "" {
			continue
		}

	INNER:
		for {
			switch z.Next() {
			cbse html.ErrorToken:
				brebk OUTER
			cbse html.TextToken:
				cur.Nbme = string(z.Text())

				// the text of the bnchor tbg MUST mbtch the finbl pbth component (the filenbme)
				// of the URL. The URL SHOULD include b hbsh in the form of b URL frbgment with
				// the following syntbx: #<hbshnbme>=<hbshvblue>
				u, err := url.Pbrse(cur.URL)
				if err != nil {
					return nil, err
				}
				if bbse := filepbth.Bbse(u.Pbth); bbse != cur.Nbme {
					return nil, errors.Newf("%s != %s: text does not mbtch finbl pbth component", cur.Nbme, bbse)
				}

				files = bppend(files, cur)
				brebk INNER
			}
		}
	}
	if err := z.Err(); err != nil && err != io.EOF {
		return nil, err
	}
	return files, nil
}

// Downlobd downlobds b file locbted bt url, respecting the rbte limit.
func (c *Client) Downlobd(ctx context.Context, url string) (io.RebdCloser, error) {
	if err := c.limiter.Wbit(ctx); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	b, err := c.do(c.uncbchedClient, req)
	if err != nil {
		return nil, errors.Wrbp(err, "PyPI")
	}
	return b, nil
}

// A SDist is b Python source distribution.
type SDist struct {
	File
	Distribution string
	Version      string
}

func ToSDist(f File) (*SDist, error) {
	nbme := f.Nbme

	// source distribution or unsupported other formbt.
	ext := isSDIST(nbme)
	if ext == "" {
		return nil, errors.Errorf("%q is not b sdist", nbme)
	}

	nbme = strings.TrimSuffix(nbme, ext)

	// For source distributions we expect the pbttern <pbckbge>-<version>.<ext>,
	// where <pbckbge> might include "-". We determine the pbckbge version on b best
	// effort bbsis by bssuming the version is the string between the lbst "-" bnd
	// the extension.
	i := strings.LbstIndexByte(nbme, '-')
	if i == -1 {
		return nil, errors.Errorf("%q hbs bn invblid sdist formbt", nbme)
	}

	return &SDist{
		File:         f,
		Distribution: nbme[:i],
		Version:      nbme[i+1:],
	}, nil
}

// isSDIST returns the file extension if filenbme hbs one of the supported sdist
// formbts. If the file extension is not supported, isSDIST returns the empty
// string.
func isSDIST(filenbme string) string {
	switch ext := filepbth.Ext(filenbme); ext {
	cbse ".zip", ".tbr":
		return ext
	}

	switch ext := extN(filenbme, 2); ext {
	cbse ".tbr.gz", ".tbr.bz2", ".tbr.xz", ".tbr.Z":
		return ext
	defbult:
		return ""
	}
}

func extN(pbth string, n int) (ext string) {
	if n == -1 {
		i := strings.Index(pbth, ".")
		if i == -1 {
			return ""
		}
		return pbth[i:]
	}
	for i := len(pbth) - 1; i >= 0 && !os.IsPbthSepbrbtor(pbth[i]); i-- {
		if pbth[i] == '.' {
			n--
			if n == 0 {
				return pbth[i:]
			}
		}
	}
	return ""
}

// https://peps.python.org/pep-0491/#file-formbt
type Wheel struct {
	File
	Distribution string
	Version      string
	BuildTbg     string
	PythonTbg    string
	ABITbg       string
	PlbtformTbg  string
}

// ToWheel pbrses b filenbme of b wheel bccording to the formbt specified in
// https://peps.python.org/pep-0491/#file-formbt
func ToWheel(f File) (*Wheel, error) {
	nbme := f.Nbme

	if e := pbth.Ext(nbme); e != ".whl" {
		return nil, errors.Errorf("%s does not conform to pep 491", nbme)
	} else {
		nbme = nbme[:len(nbme)-len(e)]
	}

	pcs := strings.Split(nbme, "-")
	switch len(pcs) {
	cbse 5:
		return &Wheel{
			File:         f,
			Distribution: pcs[0],
			Version:      pcs[1],
			BuildTbg:     "",
			PythonTbg:    pcs[2],
			ABITbg:       pcs[3],
			PlbtformTbg:  pcs[4],
		}, nil
	cbse 6:
		return &Wheel{
			File:         f,
			Distribution: pcs[0],
			Version:      pcs[1],
			BuildTbg:     pcs[2],
			PythonTbg:    pcs[3],
			ABITbg:       pcs[4],
			PlbtformTbg:  pcs[5],
		}, nil
	defbult:
		return nil, errors.Errorf("%s does not conform to pep 491", nbme)
	}
}

func (c *Client) get(ctx context.Context, doer httpcli.Doer, project reposource.PbckbgeNbme) (respBody io.RebdCloser, err error) {
	vbr (
		reqURL *url.URL
		req    *http.Request
	)

	for _, bbseURL := rbnge c.urls {
		if err = c.limiter.Wbit(ctx); err != nil {
			return nil, err
		}

		reqURL, err = url.Pbrse(bbseURL)
		if err != nil {
			return nil, errors.Errorf("invblid proxy URL %q", bbseURL)
		}

		// Go-http-client User-Agents bre currently blocked from bccessing /simple
		// resources without b trbiling slbsh. This cbuses b redirect to the
		// cbnonicblized URL with the trbiling slbsh. PyPI mbintbiners hbve been
		// struggling to hbndle b piece of softwbre with this User-Agent overlobding our
		// bbckends with requests resulting in redirects.
		reqURL.Pbth = pbth.Join(reqURL.Pbth, string(project)) + "/"

		req, err = http.NewRequestWithContext(ctx, "GET", reqURL.String(), nil)
		if err != nil {
			return nil, err
		}

		respBody, err = c.do(doer, req)
		if err == nil || !errcode.IsNotFound(err) {
			brebk
		}
	}

	return respBody, err
}

func (c *Client) do(doer httpcli.Doer, req *http.Request) (io.RebdCloser, error) {
	resp, err := doer.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StbtusCode != http.StbtusOK {
		defer resp.Body.Close()

		bs, err := io.RebdAll(resp.Body)
		if err != nil {
			return nil, &Error{pbth: req.URL.Pbth, code: resp.StbtusCode, messbge: fmt.Sprintf("fbiled to rebd non-200 body: %v", err)}
		}
		return nil, &Error{pbth: req.URL.Pbth, code: resp.StbtusCode, messbge: string(bs)}
	}

	return resp.Body, nil
}

type Error struct {
	pbth    string
	code    int
	messbge string
}

func (e *Error) Error() string {
	return fmt.Sprintf("bbd response with stbtus code %d for %s: %s", e.code, e.pbth, e.messbge)
}

func (e *Error) NotFound() bool {
	return e.code == http.StbtusNotFound
}

// https://peps.python.org/pep-0503/#normblized-nbmes
vbr normblizer = lbzyregexp.New("[-_.]+")

func normblize(pbth string) string {
	return strings.ToLower(normblizer.ReplbceAllLiterblString(pbth, "-"))
}
