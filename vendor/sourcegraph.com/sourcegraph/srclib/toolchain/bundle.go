package toolchain

import (
	"archive/tar"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/kr/fs"

	"sort"
)

// Variant represents a single permutation of variables that produces
// a single product (for toolchain bundles); e.g., {"os":"linux",
// "arch": "amd64"}.
type Variant map[string]string

func (v Variant) String() string {
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys) // sort for determinism and canonical representation

	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = key + "-" + v[key]
	}
	return strings.Join(parts, "_")
}

// ParseVariant parses a variant string (e.g., "arch-386_os-linux")
// into its key-value pairs ({"arch": "386", "os": "linux"}).
func ParseVariant(s string) Variant {
	if s == "" {
		return Variant{}
	}
	pairs := strings.Split(s, "_")
	m := make(map[string]string, len(pairs))
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "-", 2)
		if len(parts) == 1 {
			// Assume empty.
			parts = append(parts, "")
		}
		k := parts[0]
		v := parts[1]
		m[k] = v
	}
	return m
}

// Bundle builds all variants of the toolchain and creates
// separate archive files for each variant in a temporary
// directory. It returns a list of variants and the archive files
// produced from them.
//
// If variants is nil, all variants specified in the Srclibtoolchain
// file are built.
func Bundle(toolchainDir, outputDir string, variants []Variant, dryRun, verbose bool) (bundles []string, err error) {
	toolchainDir = filepath.Clean(toolchainDir)
	if toolchainDir == "." {
		toolchainDir, err = filepath.Abs(toolchainDir)
		if err != nil {
			return nil, err
		}
	}

	info := &Info{Dir: toolchainDir, ConfigFile: ConfigFilename}
	conf, err := info.ReadConfig()
	if err != nil {
		return nil, err
	}

	if conf.Bundle == nil {
		return nil, fmt.Errorf("toolchain at %s does not support bundling (%s file has no Bundle property)", toolchainDir, ConfigFilename)
	}

	if variants == nil {
		variants = conf.Bundle.Variants
	}

	toolchainName := filepath.Base(toolchainDir)
	for _, variant := range variants {
		archiveFile := filepath.Join(outputDir, toolchainName+"__bundle__"+variant.String()+".tar.gz")

		if !dryRun {
			if err := os.MkdirAll(filepath.Dir(archiveFile), 0700); err != nil {
				return nil, err
			}
		}

		if verbose {
			log.Printf("Creating bundle variant %v...", variant)
		}
		if err := createBundleVariant(info.Dir, archiveFile, conf, variant, dryRun, verbose); err != nil {
			return nil, err
		}

		bundles = append(bundles, archiveFile)

		log.Println()
		log.Println()
	}

	return bundles, nil
}

// createBundleVariant builds and archives a single variant of a
// toolchain.
func createBundleVariant(toolchainDir, archiveFile string, conf *Config, variant Variant, dryRun, verbose bool) (err error) {
	// Build.
	for _, shellCmd := range conf.Bundle.Commands {
		shellCmd = os.Expand(shellCmd, func(key string) string {
			v, present := variant[key]
			if present {
				return v
			}
			return fmt.Sprintf("${%s}", key) // remain uninterpolated, pass to `sh -c`
		})

		var buf bytes.Buffer
		cmd := exec.Command("sh", "-c", shellCmd)
		cmd.Dir = toolchainDir
		cmd.Env = append(os.Environ(), variantToEnvVars(variant)...)
		if verbose {
			log.Printf("\t> %s", shellCmd)
			cmd.Stdout = os.Stderr
			cmd.Stderr = os.Stderr
		} else {
			cmd.Stdout = &buf
			cmd.Stderr = &buf
		}
		if dryRun {
			continue
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("exec %v failed (%s)\n%s", cmd.Args, err, buf.String())
		}
	}

	// Create archive.
	log.Printf("\tCreating archive file %s", archiveFile)
	if dryRun {
		return nil
	}
	f, err := os.Create(archiveFile)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := f.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	gw := gzip.NewWriter(f)
	defer func() {
		if err2 := gw.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	tw := tar.NewWriter(gw)
	defer func() {
		if err2 := tw.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()

	for _, pathGlob := range conf.Bundle.Paths {
		paths, err := expandBundlePath(filepath.Join(toolchainDir, pathGlob))
		if err != nil {
			return err
		}
		for _, path := range paths {
			fi2, err := os.Stat(path)
			if err != nil {
				return err
			}

			if !fi2.Mode().IsRegular() && !fi2.Mode().IsDir() {
				log.Printf("Warning: while creating bundle archive for toolchain in %s (variant %v), encountered a file at %q that is neither file, dir nor symlink; skipping. (Special files are skipped, for security reasons.)", toolchainDir, variant, path)
				continue
			}

			relPath, err := filepath.Rel(toolchainDir, path)
			if err != nil {
				return err
			}

			var linkName string // only set for symlinks
			if fi2.Mode()&os.ModeSymlink != 0 {
				dest, err := os.Readlink(path)
				if err != nil {
					return err
				}
				linkName = dest
			}

			hdr, err := tar.FileInfoHeader(fi2, linkName)
			if err != nil {
				return err
			}
			if strings.HasSuffix(hdr.Name, "/") {
				relPath += "/"
			}
			hdr.Name = relPath

			// Ensure it's writable by us.
			hdr.Mode |= 0200

			if err := tw.WriteHeader(hdr); err != nil {
				return err
			}

			if fi2.Mode().IsRegular() {
				f2, err := os.Open(path)
				if err != nil {
					return err
				}
				if _, err := io.Copy(tw, f2); err != nil {
					f2.Close()
					return err
				}
				if err := f2.Close(); err != nil {
					return err
				}
			}
		}
	}

	return err
}

// expandBundlePath calls filepath.Glob first, then recursively adds
// descendents of any directories specified.
func expandBundlePath(pathGlob string) ([]string, error) {
	paths, err := filepath.Glob(pathGlob)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			return nil, err
		}

		if fi.Mode().IsDir() {
			w := fs.Walk(path)
			for w.Step() {
				if err := w.Err(); err != nil {
					return nil, err
				}
				paths = append(paths, w.Path())
			}
		}
	}

	return paths, nil
}

