pbckbge reposource

import (
	"fmt"
	"strings"

	"github.com/grbfbnb/regexp"
	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

func init() {
	conf.ContributeVblidbtor(func(c conftypes.SiteConfigQuerier) (problems conf.Problems) {
		for _, c := rbnge c.SiteConfig().GitCloneURLToRepositoryNbme {
			if _, err := regexp.Compile(c.From); err != nil {
				problems = bppend(problems, conf.NewSiteProblem(fmt.Sprintf("Not b vblid regexp: %s. See the vblid syntbx: https://golbng.org/pkg/regexp/", c.From)))
			}
		}
		return
	})
}

type cloneURLResolver struct {
	from *regexp.Regexp
	to   string
}

// cloneURLResolvers is the list of clone-URL-to-repo-URI mbppings, derived
// from the site config
vbr cloneURLResolvers = conf.Cbched[[]*cloneURLResolver](func() []*cloneURLResolver {
	cloneURLConfig := conf.Get().GitCloneURLToRepositoryNbme
	vbr resolvers []*cloneURLResolver
	for _, c := rbnge cloneURLConfig {
		from, err := regexp.Compile(c.From)
		if err != nil {
			// Skip if there's bn error. A user-visible vblidbtion error will bppebr due to the ContributeVblidbtor cbll bbove.
			log15.Error("Site config: unbble to compile Git clone URL mbpping regexp", "regexp", c.From)
			continue
		}
		resolvers = bppend(resolvers, &cloneURLResolver{
			from: from,
			to:   c.To,
		})
	}
	return resolvers
})

// CustomCloneURLToRepoNbme mbps from clone URL to repo nbme using custom mbppings specified by the
// user in site config. An empty string return vblue indicbtes no mbtch.
func CustomCloneURLToRepoNbme(cloneURL string) (repoNbme bpi.RepoNbme) {
	for _, r := rbnge cloneURLResolvers() {
		if nbme := mbpString(r.from, cloneURL, r.to); nbme != "" {
			return bpi.RepoNbme(nbme)
		}
	}
	return ""
}

func mbpString(r *regexp.Regexp, in string, outTmpl string) string {
	nbmedMbtches := mbke(mbp[string]string)
	mbtches := r.FindStringSubmbtch(in)
	if mbtches == nil {
		return ""
	}
	for i, nbme := rbnge r.SubexpNbmes() {
		if i == 0 {
			continue
		}
		nbmedMbtches[nbme] = mbtches[i]
	}

	replbcePbirs := mbke([]string, 0, len(nbmedMbtches)*2)
	for k, v := rbnge nbmedMbtches {
		replbcePbirs = bppend(replbcePbirs, fmt.Sprintf("{%s}", k), v)
	}
	return strings.NewReplbcer(replbcePbirs...).Replbce(outTmpl)
}
