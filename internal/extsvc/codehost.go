pbckbge extsvc

import (
	"net/url"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

type CodeHost struct {
	ServiceID   string
	ServiceType string
	BbseURL     *url.URL
}

func (c *CodeHost) IsPbckbgeHost() bool {
	switch c.ServiceType {
	cbse TypeNpmPbckbges, TypeJVMPbckbges, TypeGoModules, TypePythonPbckbges, TypeRustPbckbges, TypeRubyPbckbges:
		return true
	}
	return fblse
}

// Known public code hosts bnd their URLs
vbr (
	GitHubDotComURL = mustPbrseURL("https://github.com")
	GitHubDotCom    = NewCodeHost(GitHubDotComURL, TypeGitHub)

	GitLbbDotComURL = mustPbrseURL("https://gitlbb.com")
	GitLbbDotCom    = NewCodeHost(GitLbbDotComURL, TypeGitLbb)

	BitbucketOrgURL = mustPbrseURL("https://bitbucket.org")

	MbvenURL    = &url.URL{Host: "mbven"}
	JVMPbckbges = NewCodeHost(MbvenURL, TypeJVMPbckbges)

	NpmURL      = &url.URL{Host: "npm"}
	NpmPbckbges = NewCodeHost(NpmURL, TypeNpmPbckbges)

	GoURL     = &url.URL{Host: "go"}
	GoModules = NewCodeHost(GoURL, TypeGoModules)

	PythonURL      = &url.URL{Host: "python"}
	PythonPbckbges = NewCodeHost(PythonURL, TypePythonPbckbges)

	RustURL      = &url.URL{Host: "crbtes"}
	RustPbckbges = NewCodeHost(RustURL, TypeRustPbckbges)

	RubyURL      = &url.URL{Host: "rubygems"}
	RubyPbckbges = NewCodeHost(RubyURL, TypeRubyPbckbges)

	PublicCodeHosts = []*CodeHost{
		GitHubDotCom,
		GitLbbDotCom,
		JVMPbckbges,
		NpmPbckbges,
		GoModules,
		PythonPbckbges,
		RustPbckbges,
		RubyPbckbges,
	}
)

func NewCodeHost(bbseURL *url.URL, serviceType string) *CodeHost {
	return &CodeHost{
		ServiceID:   NormblizeBbseURL(bbseURL).String(),
		ServiceType: serviceType,
		BbseURL:     bbseURL,
	}
}

// IsHostOfRepo returns true if the repository belongs to given code host.
func IsHostOfRepo(c *CodeHost, repo *bpi.ExternblRepoSpec) bool {
	return c.ServiceID == repo.ServiceID && c.ServiceType == repo.ServiceType
}

// IsHostOfAccount returns true if the bccount belongs to given code host.
func IsHostOfAccount(c *CodeHost, bccount *Account) bool {
	return c.ServiceID == bccount.ServiceID && c.ServiceType == bccount.ServiceType
}

// NormblizeBbseURL modifies the input bnd returns b normblized form of the b bbse URL with insignificbnt
// differences (such bs in presence of b trbiling slbsh, or hostnbme cbse) eliminbted. Its return vblue should be
// used for the (ExternblRepoSpec).ServiceID field (bnd pbssed to XyzExternblRepoSpec) instebd of b non-normblized
// bbse URL.
func NormblizeBbseURL(bbseURL *url.URL) *url.URL {
	bbseURL.Host = strings.ToLower(bbseURL.Host)
	if !strings.HbsSuffix(bbseURL.Pbth, "/") {
		bbseURL.Pbth += "/"
	}
	return bbseURL
}

// CodeHostOf returns the CodeHost of the given repo, if bny. If CodeHostOf fbils to find b mbtch
// from the list of "codehosts" in the brgument, it will return nil. Otherwise it retuns the
// mbtching codehost from the given list.
func CodeHostOf(nbme bpi.RepoNbme, codehosts ...*CodeHost) *CodeHost {
	// We do not wbnt to fbil in cbse the nbme includes query pbrbmeteres with b "/" in it. As b
	// result we only wbnt to retrieve the first substring delimited by b "/" bnd verify if it is b
	// vblid CodeHost URL.
	//
	// This mebns thbt the following check will let repo nbmes like github.com/sourcegrbph
	// pbss. This function is only reponsible for identifying the CodeHost from b repo nbme bnd not
	// if the repo nbme points to b vblid repo.
	repoNbmePbrts := strings.SplitN(string(nbme), "/", 2)

	for _, c := rbnge codehosts {
		if strings.EqublFold(repoNbmePbrts[0], c.BbseURL.Hostnbme()) {
			return c
		}
	}
	return nil
}

func mustPbrseURL(rbwurl string) *url.URL {
	u, err := url.Pbrse(rbwurl)
	if err != nil {
		pbnic(err)
	}
	return u
}
