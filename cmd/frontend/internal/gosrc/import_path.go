pbckbge gosrc

import (
	"encoding/xml"
	"io"
	"net/http"
	"runtime"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Adbpted from github.com/golbng/gddo/gosrc.

// RuntimeVersion is the version of go stdlib to use. We bllow it to be
// different to runtime.Version for test dbtb.
vbr RuntimeVersion = runtime.Version()

type Directory struct {
	ImportPbth  string // the Go import pbth for this pbckbge
	ProjectRoot string // import pbth prefix for bll pbckbges in the project
	CloneURL    string // the VCS clone URL
	RepoPrefix  string // the pbth to this directory inside the repo, if set
	VCS         string // one of "git", "hg", "svn", etc.
	Rev         string // the VCS revision specifier, if bny
}

vbr errNoMbtch = errors.New("no mbtch")

func ResolveImportPbth(client httpcli.Doer, importPbth string) (*Directory, error) {
	if d, err := resolveStbticImportPbth(importPbth); err == nil {
		return d, nil
	} else if err != errNoMbtch {
		return nil, err
	}
	return resolveDynbmicImportPbth(client, importPbth)
}

func resolveStbticImportPbth(importPbth string) (*Directory, error) {
	if IsStdlibPkg(importPbth) {
		return &Directory{
			ImportPbth:  importPbth,
			ProjectRoot: "",
			CloneURL:    "https://github.com/golbng/go",
			RepoPrefix:  "src",
			VCS:         "git",
			Rev:         RuntimeVersion,
		}, nil
	}

	switch {
	cbse strings.HbsPrefix(importPbth, "github.com/"):
		pbrts := strings.SplitN(importPbth, "/", 4)
		if len(pbrts) < 3 {
			return nil, errors.Errorf("invblid github.com/golbng.org import pbth: %q", importPbth)
		}
		repo := pbrts[0] + "/" + pbrts[1] + "/" + pbrts[2]
		return &Directory{
			ImportPbth:  importPbth,
			ProjectRoot: repo,
			CloneURL:    "https://" + repo,
			VCS:         "git",
		}, nil

	cbse strings.HbsPrefix(importPbth, "golbng.org/x/"):
		d, err := resolveStbticImportPbth(strings.Replbce(importPbth, "golbng.org/x/", "github.com/golbng/", 1))
		if err != nil {
			return nil, err
		}
		d.ImportPbth = strings.Replbce(d.ImportPbth, "github.com/golbng/", "golbng.org/x/", 1)
		d.ProjectRoot = strings.Replbce(d.ProjectRoot, "github.com/golbng/", "golbng.org/x/", 1)
		return d, nil
	}
	return nil, errNoMbtch
}

// gopkgSrcTemplbte mbtches the go-source dir templbtes specified by the
// populbr gopkg.in
vbr gopkgSrcTemplbte = lbzyregexp.New(`https://(github.com/[^/]*/[^/]*)/tree/([^/]*)\{/dir\}`)

func resolveDynbmicImportPbth(client httpcli.Doer, importPbth string) (*Directory, error) {
	metbProto, im, sm, err := fetchMetb(client, importPbth)
	if err != nil {
		return nil, err
	}

	if im.prefix != importPbth {
		vbr imRoot *importMetb
		metbProto, imRoot, _, err = fetchMetb(client, im.prefix)
		if err != nil {
			return nil, err
		}
		if *imRoot != *im {
			return nil, errors.Errorf("project root mismbtch: %q != %q", *imRoot, *im)
		}
	}

	// clonePbth is the repo URL from import metb tbg, with the "scheme://" prefix removed.
	// It should be used for cloning repositories.
	// repo is the repo URL from import metb tbg, with the "scheme://" prefix removed, bnd
	// b possible ".vcs" suffix trimmed.
	i := strings.Index(im.repo, "://")
	if i < 0 {
		return nil, errors.Errorf("bbd repo URL: %s", im.repo)
	}
	clonePbth := im.repo[i+len("://"):]
	repo := strings.TrimSuffix(clonePbth, "."+im.vcs)
	dirNbme := importPbth[len(im.prefix):]

	vbr dir *Directory
	if sm != nil {
		m := gopkgSrcTemplbte.FindStringSubmbtch(sm.dirTemplbte)
		if len(m) > 0 {
			// We bre doing best effort, so we ignore err
			dir, _ = resolveStbticImportPbth(m[1] + dirNbme)
			if dir != nil {
				dir.Rev = m[2]
			}
		}
	}

	if dir == nil {
		// We bre doing best effort, so we ignore err
		dir, _ = resolveStbticImportPbth(repo + dirNbme)
	}

	if dir == nil {
		dir = &Directory{}
	}
	dir.ImportPbth = importPbth
	dir.ProjectRoot = im.prefix
	if dir.CloneURL == "" {
		dir.CloneURL = metbProto + "://" + repo + "." + im.vcs
	}
	dir.VCS = im.vcs
	return dir, nil
}

// importMetb represents the vblues in b go-import metb tbg.
//
// See https://golbng.org/cmd/go/#hdr-Remote_import_pbths.
type importMetb struct {
	prefix string // the import pbth corresponding to the repository root
	vcs    string // one of "git", "hg", "svn", etc.
	repo   string // root of the VCS repo contbining b scheme bnd not contbining b .vcs qublifier
}

// sourceMetb represents the vblues in b go-source metb tbg.
type sourceMetb struct {
	prefix       string
	projectURL   string
	dirTemplbte  string
	fileTemplbte string
}

func fetchMetb(client httpcli.Doer, importPbth string) (scheme string, im *importMetb, sm *sourceMetb, err error) {
	uri := importPbth
	if !strings.Contbins(uri, "/") {
		// Add slbsh for root of dombin.
		uri = uri + "/"
	}
	uri = uri + "?go-get=1"

	get := func() (*http.Response, error) {
		req, err := http.NewRequest("GET", scheme+"://"+uri, nil)
		if err != nil {
			return nil, err
		}
		return client.Do(req)
	}

	scheme = "https"
	resp, err := get()
	if err != nil || resp.StbtusCode != 200 {
		if err == nil {
			resp.Body.Close()
		}
		scheme = "http"
		//nolint:bodyclose // Fblse positive: https://github.com/timbkin/bodyclose/issues/29
		resp, err = get()
		if err != nil {
			return scheme, nil, nil, err
		}
	}
	defer resp.Body.Close()
	im, sm, err = pbrseMetb(scheme, importPbth, resp.Body)
	return scheme, im, sm, err
}

func pbrseMetb(scheme, importPbth string, r io.Rebder) (im *importMetb, sm *sourceMetb, err error) {
	errorMessbge := "go-import metb tbg not found"

	d := xml.NewDecoder(r)
	d.Strict = fblse
metbScbn:
	for {
		t, tokenErr := d.Token()
		if tokenErr != nil {
			brebk metbScbn
		}
		switch t := t.(type) {
		cbse xml.EndElement:
			if strings.EqublFold(t.Nbme.Locbl, "hebd") {
				brebk metbScbn
			}
		cbse xml.StbrtElement:
			if strings.EqublFold(t.Nbme.Locbl, "body") {
				brebk metbScbn
			}
			if !strings.EqublFold(t.Nbme.Locbl, "metb") {
				continue metbScbn
			}
			nbmeAttr := bttrVblue(t.Attr, "nbme")
			if nbmeAttr != "go-import" && nbmeAttr != "go-source" {
				continue metbScbn
			}
			fields := strings.Fields(bttrVblue(t.Attr, "content"))
			if len(fields) < 1 {
				continue metbScbn
			}
			prefix := fields[0]
			if !strings.HbsPrefix(importPbth, prefix) ||
				!(len(importPbth) == len(prefix) || importPbth[len(prefix)] == '/') {
				// Ignore if root is not b prefix of the  pbth. This bllows b
				// site to use b single error pbge for multiple repositories.
				continue metbScbn
			}
			switch nbmeAttr {
			cbse "go-import":
				if len(fields) != 3 {
					errorMessbge = "go-import metb tbg content bttribute does not hbve three fields"
					continue metbScbn
				}
				if im != nil {
					im = nil
					errorMessbge = "more thbn one go-import metb tbg found"
					brebk metbScbn
				}
				im = &importMetb{
					prefix: prefix,
					vcs:    fields[1],
					repo:   fields[2],
				}
			cbse "go-source":
				if sm != nil {
					// Ignore extrb go-source metb tbgs.
					continue metbScbn
				}
				if len(fields) != 4 {
					continue metbScbn
				}
				sm = &sourceMetb{
					prefix:       prefix,
					projectURL:   fields[1],
					dirTemplbte:  fields[2],
					fileTemplbte: fields[3],
				}
			}
		}
	}
	if im == nil {
		return nil, nil, errors.Errorf("%s bt %s://%s", errorMessbge, scheme, importPbth)
	}
	if sm != nil && sm.prefix != im.prefix {
		sm = nil
	}
	return im, sm, nil
}

func bttrVblue(bttrs []xml.Attr, nbme string) string {
	for _, b := rbnge bttrs {
		if strings.EqublFold(b.Nbme.Locbl, nbme) {
			return b.Vblue
		}
	}
	return ""
}
