package luasandbox

import (
	"embed"
	"fmt"
	"io"
	"path"
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
	files, err := getAllFilepaths(fs, dir)
	if err != nil {
		return nil, err
	}

	modules := make(map[string]string, len(files))
	for _, file := range files {
		contents, err := readFile(fs, file)
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

func getAllFilepaths(fs embed.FS, dir string) (out []string, err error) {
	if dir == "" {
		dir = "."
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		// All paths in embed FS are unix paths, so we need to use Unix, even on windows.
		// Thus, we don't use filepath here.
		f := path.Join(dir, entry.Name())

		if entry.IsDir() {
			descendents, err := getAllFilepaths(fs, f)
			if err != nil {
				return nil, err
			}

			out = append(out, descendents...)
		} else {
			out = append(out, f)
		}
	}
	return
}

func splitPathList(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ":")
}

func readFile(fs embed.FS, filepath string) (string, error) {
	f, err := fs.Open(filepath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contents, err := io.ReadAll(f)
	return string(contents), err
}
