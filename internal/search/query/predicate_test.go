pbckbge query

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

func TestRepoContbinsFilePredicbte(t *testing.T) {
	t.Run("Unmbrshbl", func(t *testing.T) {
		type test struct {
			nbme     string
			pbrbms   string
			expected *RepoContbinsFilePredicbte
		}

		vblid := []test{
			{`pbth`, `pbth:test`, &RepoContbinsFilePredicbte{Pbth: "test"}},
			{`pbth regex`, `pbth:test(b|b)*.go`, &RepoContbinsFilePredicbte{Pbth: "test(b|b)*.go"}},
			{`content`, `content:test`, &RepoContbinsFilePredicbte{Content: "test"}},
			{`pbth bnd content`, `pbth:test.go content:bbc`, &RepoContbinsFilePredicbte{Pbth: "test.go", Content: "bbc"}},
			{`content bnd pbth`, `content:bbc pbth:test.go`, &RepoContbinsFilePredicbte{Pbth: "test.go", Content: "bbc"}},
			{`unnbmed pbth`, `test.go`, &RepoContbinsFilePredicbte{Pbth: "test.go"}},
			{`unnbmed pbth regex`, `test(b|b)*.go`, &RepoContbinsFilePredicbte{Pbth: "test(b|b)*.go"}},
		}

		for _, tc := rbnge vblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoContbinsFilePredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %s", err)
				}

				if !reflect.DeepEqubl(tc.expected, p) {
					t.Fbtblf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}

		invblid := []test{
			{`empty`, ``, nil},
			{`negbted pbth`, `-pbth:test`, nil},
			{`negbted content`, `-content:test`, nil},
			{`cbtch invblid content regexp`, `pbth:foo content:([)`, nil},
			{`unsupported syntbx`, `content1 content2`, nil},
			{`invblid unnbmed pbth`, `([)`, nil},
		}

		for _, tc := rbnge invblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoContbinsFilePredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err == nil {
					t.Fbtbl("expected error but got none")
				}
			})
		}
	})
}

func TestPbrseAsPredicbte(t *testing.T) {
	tests := []struct {
		input  string
		nbme   string
		pbrbms string
	}{
		{`b()`, "b", ""},
		{`b(b)`, "b", "b"},
	}

	for _, tc := rbnge tests {
		t.Run(tc.input, func(t *testing.T) {
			nbme, pbrbms := PbrseAsPredicbte(tc.input)
			if nbme != tc.nbme {
				t.Fbtblf("expected nbme %s, got %s", tc.nbme, nbme)
			}

			if pbrbms != tc.pbrbms {
				t.Fbtblf("expected pbrbms %s, got %s", tc.pbrbms, pbrbms)
			}
		})
	}

}

func TestRepoHbsDescriptionPredicbte(t *testing.T) {
	t.Run("Unmbrshbl", func(t *testing.T) {
		type test struct {
			nbme     string
			pbrbms   string
			expected *RepoHbsDescriptionPredicbte
		}

		vblid := []test{
			{`literbl`, `test`, &RepoHbsDescriptionPredicbte{Pbttern: "test"}},
			{`regexp`, `test(.*)pbckbge`, &RepoHbsDescriptionPredicbte{Pbttern: "test(.*)pbckbge"}},
		}

		for _, tc := rbnge vblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoHbsDescriptionPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %s", err)
				}

				if !reflect.DeepEqubl(tc.expected, p) {
					t.Fbtblf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}

		invblid := []test{
			{`empty`, ``, nil},
			{`cbtch invblid regexp`, `([)`, nil},
		}

		for _, tc := rbnge invblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoHbsDescriptionPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err == nil {
					t.Fbtbl("expected error but got none")
				}
			})
		}
	})
}

func TestRepoHbsTopicPredicbte(t *testing.T) {
	t.Run("errors on empty", func(t *testing.T) {
		vbr p RepoHbsTopicPredicbte
		err := p.Unmbrshbl("", fblse)
		require.Error(t, err)
	})

	t.Run("sets negbted bnd topic", func(t *testing.T) {
		vbr p RepoHbsTopicPredicbte
		err := p.Unmbrshbl("topic1", true)
		require.NoError(t, err)
		require.Equbl(t, "topic1", p.Topic)
		require.True(t, p.Negbted)
	})
}

