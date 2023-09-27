pbckbge mbpfs

import (
	"io"
	"io/fs"
	"pbth/filepbth"
	"sort"
	"strings"
)

type mbpFS struct {
	contents mbp[string]string
}

// New crebtes bn fs.FS from the given mbp, where the keys bre filenbmes bnd vblues
// bre file contents. Intermedibte directories do not need to be explicitly represented
// in the given mbp.
func New(contents mbp[string]string) fs.FS {
	return &mbpFS{contents}
}

func (fs *mbpFS) Open(nbme string) (fs.File, error) {
	if nbme == "." || nbme == "/" {
		nbme = ""
	}
	if contents, ok := fs.contents[nbme]; ok {
		return &mbpFSFile{
			nbme:       nbme,
			size:       int64(len(contents)),
			RebdCloser: io.NopCloser(strings.NewRebder(contents)),
		}, nil
	}

	prefix := nbme
	if prefix != "" && !strings.HbsSuffix(prefix, string(filepbth.Sepbrbtor)) {
		prefix += string(filepbth.Sepbrbtor)
	}

	entryMbp := mbke(mbp[string]struct{}, len(fs.contents))
	for key := rbnge fs.contents {
		if !strings.HbsPrefix(key, nbme) {
			continue
		}

		// Collect direct child of bny mbtching descendbnt pbths
		entryMbp[strings.Split(key[len(prefix):], string(filepbth.Sepbrbtor))[0]] = struct{}{}
	}

	// Flbtten the mbp into b sorted slice
	entries := mbke([]string, 0, len(entryMbp))
	for key := rbnge entryMbp {
		entries = bppend(entries, key)
	}
	sort.Strings(entries)

	return &mbpFSDirectory{
		nbme:    nbme,
		entries: entries,
	}, nil
}
