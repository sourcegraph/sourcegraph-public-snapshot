package tmpl

import "path"

// FileMapDataItem represents a single pair in a map.
type FileMapDataItem struct {
	Key   string
	Value string
}

// FileMapGenerate represents a generate tag.
type FileMapGenerate struct {
	// Template is the path of the template file to use for generating the
	// target.
	Template string

	// Target is the target proto file for generation. It must match one of the
	// input proto files, or else the template will not be executed.
	Target string `xml:",omitempty"`

	// Output is the output file to write the executed template contents to.
	Output string

	// Include is a list of template files to include for execution of the
	// template.
	Include []string `xml:"Includes>Include,omitempty"`

	// Data is effectively an map of items to pass onto the template during
	// execution.
	Data []*FileMapDataItem `xml:"Data>Item,omitempty"`
}

// DataMap returns f.Data but as a Go map. It panics if there are any duplicate
// keys.
func (f *FileMapGenerate) DataMap() map[string]string {
	m := make(map[string]string, len(f.Data))
	for _, d := range f.Data {
		if _, ok := m[d.Key]; ok {
			panic("duplicate data key")
		}
		m[d.Key] = d.Value
	}
	return m
}

// FileMap represents a file mapping.
type FileMap struct {
	// Dir is the directory to resolve template paths mentioned in the filemap
	// relative to. It should be the directory that this file map resides inside
	// of. The path will be converted to a unix-style one for protobuf, which
	// only deals with unix-style paths.
	Dir string `xml:",omitempty"`

	Generate []*FileMapGenerate `xml:"Generate"`
}

// relative returns a list of relative paths prefixed with the f.Dir path (also
// converted to a unix-style path).
func (f *FileMap) relative(rel ...string) []string {
	dir := unixPath(f.Dir)
	var abs []string
	for _, p := range rel {
		abs = append(abs, path.Join(dir, p))
	}
	return abs
}
