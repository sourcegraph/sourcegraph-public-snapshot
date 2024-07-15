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

func TestEntriesNextProcessorWithoutCaching(t *testing.T) {
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
				NewTarReader:                        nil,
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
				return &NextRecord{
					Header: &tar.Header{
						Name: tc.files[index].Name,
						Size: int64(len(tc.files[index].Content)),
						Mode: func() int64 {
							if tc.files[index].IsDirectory {
								// 04 (in octal) sets the directory bit
								return 040000
							}
							return 0
						}(),
					},
					FileReader: io.NopCloser(strings.NewReader(tc.files[index].Content)),
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

func TestEntriesNextProcessorWithCaching(t *testing.T) {
	testCases := []struct {
		name          string
		files         []FileData
		expected      Inventory
		expectedCache map[string]Inventory
	}{
		{
			name: "Go and Objective-C files",
			files: []FileData{
				{Name: "a", Content: "", IsDirectory: true},
				{Name: "a/c.m", Content: "@interface X:NSObject {}"},
				{Name: "b.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
			expectedCache: map[string]Inventory{
				"testRepo/.:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
					{Name: "Go", TotalBytes: 12, TotalLines: 1},
				}},
				"testRepo/a:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				}},
			},
		},
		{
			name: "Single file",
			files: []FileData{
				{Name: "b.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
			expectedCache: map[string]Inventory{
				"testRepo/.:testCommit": {Languages: []Lang{
					{Name: "Go", TotalBytes: 12, TotalLines: 1},
				}},
			},
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
			expectedCache: map[string]Inventory{
				"testRepo/.:testCommit": {Languages: []Lang{
					{Name: "Go", TotalBytes: 24, TotalLines: 2},
				}},
			},
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
			expectedCache: map[string]Inventory{
				"testRepo/.:testCommit": {Languages: []Lang{
					{Name: "Go", TotalBytes: 12, TotalLines: 1},
				}},
			},
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
			expectedCache: map[string]Inventory{
				"testRepo/.:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
					{Name: "Go", TotalBytes: 12, TotalLines: 1},
				}},
				"testRepo/a:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				}},
				"testRepo/b:testCommit": {Languages: []Lang{
					{Name: "Go", TotalBytes: 12, TotalLines: 1},
				}},
			},
		},
		{
			name: "Two directories with files and a root file",
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
			expectedCache: map[string]Inventory{
				"testRepo/.:testCommit": {Languages: []Lang{
					{Name: "Go", TotalBytes: 24, TotalLines: 2},
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				}},
				"testRepo/a:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				}},
				"testRepo/b:testCommit": {Languages: []Lang{
					{Name: "Go", TotalBytes: 12, TotalLines: 1},
				}},
			},
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
				{Name: "a/E.m", Content: "@interface X:NSObject {}"},
				{Name: "F.go", Content: "package main"},
			},
			expected: Inventory{Languages: []Lang{
				{Name: "Objective-C", TotalBytes: 96, TotalLines: 4},
				{Name: "Go", TotalBytes: 12, TotalLines: 1},
			}},
			expectedCache: map[string]Inventory{
				"testRepo/.:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 96, TotalLines: 4},
					{Name: "Go", TotalBytes: 12, TotalLines: 1},
				}},
				"testRepo/a/b:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				}},
				"testRepo/a/c:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 24, TotalLines: 1},
				}},
				"testRepo/a:testCommit": {Languages: []Lang{
					{Name: "Objective-C", TotalBytes: 96, TotalLines: 4},
				}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			index := 0

			cache := map[string]Inventory{}

			c := Context{
				Repo:          "testRepo",
				CommitID:      "testCommit",
				ReadTree:      nil,
				NewFileReader: nil,
				CacheKey:      nil,
				CacheGet:      nil,
				CacheSet: func(_ context.Context, key string, inv Inventory) {
					cache[key] = inv
				},
				NewTarReader:                        nil,
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
				return &NextRecord{
					Header: &tar.Header{
						Name: tc.files[index].Name,
						Size: int64(len(tc.files[index].Content)),
						Mode: func() int64 {
							if tc.files[index].IsDirectory {
								// 04 (in octal) sets the directory bit
								return 040000
							}
							return 0
						}(),
					},
					FileReader: io.NopCloser(strings.NewReader(tc.files[index].Content)),
				}, nil
			}

			got, err := c.ArchiveProcessor(context.Background(), next)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.expected, got) {
				t.Errorf("got %+v, want %+v", got, tc.expected)
			}

			if !reflect.DeepEqual(tc.expectedCache, cache) {
				t.Errorf("got %+v, want %+v", cache, tc.expectedCache)
			}
		})
	}
}
