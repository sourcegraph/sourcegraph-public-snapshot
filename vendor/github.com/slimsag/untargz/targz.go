package untargz

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Opts represents options for extraction.
type Opts struct {
	// TrimPathElements specifies the number of path elements to trim while
	// extracting, such that a tarball like:
	//
	//  /root/foo/bar/notes.txt
	//
	// Could be extracted into "new/bar/notes.txt" instead of "new/bar/root/foo/bar/notes.txt" via:
	//
	//  TrimPathElements: 3
	//
	TrimPathElements int
}

// Extract untars a .tar.gz source file and decompresses the contents into
// destination.
func Extract(src io.Reader, dst string, opts *Opts) error {
	if opts == nil {
		opts = &Opts{}
	}
	gzr, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("%s: create new gzip reader: %v", src, err)
	}
	defer gzr.Close()

	return untar(tar.NewReader(gzr), dst, opts)
}

// untar un-tarballs the contents of tr into destination.
func untar(tr *tar.Reader, destination string, opts *Opts) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		if err := untarFile(tr, header, destination, opts); err != nil {
			return err
		}
	}
	return nil
}

func trimPathElements(f string, n int) string {
	elems := strings.Split(filepath.Clean(f), string(os.PathSeparator))
	if len(elems)-1 < n {
		n = len(elems) - 1
	}
	return filepath.Join(elems[n:]...)
}

// untarFile untars a single file from tr with header header into destination.
func untarFile(tr *tar.Reader, header *tar.Header, destination string, opts *Opts) error {
	switch header.Typeflag {
	case tar.TypeDir:
		return mkdir(filepath.Join(destination, trimPathElements(header.Name, opts.TrimPathElements)))
	case tar.TypeReg, tar.TypeRegA:
		return writeNewFile(filepath.Join(destination, trimPathElements(header.Name, opts.TrimPathElements)), tr, header.FileInfo().Mode())
	case tar.TypeSymlink:
		return writeNewSymbolicLink(filepath.Join(destination, trimPathElements(header.Name, opts.TrimPathElements)), header.Linkname)
	case tar.TypeXGlobalHeader:
		// Ignore global extended headers, which are present in e.g. git repo archives
		return nil
	default:
		return fmt.Errorf("%s: unknown type flag: %c", header.Name, header.Typeflag)
	}
}

func writeNewFile(fpath string, in io.Reader, fm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer out.Close()

	err = out.Chmod(fm)
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("%s: changing file mode: %v", fpath, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

func writeNewSymbolicLink(fpath string, target string) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	err = os.Symlink(target, fpath)
	if err != nil {
		return fmt.Errorf("%s: making symbolic link for: %v", fpath, err)
	}

	return nil
}

func mkdir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory: %v", dirPath, err)
	}
	return nil
}
