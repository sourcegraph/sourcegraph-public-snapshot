package gojq

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ModuleLoader is the interface for loading modules.
//
// Implement following optional methods. Use [NewModuleLoader] to load local modules.
//
//	LoadInitModules() ([]*Query, error)
//	LoadModule(string) (*Query, error)
//	LoadModuleWithMeta(string, map[string]any) (*Query, error)
//	LoadJSON(string) (any, error)
//	LoadJSONWithMeta(string, map[string]any) (any, error)
type ModuleLoader any

// NewModuleLoader creates a new [ModuleLoader] loading local modules in the paths.
// Note that user can load modules outside the paths using "search" path of metadata.
// Empty paths are ignored, so specify "." for the current working directory.
func NewModuleLoader(paths []string) ModuleLoader {
	ps := make([]string, 0, len(paths))
	for _, path := range paths {
		if path = resolvePath(path, ""); path != "" {
			ps = append(ps, path)
		}
	}
	return &moduleLoader{ps}
}

type moduleLoader struct {
	paths []string
}

func (l *moduleLoader) LoadInitModules() ([]*Query, error) {
	var qs []*Query
	for _, path := range l.paths {
		if filepath.Base(path) != ".jq" {
			continue
		}
		fi, err := os.Stat(path)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		if fi.IsDir() {
			continue
		}
		cnt, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		q, err := parseModule(string(cnt), filepath.Dir(path))
		if err != nil {
			return nil, &queryParseError{path, string(cnt), err}
		}
		qs = append(qs, q)
	}
	return qs, nil
}

func (l *moduleLoader) LoadModuleWithMeta(name string, meta map[string]any) (*Query, error) {
	path, err := l.lookupModule(name, ".jq", meta)
	if err != nil {
		return nil, err
	}
	cnt, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	q, err := parseModule(string(cnt), filepath.Dir(path))
	if err != nil {
		return nil, &queryParseError{path, string(cnt), err}
	}
	return q, nil
}

func (l *moduleLoader) LoadJSONWithMeta(name string, meta map[string]any) (any, error) {
	path, err := l.lookupModule(name, ".json", meta)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	vals := []any{}
	dec := json.NewDecoder(f)
	dec.UseNumber()
	for {
		var val any
		if err := dec.Decode(&val); err != nil {
			if err == io.EOF {
				break
			}
			if _, err := f.Seek(0, io.SeekStart); err != nil {
				return nil, err
			}
			cnt, er := io.ReadAll(f)
			if er != nil {
				return nil, er
			}
			return nil, &jsonParseError{path, string(cnt), err}
		}
		vals = append(vals, val)
	}
	return vals, nil
}

func (l *moduleLoader) lookupModule(name, extension string, meta map[string]any) (string, error) {
	paths := l.paths
	if path, ok := meta["search"].(string); ok {
		paths = append([]string{path}, paths...)
	}
	for _, base := range paths {
		path := filepath.Join(base, name+extension)
		if _, err := os.Stat(path); err == nil {
			return path, err
		}
		path = filepath.Join(base, name, filepath.Base(name)+extension)
		if _, err := os.Stat(path); err == nil {
			return path, err
		}
	}
	return "", fmt.Errorf("module not found: %q", name)
}

func parseModule(cnt, dir string) (*Query, error) {
	q, err := Parse(cnt)
	if err != nil {
		return nil, err
	}
	for _, i := range q.Imports {
		if i.Meta != nil {
			for _, e := range i.Meta.KeyVals {
				if e.Key == "search" || e.KeyString == "search" {
					if path, ok := e.Val.toString(); ok {
						if path = resolvePath(path, dir); path != "" {
							e.Val.Str = path
						} else {
							e.Val.Null = true
						}
					}
				}
			}
		}
	}
	return q, nil
}

func resolvePath(path, dir string) string {
	switch {
	case filepath.IsAbs(path):
		return path
	case strings.HasPrefix(path, "~/"):
		dir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		return filepath.Join(dir, path[2:])
	case strings.HasPrefix(path, "$ORIGIN/"):
		exe, err := os.Executable()
		if err != nil {
			return ""
		}
		exe, err = filepath.EvalSymlinks(exe)
		if err != nil {
			return ""
		}
		return filepath.Join(filepath.Dir(exe), path[8:])
	default:
		return filepath.Join(dir, path)
	}
}
