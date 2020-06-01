package paths

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// Migrate ensure that paths are named correctly after changes to the filesystem layout between
// releases. This method blocks the startup of the bundle manager.
//
// 3.16 -> 3.17:
//   Rename phase
//     - dbs/{id}.lsif.db           -> dbs/{id}/sqlite.db
//     - dbs/{id}.{idx}.lsif.db     -> db-parts/{id}.{idx}.gz
//     - uploads/{id}.lsif.gz       -> uploads/{id}.gz
//     - uploads/{id}.{idx}.lsif.gz -> upload-parts/{id}.{idx}.gz
func Migrate(bundleDir string) error {
	if err := migratePaths(bundleDir, UploadsDir, UploadFilename, UploadPartFilename); err != nil {
		return err
	}

	if err := migratePaths(bundleDir, DBsDir, SQLiteDBFilename, DBPartFilename); err != nil {
		return err
	}

	return nil
}

// migratePaths renames all (non-directory) files in the root dir constructed by makeRootDir to
// the filename constructed by either makeFilename or makePartFilename, depending on if the file
// specifies a part index.
func migratePaths(
	bundleDir string,
	makeRootDir func(bundleDir string) string,
	makeFilename func(bundleDir string, id int64) string,
	makePartFilename func(bundleDir string, id, index int64) string,
) error {
	root := makeRootDir(bundleDir)

	infos, err := ioutil.ReadDir(root)
	if err != nil {
		return err
	}

	for _, info := range infos {
		if info.IsDir() {
			continue
		}

		path := filepath.Join(root, info.Name())

		id, partIndex := getIDAndPartIndex(info.Name())
		if id == -1 {
			continue
		}

		var newPath string
		if partIndex == -1 {
			newPath = makeFilename(bundleDir, int64(id))
		} else {
			newPath = makePartFilename(bundleDir, int64(id), int64(partIndex))
		}

		if err := os.MkdirAll(filepath.Dir(newPath), os.ModePerm); err != nil {
			return err
		}

		if err := os.Rename(path, newPath); err != nil {
			return err
		}
	}

	return nil
}

var filenamePattern = regexp.MustCompile(`([0-9]+)(?:\.([0-9]+))?\.lsif\.(db|gz)`)

// getIDAndPartIndex extracts the bundle id and part index from the given filename.
// If the filename does not describe a part file, -1 is returned for the part index.
func getIDAndPartIndex(filename string) (id, index int) {
	matches := filenamePattern.FindAllStringSubmatch(filename, -1)
	if len(matches) == 0 {
		return -1, -1
	}

	return toInt(matches[0][1]), toInt(matches[0][2])
}

func toInt(raw string) int {
	if val, err := strconv.Atoi(raw); err == nil {
		return val
	}
	return -1
}
