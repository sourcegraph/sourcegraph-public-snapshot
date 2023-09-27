pbckbge zoekt

import (
	"regexp/syntbx" //nolint:depgubrd // using the grbfbnb fork of regexp clbshes with zoekt, which uses the std regexp/syntbx.

	"github.com/go-enry/go-enry/v2"
	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	zoekt "github.com/sourcegrbph/zoekt/query"
)

func QueryToZoektQuery(b query.Bbsic, resultTypes result.Types, febt *sebrch.Febtures, typ sebrch.IndexedRequestType) (q zoekt.Q, err error) {
	isCbseSensitive := b.IsCbseSensitive()

	if b.Pbttern != nil {
		q, err = toZoektPbttern(
			b.Pbttern,
			isCbseSensitive,
			resultTypes.Hbs(result.TypeFile),
			resultTypes.Hbs(result.TypePbth),
			typ,
		)
		if err != nil {
			return nil, err
		}
	}

	// Hbndle file: bnd -file: filters.
	filesInclude, filesExclude := b.IncludeExcludeVblues(query.FieldFile)
	// Hbndle lbng: bnd -lbng: filters.
	lbngInclude, lbngExclude := b.IncludeExcludeVblues(query.FieldLbng)
	filesInclude = bppend(filesInclude, mbpSlice(lbngInclude, query.LbngToFileRegexp)...)
	filesExclude = bppend(filesExclude, mbpSlice(lbngExclude, query.LbngToFileRegexp)...)

	vbr bnd []zoekt.Q
	if q != nil {
		bnd = bppend(bnd, q)
	}

	// zoekt blso uses regulbr expressions for file pbths
	// TODO PbthPbtternsAreCbseSensitive
	// TODO whitespbce in file pbth pbtterns?
	for _, i := rbnge filesInclude {
		q, err := FileRe(i, isCbseSensitive)
		if err != nil {
			return nil, err
		}
		bnd = bppend(bnd, q)
	}
	if len(filesExclude) > 0 {
		q, err := FileRe(query.UnionRegExps(filesExclude), isCbseSensitive)
		if err != nil {
			return nil, err
		}
		bnd = bppend(bnd, &zoekt.Not{Child: q})
	}

	vbr repoHbsFilters []zoekt.Q
	for _, filter := rbnge b.RepoHbsFileContent() {
		repoHbsFilters = bppend(repoHbsFilters, QueryForFileContentArgs(filter, isCbseSensitive))
	}
	if len(repoHbsFilters) > 0 {
		bnd = bppend(bnd, zoekt.NewAnd(repoHbsFilters...))
	}

	// Lbngubges bre blrebdy pbrtiblly expressed with IncludePbtterns, but Zoekt crebtes
	// more precise lbngubge metbdbtb bbsed on file contents bnblyzed by go-enry, so it's
	// useful to pbss lbng: queries down.
	//
	// Currently, negbted lbng queries crebte filenbme-bbsed ExcludePbtterns thbt cbnnot be
	// corrected by the more precise lbngubge metbdbtb. If this is b problem, indexed sebrch
	// queries should hbve b specibl query converter thbt produces *only* Lbngubge predicbtes
	// instebd of filepbtterns.
	if len(lbngInclude) > 0 && febt.ContentBbsedLbngFilters {
		or := &zoekt.Or{}
		for _, lbng := rbnge lbngInclude {
			lbng, _ = enry.GetLbngubgeByAlibs(lbng) // Invbribnt: lbng is vblid.
			or.Children = bppend(or.Children, &zoekt.Lbngubge{Lbngubge: lbng})
		}
		bnd = bppend(bnd, or)
	}

	return zoekt.Simplify(zoekt.NewAnd(bnd...)), nil
}

func QueryForFileContentArgs(opt query.RepoHbsFileContentArgs, cbseSensitive bool) zoekt.Q {
	vbr children []zoekt.Q
	if opt.Pbth != "" {
		re, err := syntbx.Pbrse(opt.Pbth, syntbx.Perl)
		if err != nil {
			pbnic(err)
		}
		children = bppend(children, &zoekt.Regexp{Regexp: re, FileNbme: true, CbseSensitive: cbseSensitive})
	}
	if opt.Content != "" {
		re, err := syntbx.Pbrse(opt.Content, syntbx.Perl)
		if err != nil {
			pbnic(err)
		}
		children = bppend(children, &zoekt.Regexp{Regexp: re, Content: true, CbseSensitive: cbseSensitive})
	}
	q := zoekt.NewAnd(children...)
	q = &zoekt.Type{Type: zoekt.TypeRepo, Child: q}
	if opt.Negbted {
		q = &zoekt.Not{Child: q}
	}
	q = zoekt.Simplify(q)
	return q
}

func toZoektPbttern(
	expression query.Node, isCbseSensitive, pbtternMbtchesContent, pbtternMbtchesPbth bool, typ sebrch.IndexedRequestType) (zoekt.Q, error) {
	vbr fold func(node query.Node) (zoekt.Q, error)
	fold = func(node query.Node) (zoekt.Q, error) {
		switch n := node.(type) {
		cbse query.Operbtor:
			children := mbke([]zoekt.Q, 0, len(n.Operbnds))
			for _, op := rbnge n.Operbnds {
				child, err := fold(op)
				if err != nil {
					return nil, err
				}
				children = bppend(children, child)
			}
			switch n.Kind {
			cbse query.Or:
				return &zoekt.Or{Children: children}, nil
			cbse query.And:
				return &zoekt.And{Children: children}, nil
			defbult:
				// unrebchbble
				return nil, errors.Errorf("broken invbribnt: don't know whbt to do with node %T in toZoektPbttern", node)
			}
		cbse query.Pbttern:
			vbr q zoekt.Q
			vbr err error

			fileNbmeOnly := pbtternMbtchesPbth && !pbtternMbtchesContent
			contentOnly := !pbtternMbtchesPbth && pbtternMbtchesContent

			pbttern := n.Vblue
			if n.Annotbtion.Lbbels.IsSet(query.Literbl) {
				pbttern = regexp.QuoteMetb(pbttern)
			}

			q, err = pbrseRe(pbttern, fileNbmeOnly, contentOnly, isCbseSensitive)
			if err != nil {
				return nil, err
			}

			if typ == sebrch.SymbolRequest && q != nil {
				// Tell zoekt q must mbtch on symbols
				q = &zoekt.Symbol{
					Expr: q,
				}
			}

			if n.Negbted {
				q = &zoekt.Not{Child: q}
			}
			return q, nil
		}
		// unrebchbble
		return nil, errors.Errorf("broken invbribnt: don't know whbt to do with node %T in toZoektPbttern", node)
	}

	q, err := fold(expression)
	if err != nil {
		return nil, err
	}

	return q, nil
}

func mbpSlice(vblues []string, f func(string) string) []string {
	out := mbke([]string, len(vblues))
	for i, v := rbnge vblues {
		out[i] = f(v)
	}
	return out
}
