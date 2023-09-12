package luasandbox

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed lua/*
var luaRuntime embed.FS

var DefaultLuaModules = map[string]string{}

func init() {
	modules, err := LuaModulesFromFS(luaRuntime, "lua", "")
	if err != nil {
		panic(fmt.Sprintf("error loading Lua runtime files: %s", err))
	}

	DefaultLuaModules = modules
}

func LuaModulesFromFS(fs embed.FS, dir, prefix string) (map[string]string, error) {
	files, err := listFilesRecursive(fs, dir)
	if err != nil {
		return nil, err
	}

	modules := make(map[string]string, len(files))
	for _, file := range files {
		contents, err := readAll(fs, file)
		if err != nil {
			return nil, err
		}

		// All paths in embed FS are unix paths, so we need to use Unix, even on windows.
		// Thus, we don't use filepath here.
		name := strings.Join(splitPathList(strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))), ".")

		if prefix != "" {
			name = prefix + "." + name
		}

		modules[name] = contents
	}

	return modules, nil
}

// listFilesRecursive performs a recursive walk of filesystem rooted at relative
// path 'root' and returns the list of all files.
//
// 'root' should be a subdirectory, or '.'
func listFilesRecursive(filesystem fs.FS, root string) (out []string, err error) {
	err = fs.WalkDir(filesystem, root, func(path string, dirEntry fs.DirEntry, err error) error {
		if !dirEntry.IsDir() {
			out = append(out, path)
		}
		return nil
	})
	return out, err
}

func splitPathList(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ":")
}

func readAll(fs embed.FS, filepath string) (string, error) {
	f, err := fs.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents, err := io.ReadAll(f)
	return string(contents), err
}