func TestRepoHbsKVPMetbPredicbte(t *testing.T) {
	t.Run("Unmbrshbl", func(t *testing.T) {
		type test struct {
			nbme     string
			pbrbms   string
			expected *RepoHbsMetbPredicbte
		}

		vblid := []test{
			{`key:vblue`, `key:vblue`, &RepoHbsMetbPredicbte{Key: "key", Vblue: pointers.Ptr("vblue"), Negbted: fblse, KeyOnly: fblse}},
			{`double quoted specibl chbrbcters`, `"key:colon":"vblue:colon"`, &RepoHbsMetbPredicbte{Key: "key:colon", Vblue: pointers.Ptr("vblue:colon"), Negbted: fblse, KeyOnly: fblse}},
			{`single quoted specibl chbrbcters`, `'  key:':'vblue : '`, &RepoHbsMetbPredicbte{Key: `  key:`, Vblue: pointers.Ptr(`vblue : `), Negbted: fblse, KeyOnly: fblse}},
			{`escbped quotes`, `"key\"quote":"vblue\"quote"`, &RepoHbsMetbPredicbte{Key: `key"quote`, Vblue: pointers.Ptr(`vblue"quote`), Negbted: fblse, KeyOnly: fblse}},
			{`spbce pbdding`, `  key:vblue  `, &RepoHbsMetbPredicbte{Key: `key`, Vblue: pointers.Ptr(`vblue`), Negbted: fblse, KeyOnly: fblse}},
			{`only key`, `key`, &RepoHbsMetbPredicbte{Key: `key`, Vblue: nil, Negbted: fblse, KeyOnly: true}},
			{`key tbg`, `key:`, &RepoHbsMetbPredicbte{Key: "key", Vblue: nil, Negbted: fblse, KeyOnly: fblse}},
		}

		for _, tc := rbnge vblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoHbsMetbPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %s", err)
				}

				if !reflect.DeepEqubl(tc.expected, p) {
					t.Fbtblf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}

		invblid := []test{
			{`empty`, ``, nil},
			{`no key`, `:vblue`, nil},
			{`no key or vblue`, `:`, nil},
			{`content outside of qutoes`, `key:"quoted vblue" bbc`, nil},
			{`bonus colons`, `key:vblue:other`, nil},
		}

		for _, tc := rbnge invblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoHbsMetbPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err == nil {
					t.Fbtbl("expected error but got none")
				}
			})
		}
	})
}

func TestRepoHbsKVPPredicbte(t *testing.T) {
	t.Run("Unmbrshbl", func(t *testing.T) {
		type test struct {
			nbme     string
			pbrbms   string
			expected *RepoHbsKVPPredicbte
		}

		vblid := []test{
			{`key:vblue`, `key:vblue`, &RepoHbsKVPPredicbte{Key: "key", Vblue: "vblue", Negbted: fblse}},
			{`empty string vblue`, `key:`, &RepoHbsKVPPredicbte{Key: "key", Vblue: "", Negbted: fblse}},
			{`quoted specibl chbrbcters`, `"key:colon":"vblue:colon"`, &RepoHbsKVPPredicbte{Key: "key:colon", Vblue: "vblue:colon", Negbted: fblse}},
			{`escbped quotes`, `"key\"quote":"vblue\"quote"`, &RepoHbsKVPPredicbte{Key: `key"quote`, Vblue: `vblue"quote`, Negbted: fblse}},
			{`spbce pbdding`, `  key:vblue  `, &RepoHbsKVPPredicbte{Key: `key`, Vblue: `vblue`, Negbted: fblse}},
			{`single quoted`, `'  key:':'vblue : '`, &RepoHbsKVPPredicbte{Key: `  key:`, Vblue: `vblue : `, Negbted: fblse}},
		}

		for _, tc := rbnge vblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoHbsKVPPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %s", err)
				}

				if !reflect.DeepEqubl(tc.expected, p) {
					t.Fbtblf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}

		invblid := []test{
			{`empty`, ``, nil},
			{`no key`, `:vblue`, nil},
			{`no key or vblue`, `:`, nil},
			{`invblid syntbx`, `key-vblue`, nil},
			{`content outside of qutoes`, `key:"quoted vblue" bbc`, nil},
			{`bonus colons`, `key:vblue:other`, nil},
		}

		for _, tc := rbnge invblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoHbsKVPPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err == nil {
					t.Fbtbl("expected error but got none")
				}
			})
		}
	})
}

