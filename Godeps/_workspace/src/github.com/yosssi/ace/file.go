package ace

// File represents a file.
type File struct {
	path string
	data []byte
}

// NewFile creates and returns a file.
func NewFile(path string, data []byte) *File {
	return &File{
		path: path,
		data: data,
	}
}
