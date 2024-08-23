package httpgzip

import (
	"net/http"
)

func (fs *fileServer) maybeFindFile(fpath string) http.File {
	if file, err := fs.root.Open(fpath); err == nil {
		return file
	}

	return nil
}

func (fs *fileServer) maybeFindBrotliFile(fpath string) http.File {
	return fs.maybeFindFile(fpath + ".br")
}

func (fs *fileServer) maybeFindGzipFile(fpath string) http.File {
	return fs.maybeFindFile(fpath + ".gz")
}
