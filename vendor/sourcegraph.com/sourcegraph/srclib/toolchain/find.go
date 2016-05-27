package toolchain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"

	"github.com/kr/fs"
	"sourcegraph.com/sourcegraph/srclib"
)

// Lookup finds a toolchain by path in the SRCLIBPATH. For each DIR in
// SRCLIBPATH, it checks for the existence of DIR/PATH/Srclibtoolchain.
func Lookup(path string) (*Info, error) {
	path = filepath.Clean(path)

	dir, err := Dir(path)
	if err != nil {
		return nil, err
	}

	// Ensure it exists.
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsDir() {
		return nil, &os.PathError{Op: "toolchain.Lookup", Path: dir, Err: errors.New("not a directory")}
	}

	return newInfo(path, dir, ConfigFilename)
}

func lookupToolchain(toolchainPath string) (string, error) {
	matches, err := lookInPaths(filepath.Join(toolchainPath, ConfigFilename), srclib.Path)
	if err != nil {
		return "", err
	}
	if len(matches) > 1 {
		return "", fmt.Errorf("shadowed toolchain path %q (toolchains: %v)", toolchainPath, matches)
	}
	if len(matches) == 0 {
		return "", &os.PathError{Op: "lookupToolchain", Path: toolchainPath, Err: os.ErrNotExist}
	}
	return filepath.Dir(matches[0]), nil
}

// List finds all toolchains in the SRCLIBPATH.
//
// List does not find nested toolchains; i.e., if DIR is a toolchain
// dir (with a DIR/Srclibtoolchain file), then none of DIR's
// subdirectories are searched for toolchains.
func List() ([]*Info, error) {
	var found []*Info
	seen := map[string]string{}

	dirs := filepath.SplitList(srclib.Path)

	// maps symlinked trees to their original path
	origDirs := map[string]string{}

	for i := 0; i < len(dirs); i++ {
		dir := dirs[i]
		if dir == "" {
			dir = "."
		}
		w := fs.Walk(dir)
		for w.Step() {
			if err := w.Err(); err != nil {
				if w.Path() == dir && os.IsNotExist(err) {
					return nil, nil
				}
				return nil, w.Err()
			}
			fi := w.Stat()
			name := fi.Name()
			path := w.Path()
			if path != dir && (name[0] == '.' || name[0] == '_') {
				w.SkipDir()
			} else if fi.Mode()&os.ModeSymlink != 0 {
				// Check if symlink points to a directory.
				if sfi, err := os.Stat(path); err == nil {
					if !sfi.IsDir() {
						continue
					}
				} else if os.IsNotExist(err) {
					continue
				} else {
					return nil, err
				}

				targetPath, err := os.Readlink(path)
				if err != nil {
					return nil, err
				}

				if _, traversed := origDirs[targetPath]; !traversed {
					// traverse symlinks but refer to symlinked trees' toolchains using
					// the path to them through the original entry in SRCLIBPATH
					dirs = append(dirs, targetPath)
					origDirs[targetPath], _ = filepath.Rel(dir, path)
				}
			} else if fi.Mode().IsDir() {
				// Check for Srclibtoolchain file in this dir.

				if _, err := os.Stat(filepath.Join(path, ConfigFilename)); os.IsNotExist(err) {
					continue
				} else if err != nil {
					return nil, err
				}

				// Found a Srclibtoolchain file.
				path = filepath.Clean(path)

				var toolchainPath string
				if orig, present := origDirs[dir]; present {
					toolchainPath, _ = filepath.Rel(dir, path)
					if toolchainPath == "." {
						toolchainPath = ""
					}
					toolchainPath = orig + toolchainPath
				} else {
					toolchainPath, _ = filepath.Rel(dir, path)
				}
				toolchainPath = filepath.ToSlash(toolchainPath)

				if otherDir, seen := seen[toolchainPath]; seen {
					return nil, fmt.Errorf("saw 2 toolchains at path %s in dirs %s and %s", toolchainPath, otherDir, path)
				}
				seen[toolchainPath] = path

				info, err := newInfo(toolchainPath, path, ConfigFilename)
				if err != nil {
					return nil, err
				}
				found = append(found, info)

				// Disallow nested toolchains to speed up List. This
				// means that if DIR/Srclibtoolchain exists, no other
				// Srclibtoolchain files underneath DIR will be read.
				w.SkipDir()
			}
		}
	}
	return found, nil
}

func newInfo(toolchainPath, dir, configFile string) (*Info, error) {
	prog := filepath.Join(".bin", filepath.Base(toolchainPath))
	if runtime.GOOS == "windows" {
		prog = winExe(dir, prog)
	} else {
		cmdName := filepath.Join(dir, prog)
		if err := checkRegularExecutableFile(cmdName); err != nil {
			return nil, err
		}
	}

	return &Info{
		Path:       toolchainPath,
		Dir:        dir,
		ConfigFile: configFile,
		Program:    prog,
	}, nil
}

// checkRegularExecutableFile verifies that the file pointed to by name is a regular executable file.
func checkRegularExecutableFile(name string) error {
	switch info, err := os.Stat(name); {
	case os.IsNotExist(err):
		return fmt.Errorf("toolchain %s does not exist (should be a regular executable file)", name)
	case err != nil:
		return fmt.Errorf("toolchain %s caused %s (should be a regular executable file)", name, err)
	case info.IsDir():
		return fmt.Errorf("toolchain %s is a directory (should be a regular executable file)", name)
	case !info.Mode().IsRegular():
		return fmt.Errorf("toolchain %s is not a regular file (should be a regular executable file)", name)
	case info.Mode().Perm()&0111 == 0:
		return fmt.Errorf("toolchain %s is not executable (should be a regular executable file)", name)
	default:
		return nil
	}
}

// lookInPaths returns all files in paths (a colon-separated list of
// directories) matching the glob pattern.
func lookInPaths(pattern string, paths string) ([]string, error) {
	var found []string
	seen := map[string]struct{}{}
	for _, dir := range filepath.SplitList(paths) {
		if dir == "" {
			dir = "."
		}
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			return nil, err
		}
		for _, m := range matches {
			if _, seen := seen[m]; seen {
				continue
			}
			seen[m] = struct{}{}
			found = append(found, m)
		}
	}
	sort.Strings(found)
	return found, nil
}

// searches for matching Windows executable (.exe, .bat, .cmd)
func winExe(dir string, program string) string {
	candidate := program + ".exe"
	if _, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
		return candidate
	}
	candidate = program + ".bat"
	if _, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
		return candidate
	}
	candidate = program + ".cmd"
	if _, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
		return candidate
	}
	return ""
}
