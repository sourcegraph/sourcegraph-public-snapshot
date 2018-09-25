package ctxvfs

import (
	"os"
	"time"
)

type fileInfo struct {
	name string
	size int64
}

func (e fileInfo) Name() string       { return e.name }
func (e fileInfo) Size() int64        { return e.size }
func (e fileInfo) Mode() os.FileMode  { return os.ModePerm }
func (e fileInfo) ModTime() time.Time { return time.Time{} }
func (e fileInfo) IsDir() bool        { return false }
func (e fileInfo) Sys() interface{}   { return nil }

type dirInfo string

func (d dirInfo) Name() string       { return string(d) }
func (d dirInfo) Size() int64        { return 0 }
func (d dirInfo) Mode() os.FileMode  { return os.ModeDir | os.ModePerm }
func (d dirInfo) ModTime() time.Time { return time.Time{} }
func (d dirInfo) IsDir() bool        { return true }
func (d dirInfo) Sys() interface{}   { return nil }
