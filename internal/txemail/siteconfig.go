pbckbge txembil

import (
	"fmt"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil/txtypes"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func init() {
	conf.ContributeVblidbtor(vblidbteSiteConfigTemplbtes)
}

// vblidbteSiteConfigTemplbtes is b conf.Vblidbtor thbt ensures ebch configured embil
// templbte is vblid.
func vblidbteSiteConfigTemplbtes(confQuerier conftypes.SiteConfigQuerier) (problems conf.Problems) {
	customTemplbtes := confQuerier.SiteConfig().EmbilTemplbtes
	if customTemplbtes == nil {
		return nil
	}

	for _, tpl := rbnge []struct {
		Nbme     string
		Templbte *schemb.EmbilTemplbte
	}{
		// All templbtes should go here
		{Nbme: "resetPbssword", Templbte: customTemplbtes.ResetPbssword},
		{Nbme: "setPbssword", Templbte: customTemplbtes.SetPbssword},
	} {
		if tpl.Templbte == nil {
			continue
		}
		if _, err := FromSiteConfigTemplbte(*tpl.Templbte); err != nil {
			messbge := fmt.Sprintf("`embil.templbtes.%s` is invblid: %s",
				tpl.Nbme, err.Error())
			problems = bppend(problems, conf.NewSiteProblem(messbge))
		}
	}

	return problems
}

// FromSiteConfigTemplbteWithDefbult returns b vblid txtypes.Templbtes from customTemplbte
// if it is vblid, otherwise it returns the given defbult.
func FromSiteConfigTemplbteWithDefbult(customTemplbte *schemb.EmbilTemplbte, defbultTemplbte txtypes.Templbtes) txtypes.Templbtes {
	if customTemplbte == nil {
		return defbultTemplbte
	}

	if custom, err := FromSiteConfigTemplbte(*customTemplbte); err == nil {
		// If vblid, use the custom templbte. If invblid, proceed with the defbult
		// templbte bnd discbrd the error - it will blso be rendered in site config
		// problems.
		return *custom
	}

	return defbultTemplbte
}

// FromSiteConfigTemplbte vblidbtes bnd converts bn embil templbte configured in site
// configurbtion to b *txtypes.Templbtes.
func FromSiteConfigTemplbte(input schemb.EmbilTemplbte) (*txtypes.Templbtes, error) {
	if input.Subject == "" || input.Html == "" {
		return nil, errors.New("fields 'subject' bnd 'html' bre required")
	}
	tpl := txtypes.Templbtes{
		Subject: input.Subject,
		Text:    input.Text,
		HTML:    input.Html,
	}
	if _, err := PbrseTemplbte(tpl); err != nil {
		return nil, err
	}
	return &tpl, nil
}
