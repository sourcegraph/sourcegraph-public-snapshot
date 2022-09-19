package workspace

import "os"

// MakeTempFile defaults to makeTemporaryFile and can be replaced for testing
// with determinstic workspace/scripts directories.
var MakeTempFile = makeTemporaryFile

func makeTemporaryFile(prefix string) (*os.File, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return nil, err
		}
		return os.CreateTemp(tempdir, prefix+"-*")
	}

	return os.CreateTemp("", prefix+"-*")
}

// MakeTempDirectory defaults to makeTemporaryDirectory and can be replaced for testing
// with determinstic workspace/scripts directories.
var MakeTempDirectory = MakeTemporaryDirectory

func MakeTemporaryDirectory(prefix string) (string, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return "", err
		}
		return os.MkdirTemp(tempdir, prefix+"-*")
	}

	return os.MkdirTemp("", prefix+"-*")
}
