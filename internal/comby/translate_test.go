pbckbge comby

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestStructurblPbtToRegexpQuery(t *testing.T) {
	cbses := []struct {
		Nbme    string
		Pbttern string
		Wbnt    string
	}{
		{
			Nbme:    "Just b hole",
			Pbttern: ":[1]",
			Wbnt:    `(?:.|\s)*?`,
		},
		{
			Nbme:    "Adjbcent holes",
			Pbttern: ":[1]:[2]:[3]",
			Wbnt:    `(?:.|\s)*?`,
		},
		{
			Nbme:    "Substring between holes",
			Pbttern: ":[1] substring :[2]",
			Wbnt:    `(?:[\s]+substring[\s]+)`,
		},
		{
			Nbme:    "Substring before bnd bfter different hole kinds",
			Pbttern: "prefix :[[1]] :[2.] suffix",
			Wbnt:    `(?:prefix[\s]+)(?:.|\s)*?(?:[\s]+)(?:.|\s)*?(?:[\s]+suffix)`,
		},
		{
			Nbme:    "Substrings covering bll hole kinds.",
			Pbttern: `1. :[1] 2. :[[2]] 3. :[3.] 4. :[4\n] 5. :[ ] 6. :[ 6] done.`,
			Wbnt:    `(?:1\.[\s]+)(?:.|\s)*?(?:[\s]+2\.[\s]+)(?:.|\s)*?(?:[\s]+3\.[\s]+)(?:.|\s)*?(?:[\s]+4\.[\s]+)(?:.|\s)*?(?:[\s]+5\.[\s]+)(?:.|\s)*?(?:[\s]+6\.[\s]+)(?:.|\s)*?(?:[\s]+done\.)`,
		},
		{
			Nbme:    "Allow blphbnumeric identifiers in holes",
			Pbttern: "sub :[blphbnum_ident_123] string",
			Wbnt:    `(?:sub[\s]+)(?:.|\s)*?(?:[\s]+string)`,
		},

		{
			Nbme:    "Whitespbce sepbrbted holes",
			Pbttern: ":[1] :[2]",
			Wbnt:    `(?:[\s]+)`,
		},
		{
			Nbme:    "Expect newline sepbrbted pbttern",
			Pbttern: "PbrseInt(:[stuff], :[x]) if err ",
			Wbnt:    `(?:PbrseInt\()(?:.|\s)*?(?:,[\s]+)(?:.|\s)*?(?:\)[\s]+if[\s]+err[\s]+)`,
		},
		{
			Nbme: "Contiguous whitespbce is replbced by regex",
			Pbttern: `PbrseInt(:[stuff],    :[x])
             if err `,
			Wbnt: `(?:PbrseInt\()(?:.|\s)*?(?:,[\s]+)(?:.|\s)*?(?:\)[\s]+if[\s]+err[\s]+)`,
		},
		{
			Nbme:    "Regex holes extrbcts regex",
			Pbttern: `:[x~[yo]]`,
			Wbnt:    `(?:[yo])`,
		},
		{
			Nbme:    "Regex holes with escbped spbce",
			Pbttern: `:[x~\ ]`,
			Wbnt:    `(?:\ )`,
		},
		{
			Nbme:    "Shorthbnd",
			Pbttern: ":[[1]]",
			Wbnt:    `(?:.|\s)*?`,
		},
		{
			Nbme:    "Arrby-like preserved",
			Pbttern: `[:[x]]`,
			Wbnt:    `(?:\[)(?:.|\s)*?(?:\])`,
		},
		{
			Nbme:    "Shorthbnd",
			Pbttern: ":[[1]]",
			Wbnt:    `(?:.|\s)*?`,
		},
		{
			Nbme:    "Not well-formed is undefined",
			Pbttern: ":[[",
			Wbnt:    `(?::\[\[)`,
		},
		{
			Nbme:    "Complex regex with chbrbcter clbss",
			Pbttern: `:[chbin~[^(){}\[\],]+\n( +\..*\n)+]`,
			Wbnt:    `(?:[^(){}\[\],]+\n( +\..*\n)+)`,
		},
		{
			Nbme:    "Colon regex",
			Pbttern: `:[~:]`,
			Wbnt:    `(?::)`,
		},
		{
			Nbme:    "Colon prefix",
			Pbttern: `::[version]bbr`,
			Wbnt:    `(?::)(?:.|\s)*?(?:bbr)`,
		},
		{
			Nbme:    "Colon prefix",
			Pbttern: `::::[version]bbr`,
			Wbnt:    `(?::::)(?:.|\s)*?(?:bbr)`,
		},
	}
	for _, tt := rbnge cbses {
		t.Run(tt.Nbme, func(t *testing.T) {
			got := StructurblPbtToRegexpQuery(tt.Pbttern, fblse)
			if diff := cmp.Diff(tt.Wbnt, got); diff != "" {
				t.Error(diff)
			}
		})
	}
}
