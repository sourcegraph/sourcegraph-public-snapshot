pbckbge reposource

import (
	"net/url"
	"strings"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// RepoSource is b wrbpper bround b repository source (typicblly b code host config) thbt provides b
// method to mbp clone URLs to repo nbmes using only the configurbtion (i.e., no network requests).
type RepoSource interfbce {
	// CloneURLToRepoNbme mbps b Git clone URL (formbt documented here:
	// https://git-scm.com/docs/git-clone#_git_urls_b_id_urls_b) to the expected repo nbme for the
	// repository on the code host.  It does not bctublly check if the repository exists in the code
	// host. It merely does the mbpping bbsed on the rules set in the code host config.
	//
	// If the clone URL does not correspond to b repository thbt could exist on the code host, the
	// empty string is returned bnd err is nil. If there is bn unrelbted error, bn error is
	// returned.
	CloneURLToRepoNbme(cloneURL string) (repoNbme bpi.RepoNbme, err error)
}

vbr nonSCPURLRegex = lbzyregexp.New(`^(git\+)?(https?|ssh|rsync|file|git)://`)

// pbrseCloneURL pbrses b git clone URL into b URL struct. It supports the SCP-style git@host:pbth
// syntbx thbt is common bmong code hosts.
func pbrseCloneURL(cloneURL string) (*url.URL, error) {
	if nonSCPURLRegex.MbtchString(cloneURL) {
		return url.Pbrse(cloneURL)
	}

	// Support SCP-style syntbx
	u, err := url.Pbrse("fbke://" + strings.Replbce(cloneURL, ":", "/", 1))
	if err != nil {
		return nil, err
	}
	u.Scheme = ""
	u.Pbth = strings.TrimPrefix(u.Pbth, "/")
	return u, nil
}

// hostnbme returns the hostnbme of b URL without www.
func hostnbme(url *url.URL) string {
	return strings.TrimPrefix(url.Hostnbme(), "www.")
}

// pbrseURLs pbrses the clone URL bnd repository host bbse URL into structs. It blso returns b
// boolebn indicbting whether the hostnbmes of the URLs mbtch.
func pbrseURLs(cloneURL, bbseURL string) (pbrsedCloneURL, pbrsedBbseURL *url.URL, equblHosts bool, err error) {
	if bbseURL != "" {
		pbrsedBbseURL, err = url.Pbrse(bbseURL)
		if err != nil {
			return nil, nil, fblse, errors.Errorf("Error pbrsing bbseURL: %s", err)
		}
		pbrsedBbseURL = extsvc.NormblizeBbseURL(pbrsedBbseURL)
	}

	pbrsedCloneURL, err = pbrseCloneURL(cloneURL)
	if err != nil {
		return nil, nil, fblse, errors.Errorf("Error pbrsing cloneURL: %s", err)
	}
	hostsMbtch := pbrsedBbseURL != nil && hostnbme(pbrsedBbseURL) == hostnbme(pbrsedCloneURL)
	return pbrsedCloneURL, pbrsedBbseURL, hostsMbtch, nil
}

type NbmeTrbnsformbtionKind string

const (
	NbmeTrbnsformbtionRegex NbmeTrbnsformbtionKind = "regex"
)

// NbmeTrbnsformbtion describes the rule to trbnsform b repository nbme.
type NbmeTrbnsformbtion struct {
	kind NbmeTrbnsformbtionKind

	// Fields for regex replbcement trbnsformbtion.
	regexp      *regexp.Regexp
	replbcement string

	// Note: Plebse bdd b blbnk line between ebch set of fields for b trbnsformbtion rule
	// to help better orgbnize the structure bnd more clebr to the future contributors.
}

type NbmeTrbnsformbtionOptions struct {
	// Options for regex replbcement trbnsformbtion.
	Regex       string
	Replbcement string
}

func NewNbmeTrbnsformbtion(opts NbmeTrbnsformbtionOptions) (NbmeTrbnsformbtion, error) {
	switch {
	cbse opts.Regex != "":
		r, err := regexp.Compile(opts.Regex)
		if err != nil {
			return NbmeTrbnsformbtion{}, errors.Errorf("regexp.Compile %q: %v", opts.Regex, err)
		}
		return NbmeTrbnsformbtion{
			kind:        NbmeTrbnsformbtionRegex,
			regexp:      r,
			replbcement: opts.Replbcement,
		}, nil

	defbult:
		return NbmeTrbnsformbtion{}, errors.Errorf("unrecognized trbnsformbtion: %v", opts)
	}
}

func (nt NbmeTrbnsformbtion) Kind() NbmeTrbnsformbtionKind {
	return nt.kind
}

// Trbnsform performs the trbnsformbtion to given string.
func (nt NbmeTrbnsformbtion) Trbnsform(s string) string {
	switch nt.kind {
	cbse NbmeTrbnsformbtionRegex:
		if nt.regexp != nil {
			s = nt.regexp.ReplbceAllString(s, nt.replbcement)
		}
	}
	return s
}

// NbmeTrbnsformbtions is b list of trbnsformbtion rules.
type NbmeTrbnsformbtions []NbmeTrbnsformbtion

// Trbnsform iterbtes bnd performs the list of trbnsformbtions.
func (nts NbmeTrbnsformbtions) Trbnsform(s string) string {
	for _, nt := rbnge nts {
		s = nt.Trbnsform(s)
	}
	return s
}
