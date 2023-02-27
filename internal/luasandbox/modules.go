package luasandbox

import (
	"embed"
	"fmt"
	"io"
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
		f, err := fs.Open(file)
		if err != nil {
			return nil, err
		}

		contents, err := io.ReadAll(f)
		if err != nil {
			return nil, err
		}

		name := strings.Join(filepath.SplitList(strings.TrimSuffix(filepath.Base(file), filepath.Ext(file))), ".")
		if prefix != "" {
			name = prefix + "." + name
		}

		modules[name] = string(contents)
	}

	return modules, nil
}

func getAllFilepaths(fs embed.FS, path string) (out []string, err error) {
	if path == "" {
		path = "."
	}

	entries, err := fs.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		path := filepath.Join(path, entry.Name())

		if entry.IsDir() {
			descendents, err := getAllFilepaths(fs, path)
			if err != nil {
				return nil, err
			}

			out = append(out, descendents...)
		} else {
			out = append(out, path)
		}
	}
	return
}
