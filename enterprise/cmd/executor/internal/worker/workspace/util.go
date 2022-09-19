package workspace

import "os"

// makeTempFile defaults to makeTemporaryFile and can be replaced for testing
// with determinstic workspace/scripts directories.
var makeTempFile = makeTemporaryFile

func makeTemporaryFile(prefix string) (*os.File, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return nil, err
		}
		return os.CreateTemp(tempdir, prefix+"-*")
	}

	return os.CreateTemp("", prefix+"-*")
}

// makeTempDirectory defaults to makeTemporaryDirectory and can be replaced for testing
// with determinstic workspace/scripts directories.
var makeTempDirectory = makeTemporaryDirectory

func makeTemporaryDirectory(prefix string) (string, error) {
	if tempdir := os.Getenv("TMPDIR"); tempdir != "" {
		if err := os.MkdirAll(tempdir, os.ModePerm); err != nil {
			return "", err
		}
		return os.MkdirTemp(tempdir, prefix+"-*")
	}

	return os.MkdirTemp("", prefix+"-*")
}
