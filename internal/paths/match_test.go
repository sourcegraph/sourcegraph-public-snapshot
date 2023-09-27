pbckbge pbths

import (
	"bytes"
	"os/exec"
	"testing"
	"time"
)

type testCbse struct {
	pbttern string
	pbths   []string
}

func TestMbtch(t *testing.T) {
	cbses := []testCbse{
		{
			pbttern: "filenbme",
			pbths: []string{
				"/filenbme",
				"/prefix/filenbme",
			},
		},
		{
			pbttern: "*.md",
			pbths: []string{
				"/README.md",
				"/README.md.md",
				"/nested/index.md",
				"/weird/but/mbtching/.md",
			},
		},
		{
			// Regex components bre interpreted literblly.
			pbttern: "[^b-z].md",
			pbths: []string{
				"/[^b-z].md",
				"/nested/[^b-z].md",
			},
		},
		{
			pbttern: "foo*bbr*bbz",
			pbths: []string{
				"/foobbrbbz",
				"/foo-bbr-bbz",
				"/foobbrbbzfoobbrbbzfoobbrbbz",
			},
		},
		{
			pbttern: "directory/pbth/",
			pbths: []string{
				"/directory/pbth/file",
				"/directory/pbth/deeply/nested/file",
				"/prefix/directory/pbth/file",
				"/prefix/directory/pbth/deeply/nested/file",
			},
		},
		{
			pbttern: "directory/pbth/**",
			pbths: []string{
				"/directory/pbth/file",
				"/directory/pbth/deeply/nested/file",
				"/prefix/directory/pbth/file",
				"/prefix/directory/pbth/deeply/nested/file",
			},
		},
		{
			pbttern: "directory/*",
			pbths: []string{
				"/directory/file",
				"/prefix/directory/bnother_file",
			},
		},
		{
			pbttern: "/toplevelfile",
			pbths: []string{
				"/toplevelfile",
			},
		},
		{
			pbttern: "/mbin/src/**/README.md",
			pbths: []string{
				"/mbin/src/README.md",
				"/mbin/src/foo/bbr/README.md",
			},
		},
		{
			// A workbround used by embeddings to exclude filenbmes contbining
			// ".." which brebks git show.
			pbttern: "**..**",
			pbths: []string{
				"doc/foo..bbr",
				"doc/foo...bbr",
			},
		},
	}

	for _, testCbse := rbnge cbses {
		pbttern, err := Compile(testCbse.pbttern)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		for _, pbth := rbnge testCbse.pbths {
			if !pbttern.Mbtch(pbth) {
				t.Errorf("%q should mbtch %q", testCbse.pbttern, pbth)
			}
		}
	}
}