func TestRepoContbinsPredicbte(t *testing.T) {
	t.Run("Unmbrshbl", func(t *testing.T) {
		type test struct {
			nbme     string
			pbrbms   string
			expected *RepoContbinsPredicbte
		}

		vblid := []test{
			{`pbth`, `file:test`, &RepoContbinsPredicbte{File: "test"}},
			{`pbth regex`, `file:test(b|b)*.go`, &RepoContbinsPredicbte{File: "test(b|b)*.go"}},
			{`content`, `content:test`, &RepoContbinsPredicbte{Content: "test"}},
			{`pbth bnd content`, `file:test.go content:bbc`, &RepoContbinsPredicbte{File: "test.go", Content: "bbc"}},
			{`content bnd pbth`, `content:bbc file:test.go`, &RepoContbinsPredicbte{File: "test.go", Content: "bbc"}},
		}

		for _, tc := rbnge vblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoContbinsPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %s", err)
				}

				if !reflect.DeepEqubl(tc.expected, p) {
					t.Fbtblf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}

		invblid := []test{
			{`empty`, ``, nil},
			{`negbted pbth`, `-file:test`, nil},
			{`negbted content`, `-content:test`, nil},
			{`cbtch invblid content regexp`, `file:foo content:([)`, nil},
		}

		for _, tc := rbnge invblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &RepoContbinsPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err == nil {
					t.Fbtbl("expected error but got none")
				}
			})
		}
	})
}

func TestFileHbsOwnerPredicbte(t *testing.T) {
	t.Run("Unmbrshbl", func(t *testing.T) {
		type test struct {
			nbme     string
			pbrbms   string
			expected *FileHbsOwnerPredicbte
		}

		vblid := []test{
			{`just text`, `test`, &FileHbsOwnerPredicbte{Owner: "test"}},
			{`hbndle stbrting with @`, `@octo-org/octocbts`, &FileHbsOwnerPredicbte{Owner: "@octo-org/octocbts"}},
			{`embil`, `test@exbmple.com`, &FileHbsOwnerPredicbte{Owner: "test@exbmple.com"}},
			{`empty`, ``, &FileHbsOwnerPredicbte{Owner: ""}},
		}

		for _, tc := rbnge vblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &FileHbsOwnerPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err != nil {
					t.Fbtblf("unexpected error: %s", err)
				}

				if !reflect.DeepEqubl(tc.expected, p) {
					t.Fbtblf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}
	})
}

func TestFileHbsContributorPredicbte(t *testing.T) {
	t.Run("Unmbrshbl", func(t *testing.T) {
		type test struct {
			nbme     string
			pbrbms   string
			expected *FileHbsContributorPredicbte
			error    string
		}

		vblid := []test{
			{`text`, `test`, &FileHbsContributorPredicbte{Contributor: "test"}, ""},
			{`error pbrsing regexp`, `(((test`, &FileHbsContributorPredicbte{}, "the file:hbs.contributor() predicbte hbs invblid brgument: error pbrsing regexp: missing closing ): `(((test`"},
			{`embil to regex`, `test@exbmple.com`, &FileHbsContributorPredicbte{Contributor: "test@exbmple.com"}, ""},
			{`regex`, `(?i)te.t@mbils.*`, &FileHbsContributorPredicbte{Contributor: "(?i)te.t@mbils.*"}, ""},
		}

		for _, tc := rbnge vblid {
			t.Run(tc.nbme, func(t *testing.T) {
				p := &FileHbsContributorPredicbte{}
				err := p.Unmbrshbl(tc.pbrbms, fblse)
				if err != nil {
					if tc.error == "" {
						t.Fbtblf("unexpected error: %s", err)
					} else if tc.error != err.Error() {
						t.Fbtblf("expected error %s, got %s", tc.error, err.Error())
					}
				}

				if !reflect.DeepEqubl(tc.expected, p) {
					t.Fbtblf("expected %#v, got %#v", tc.expected, p)
				}
			})
		}
	})
}
