package pypi

import (
	"bytes"
	"io"
	"io/fs"

	"github.com/sourcegraph/sourcegraph/internal/unpack"
)

// Unpacker abstracts the various archive formats pypi offers, such as tar.gzip and wheel.
type Unpacker interface {
	Unpack(dstDir string) error
}

type wheelUnpacker struct {
	data []byte
}

func (k *wheelUnpacker) Unpack(dstDir string) error {
	// unzip the wheel into dstDir, skipping any files that are potentially
	// malicious.
	br := bytes.NewReader(k.data)
	return unpack.Zip(br, int64(br.Len()), dstDir, unpack.Opts{
		SkipInvalid: true,
		Filter: func(path string, file fs.FileInfo) bool {
			_, malicious := unpack.IsPotentiallyMaliciousFilepathInArchive(path, dstDir)
			return !malicious
		},
	})
}

type tarballUnpacker struct {
	rc io.ReadCloser
}

func (t tarballUnpacker) Unpack(dstDir string) error {
	return unpack.DecompressTgz(t.rc, dstDir)
}
