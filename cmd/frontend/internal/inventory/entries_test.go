package inventory

import (
	"archive/tar"
	"context"
	"io"
	"reflect"
	"strings"
	"testing"
)

type FileData struct {
	Name        string
	Content     string
	IsDirectory bool
}

func TestFileMode(t *testing.T) {
	testCases := []struct {
		name        string
		isDirectory bool
	}{
		{
			name:        "regular file",
			isDirectory: false,
		},
		{
			name:        "directory",
			isDirectory: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			h := &tar.Header{
				Name: tc.name,
				Mode: func() int64 {
					if tc.isDirectory {
						// 04 (in octal) sets the directory bit
						return 040000
					}
					return 0
				}(),
			}

			got := h.FileInfo().Mode().IsDir()
			if got != tc.isDirectory {
				t.Errorf("got %v, want %v", got, tc.isDirectory)
			}
		})
	}

}

func TestEntriesNextProcessor(t *testing.T) {
	testCases := []struct {
		name     string
		files    []FileData
		expected Inventory
	}{
		{
			name: "Go and Objective-C (in directory) files",
			files: []FileData{
				{Name: "b.go", Content: "package main"},
				{Name: "a", Content: "", IsDirectory: true},
				{Name: "a/c.m", Content: "@interface X:NSObject {}"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
		},
		{
			name: "Single file",
			files: []FileData{
				{Name: "b.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
		},
		{
			name: "Two root file",
			files: []FileData{
				{Name: "a.go", Content: "package main"},
				{Name: "b.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Go", TotalBytes: 24, TotalLines: 2},
			}},
		},
		{
			name: "File and empty directory",
			files: []FileData{
				{Name: "a.go", Content: "package main"},
				{Name: "b", Content: "", IsDirectory: true},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
		},
		{
			name: "Two directories with files",
			files: []FileData{
				{Name: "a", Content: "", IsDirectory: true},
				{Name: "a/a.m", Content: "@interface X:NSObject {}"},
				{Name: "b", Content: "", IsDirectory: true},
				{Name: "b/b.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
		},
		{
			name: "Two directories with files, and a root file",
			files: []FileData{
				{Name: "a", Content: "", IsDirectory: true},
				{Name: "a/a.m", Content: "@interface X:NSObject {}"},
				{Name: "b", Content: "", IsDirectory: true},
				{Name: "b/b.go", Content: "package main"},
				{Name: "c.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Go", TotalBytes: 24, TotalLines: 2},
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
			}},
		},
		{
			name: "Directory with sub-directory and surrounding files",
			files: []FileData{
				{Name: "a", Content: "", IsDirectory: true},
				{Name: "a/A.m", Content: "@interface X:NSObject {}"},
				{Name: "a/b", Content: "", IsDirectory: true},
				{Name: "a/b/B.m", Content: "@interface X:NSObject {}"},
				{Name: "a/c", Content: "", IsDirectory: true},
				{Name: "a/c/D.m", Content: "@interface X:NSObject {}"},
				{Name: "E.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 72, TotalLines: 3},
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			index := 0

			c := Context{
				Repo:                                "testRepo",
				CommitID:                            "testCommit",
				ReadTree:                            nil,
				NewFileReader:                       nil,
				CacheKey:                            nil,
				CacheGet:                            nil,
				CacheSet:                            nil,
				ShouldSkipEnhancedLanguageDetection: false,
				GitServerClient:                     nil,
			}

			next := func() (*NextRecord, error) {
				defer func() {
					index += 1
				}()
				if index >= len(tc.files) {
					return nil, io.EOF
				}
				content := tc.files[index].Content
				return &NextRecord{
					Header: &tar.Header{
						Name: tc.files[index].Name,
						Size: int64(len(content)),
						Mode: func() int64 {
							if tc.files[index].IsDirectory {
								// 04 (in octal) sets the directory bit
								return 040000
							}
							return 0
						}(),
					},
					GetFileReader: func(ctx context.Context, path string) (io.ReadCloser, error) {
						return io.NopCloser(strings.NewReader(content)), nil
					},
				}, nil
			}

			got, err := c.ArchiveProcessor(context.Background(), next)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("got %+v, want %+v", got, tc.expected)
			}
		})
	}

}

func TestSum(t *testing.T) {
	testCases := []struct {
		name     string
		invs     []Inventory
		expected Inventory
	}{
		{
			name: "empty input",
			invs: []Inventory{},
			expected: Inventory{
				Languages: []Lang{},
			},
		},
		{
			name: "single inventory",
			invs: []Inventory{
				{
					Languages: []Lang{
						{Name: "Go", TotalBytes: 100, TotalLines: 10},
						{Name: "Python", TotalBytes: 200, TotalLines: 20},
					},
				},
			},
			expected: Inventory{
				Languages: []Lang{
					{Name: "Python", TotalBytes: 200, TotalLines: 20},
					{Name: "Go", TotalBytes: 100, TotalLines: 10},
				},
			},
		},
		{
			name: "multiple inventories",
			invs: []Inventory{
				{
					Languages: []Lang{
						{Name: "Go", TotalBytes: 100, TotalLines: 10},
						{Name: "Python", TotalBytes: 200, TotalLines: 20},
					},
				},
				{
					Languages: []Lang{
						{Name: "Go", TotalBytes: 50, TotalLines: 5},
						{Name: "Ruby", TotalBytes: 300, TotalLines: 30},
					},
				},
			},
			expected: Inventory{
				Languages: []Lang{
					{Name: "Ruby", TotalBytes: 300, TotalLines: 30},
					{Name: "Python", TotalBytes: 200, TotalLines: 20},
					{Name: "Go", TotalBytes: 150, TotalLines: 15},
				},
			},
		},
		{
			name: "empty language name",
			invs: []Inventory{
				{
					Languages: []Lang{
						{Name: "", TotalBytes: 100, TotalLines: 10},
						{Name: "Python", TotalBytes: 200, TotalLines: 20},
					},
				},
			},
			expected: Inventory{
				Languages: []Lang{
					{Name: "Python", TotalBytes: 200, TotalLines: 20},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Sum(tc.invs)
			if !reflect.DeepEqual(got, tc.expected) {
				t.Errorf("got %+v, want %+v", got, tc.expected)
			}
		})
	}
}

func BenchmarkSum(b *testing.B) {
	// Benchmark results
	// n=10000: ~8400 ns/op => 8.4 ns/record
	// n=1000: ~8500 ns/op => 8.5 ns/record
	// n=100: ~1100 ns/op => 11 ns/record
	// n=10: ~360 ns/op => 36 ns/record
	// n=5: ~300 ns/op => 60 ns/record
	// n=2: ~280 ns/op => 140 ns/record
	n := 10_000
	invs := make([]Inventory, n)
	for i := 0; i < n; i++ {
		invs[i] = Inventory{
			Languages: []Lang{
				{Name: "Go", TotalBytes: uint64(100 + int64(i)), TotalLines: uint64(10 + int64(i))},
				{Name: "Python", TotalBytes: uint64(200 + int64(i)), TotalLines: uint64(20 + int64(i))},
			},
		}
		if i%2 == 1 {
			invs[i].Languages[1] = Lang{Name: "Ruby", TotalBytes: uint64(300 + int64(i)), TotalLines: uint64(30 + int64(i))}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Sum(invs)
	}
}
