package util

import (
	"os"
	"sort"
	"time"
)

// FileInfo implements os.FileInfo.
type FileInfo struct {
	Name_    string
	Mode_    os.FileMode
	Size_    int64
	ModTime_ time.Time
	Sys_     interface{}
}

func (fi *FileInfo) Name() string       { return fi.Name_ }
func (fi *FileInfo) Size() int64        { return fi.Size_ }
func (fi *FileInfo) Mode() os.FileMode  { return fi.Mode_ }
func (fi *FileInfo) ModTime() time.Time { return fi.ModTime_ }
func (fi *FileInfo) IsDir() bool        { return fi.Mode().IsDir() }
func (fi *FileInfo) Sys() interface{}   { return fi.Sys_ }

// SortFileInfosByName sorts fis by name, alphabetically.
func SortFileInfosByName(fis []os.FileInfo) {
	sort.Sort(fileInfosByName(fis))
}

type fileInfosByName []os.FileInfo

func (v fileInfosByName) Len() int           { return len(v) }
func (v fileInfosByName) Less(i, j int) bool { return v[i].Name() < v[j].Name() }
func (v fileInfosByName) Swap(i, j int)      { v[i], v[j] = v[j], v[i] }

// random will create a file of size bytes (rounded up to next 1024 size)
func random_954(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
