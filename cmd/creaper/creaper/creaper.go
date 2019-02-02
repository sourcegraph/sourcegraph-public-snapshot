package creaper

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"time"

	logger "github.com/sourcegraph/sourcegraph/cmd/creaper/logger"
	times "gopkg.in/djherbis/times.v1"
)

func Reap(
	ctx context.Context,
	cacheDir string,
	frequency time.Duration,
	maxCacheSize uint64,
) {
	ctx = logger.WithLogger(ctx, "cacheDir", cacheDir, "maxCacheSize", maxCacheSize)
	timer := time.NewTimer(frequency)

	for {
		select {
		case <-timer.C:
			reap(ctx, cacheDir, maxCacheSize)
			timer.Reset(frequency)
		}
	}

}

func reap(ctx context.Context, directory string, maxSize uint64) {
	logger.Info(ctx, "Fetching directory info")

	files, totalSize := getDirectoryInfo(ctx, directory)

	ctx = logger.WithLogger(ctx, "fileCount", len(files), "size", totalSize)

	if totalSize < maxSize {
		logger.Info(ctx, "Cache size within limit")
		return
	}

	logger.Info(ctx, "Cache is overgrown. Reaping.")

	clearSpace(ctx, files, totalSize, 4*maxSize/5)

	logger.Info(ctx, "Reap complete.")
}

func getDirectoryInfo(ctx context.Context, directory string) ([]cachedFile, uint64) {
	files := make([]cachedFile, 0)
	totalSize := uint64(0)

	if err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Warn(
				ctx,
				"Error walking cache directory",
				"path", path,
				"err", err,
			)

			// Try to proceed.
			return filepath.SkipDir
		}

		if info.Mode().IsRegular() {
			files = append(
				files,
				cachedFile{
					Path:     path,
					Size:     uint64(info.Size()),
					Timespec: times.Get(info),
				},
			)
			totalSize += uint64(info.Size())
		}
		return nil
	}); err != nil {
		logger.Warn(
			ctx,
			"Error when walking cache directory",
			"err", err,
		)
	}

	return files, totalSize
}

func clearSpace(ctx context.Context, cachedFiles []cachedFile, totalSize uint64, desiredSize uint64) {
	ctx = logger.WithLogger(ctx, "totalSize", totalSize, "desiredSize", desiredSize)

	logger.Info(ctx, "Sorting cached files")

	// Sort the cachedFiles by last access time ascending (least-recently-accessed at the top).
	sort.Slice(
		cachedFiles,
		func(i int, j int) bool {
			return cachedFiles[i].Timespec.AccessTime().Before(cachedFiles[j].Timespec.AccessTime())
		},
	)

	logger.Info(ctx, "Done sorting cached files")

	// Iterate over the files, removing them until the size is at or below desiredSize.
	currentSize := totalSize
	for _, file := range cachedFiles {
		if err := os.Remove(file.Path); err != nil {
			// Log the error but keep on trying.
			logger.Warn(ctx, "Error removing file.", "file", file.Path, "err", err)
			continue
		}

		currentSize -= file.Size

		logger.Debug(ctx, "Deleted file", "file", file.Path, "currentSize", currentSize)

		if currentSize <= desiredSize {
			logger.Info(
				ctx,
				"Enough files have been removed. Finishing.",
				"currentSize", currentSize,
			)
			return
		}
	}

	logger.Warn(ctx, "Unable to clear enough space", "currentSize", currentSize)
}

type cachedFile struct {
	Path     string
	Size     uint64
	Timespec times.Timespec
}