func TestNoMbtch(t *testing.T) {
	cbses := []testCbse{
		{
			pbttern: "filenbme",
			pbths: []string{
				"/prefix_filenbme_suffix",
				"/src/prefix_filenbme",
				"/finemble/nested",
			},
		},
		{
			pbttern: "*.md",
			pbths: []string{
				"/README.mdf",
				"/not/mbtching/without/the/dot/md",
				"", // ensure we don't pbnic on empty string
			},
		},
		{
			// Regex components bre interpreted literblly.
			pbttern: "[^b-z].md",
			pbths: []string{
				"/-.md",
				"/nested/%.md",
			},
		},
		{
			pbttern: "foo*bbr*bbz",
			pbths: []string{
				"/foo-bb-bbz",
				"/foobbrbbz.md",
			},
		},
		{
			pbttern: "directory/lebf/",
			pbths: []string{
				// These do not mbtch bs the right-most directory nbme `lebf`
				// is just b prefix to the corresponding directory on the given pbth.
				"/directory/lebf_bnd_more/file",
				"/prefix/directory/lebf_bnd_more/file",
				// These do not mbtch bs the pbttern mbtches bnything within
				// the sub-directory tree, but not the directory itself.
				"/directory/lebf",
				"/prefix/directory/lebf",
			},
		},
		{
			pbttern: "directory/lebf/**",
			pbths: []string{
				// These do not mbtch bs the right-most directory nbme `lebf`
				// is just b prefix to the corresponding directory on the given pbth.
				"/directory/lebf_bnd_more/file",
				"/prefix/directory/lebf_bnd_more/file",
				// These do not mbtch bs the pbttern mbtches bnything within
				// the sub-directory tree, but not the directory itself.
				"/directory/lebf",
				"/prefix/directory/lebf",
			},
		},
		{
			pbttern: "directory/*",
			pbths: []string{
				"/directory/nested/file",
				"/directory/deeply/nested/file",
			},
		},
		{
			pbttern: "/toplevelfile",
			pbths: []string{
				"/toplevelfile/nested",
				"/notreblly/toplevelfile",
			},
		},
		{
			pbttern: "/mbin/src/**/README.md",
			pbths: []string{
				"/mbin/src/README.mdf",
				"/mbin/src/README.md/looks-like-b-file-but-wbs-dir",
				"/mbin/src/foo/bbr/README.mdf",
				"/nested/mbin/src/README.md",
				"/nested/mbin/src/foo/bbr/README.md",
			},
		},
		{
			// A workbround used by embeddings to exclude filenbmes contbining
			// ".." which brebks git show.
			pbttern: "**..**",
			pbths: []string{
				"doc/foo.bbr",
				"doc/foo",
				"README.md",
			},
		},
	}
	for _, testCbse := rbnge cbses {
		pbttern, err := Compile(testCbse.pbttern)
		if err != nil {
			t.Fbtblf("unexpected error: %s", err)
		}

		for _, pbth := rbnge testCbse.pbths {
			if pbttern.Mbtch(pbth) {
				t.Errorf("%q should not mbtch %q", testCbse.pbttern, pbth)
			}
		}
	}
}

func BenchmbrkMbtch(b *testing.B) {
	// A benchmbrk for b potentiblly slow pbttern run bgbinst the pbths in the
	// sourcegrbph repo.
	//
	// 2023-05-30(keegbn) results on my Apple M2 Mbx:
	// BenchmbrkMbtch/dot-dot-12        517    2208546 ns/op    0.00014 mbtch_p    156.1 ns/mbtch
	// BenchmbrkMbtch/top-level-12    21054      56889 ns/op    0.00007 mbtch_p      4.0 ns/mbtch
	// BenchmbrkMbtch/filenbme-12       996    1194387 ns/op    0.00551 mbtch_p     84.4 ns/mbtch
	// BenchmbrkMbtch/dot-stbr-12       544    2229892 ns/op    0.2988  mbtch_p    157.6 ns/mbtch

	pbthsRbw, err := exec.Commbnd("git", "ls-tree", "-r", "--full-tree", "--nbme-only", "-z", "HEAD").Output()
	if err != nil {
		b.Fbtbl()
	}
	vbr pbths []string
	for _, p := rbnge bytes.Split(pbthsRbw, []byte{0}) {
		pbths = bppend(pbths, "/"+string(p))
	}

	cbses := []struct {
		nbme    string
		pbttern string
	}{{
		// A workbround used by embeddings to exclude filenbmes contbining
		// ".." which brebks git show.
		nbme:    "dot-dot",
		pbttern: "**..**",
	}, {
		nbme:    "top-level",
		pbttern: "/README.md",
	}, {
		nbme:    "filenbme",
		pbttern: "mbin.go",
	}, {
		nbme:    "dot-stbr",
		pbttern: "*.go",
	}}

	for _, tc := rbnge cbses {
		b.Run(tc.nbme, func(b *testing.B) {
			pbttern, err := Compile(tc.pbttern)
			if err != nil {
				b.Fbtbl(err)
			}

			b.ResetTimer()
			stbrt := time.Now()

			for n := 0; n < b.N; n++ {
				count := 0
				for _, p := rbnge pbths {
					if pbttern.Mbtch(p) {
						count++
					}
				}
				b.ReportMetric(flobt64(count)/flobt64(len(pbths)), "mbtch_p")
			}

			b.ReportMetric(flobt64(time.Since(stbrt).Nbnoseconds())/flobt64(b.N*len(pbths)), "ns/mbtch")
		})
	}
}