func variantToEnvVars(v Variant) []string {
	var s []string
	for k, v := range v {
		s = append(s, k+"="+v)
	}
	return s
}

// Unbundle unpacks the archive r to the toolchain named by the given
// toolchain path. The archiveName indicates what type of archive r
// contains; e.g., "foo.tar.gz" is a gzipped tar archive, "foo.tar" is
// just a tar archive, etc. An error is returned if archiveName's
// suffix isn't that of a recognized format.
func Unbundle(toolchainPath, archiveName string, r io.Reader) (err error) {
	ext := filepath.Ext(archiveName)
	switch ext {
	case ".tgz":
		return Unbundle(toolchainPath, strings.TrimSuffix(archiveName, ext)+".tar.gz", r)
	case ".gz":
		gr, err := gzip.NewReader(r)
		if err != nil {
			return err
		}
		defer func() {
			if err2 := gr.Close(); err2 != nil && err == nil {
				err = err2
			}
		}()
		return Unbundle(toolchainPath, strings.TrimSuffix(archiveName, ext), gr)

	case ".tbz2":
		return Unbundle(toolchainPath, strings.TrimSuffix(archiveName, ext)+".tar.bz2", r)
	case ".bz2":
		return Unbundle(toolchainPath, strings.TrimSuffix(archiveName, ext), bzip2.NewReader(r))

	case ".tar":

		dir, err := Dir(toolchainPath)
		if err != nil {
			return err
		}

		writeTarEntry := func(hdr *tar.Header, r io.Reader) (err error) {
			mode := hdr.FileInfo().Mode()

			name := filepath.Clean(hdr.Name)
			if strings.Contains(name, "..") {
				return fmt.Errorf("invalid tar archive entry path %q", name)
			}

			path := filepath.Join(dir, name)

			switch {
			case mode.IsDir():
				return os.MkdirAll(path, mode.Perm())

			case mode.IsRegular():
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					return err
				}

				// Check file perms to overwrite if needed.
				fi, err := os.Stat(path)
				if err == nil {
					perm := fi.Mode().Perm()
					if perm&0200 /* write */ == 0 {
						if err := os.Chmod(path, perm|0200); err != nil {
							return err
						}
					}
				}

				f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode.Perm())
				if err != nil {
					return err
				}
				defer func() {
					if err2 := f.Close(); err2 != nil && err == nil {
						err = err2
					}
				}()
				if _, err := io.Copy(f, r); err != nil {
					return err
				}

			case mode&os.ModeSymlink != 0:
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					return err
				}

				if err := os.Symlink(hdr.Linkname, path); err != nil {
					return err
				}

			default:
				log.Printf("Warning: while extracting bundle for toolchain %s to %s, encountered a tar entry at %q that is neither file, dir, nor symlink; skipping. (Special files are skipped, for security reasons.)", toolchainPath, dir, hdr.Name)

			}
			return nil
		}

		tr := tar.NewReader(r)
		for {
			hdr, err := tr.Next()
			if err == io.EOF {
				// end of tar archive
				break
			}
			if err != nil {
				return err
			}
			if err := writeTarEntry(hdr, tr); err != nil {
				return err
			}
		}
		return nil

	}
	return fmt.Errorf("bad toolchain bundle archive file %q (suffix not recognized)", archiveName)
}
