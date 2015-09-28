package ace

import (
	"html/template"
	"sync"
)

var cache = make(map[string]template.Template)
var cacheMutex = new(sync.RWMutex)

// Load loads and returns an HTML template. Each Ace templates are parsed only once
// and cached if the "DynamicReload" option are not set.
func Load(basePath, innerPath string, opts *Options) (*template.Template, error) {
	// Initialize the options.
	opts = InitializeOptions(opts)

	name := basePath + colon + innerPath

	if !opts.DynamicReload {
		if tpl, ok := getCache(name); ok {
			return &tpl, nil
		}
	}

	// Read files.
	src, err := readFiles(basePath, innerPath, opts)
	if err != nil {
		return nil, err
	}

	// Parse the source.
	rslt, err := ParseSource(src, opts)
	if err != nil {
		return nil, err
	}

	// Compile the parsed result.
	tpl, err := CompileResult(name, rslt, opts)
	if err != nil {
		return nil, err
	}

	if !opts.DynamicReload {
		setCache(name, *tpl)
	}

	return tpl, nil
}

// getCache returns the cached template.
func getCache(name string) (template.Template, bool) {
	cacheMutex.RLock()
	tpl, ok := cache[name]
	cacheMutex.RUnlock()
	return tpl, ok
}

// setCache sets the template to the cache.
func setCache(name string, tpl template.Template) {
	cacheMutex.Lock()
	cache[name] = tpl
	cacheMutex.Unlock()
}

// FlushCache clears all cached templates.
func FlushCache() {
       cacheMutex.Lock()
       cache = make(map[string]template.Template)
       cacheMutex.Unlock()
}
