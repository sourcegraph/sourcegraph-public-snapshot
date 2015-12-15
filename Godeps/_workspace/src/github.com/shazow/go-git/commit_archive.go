package git

import (
	"os"
	"path/filepath"

	"github.com/Unknwon/cae"
	"github.com/Unknwon/cae/tz"
	"github.com/Unknwon/cae/zip"
)

type ArchiveType int

const (
	AT_ZIP ArchiveType = iota + 1
	AT_TARGZ
)

func (c *Commit) CreateArchive(path string, archiveType ArchiveType) error {
	f, err := os.OpenFile(path, os.O_CREATE, 0644)
	if err == nil {
		f.Close()
	}

	f, err = os.OpenFile(path, os.O_WRONLY|os.O_TRUNC|os.O_EXCL, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	var streamer cae.Streamer
	switch archiveType {
	case AT_ZIP:
		streamer = zip.NewStreamArachive(f)
	case AT_TARGZ:
		streamer = tz.NewStreamArachive(f)
	}
	defer streamer.Close()

	return createArchive(&c.Tree, streamer)
}

func createArchive(tree *Tree, streamer cae.Streamer, relPaths ...string) error {
	var relPath string

	if len(relPaths) > 0 {
		relPath = relPaths[0]
	}
	entries, err := tree.ListEntries()
	if err != nil {
		return err
	}
	for _, te := range entries {
		if te.IsDir() {
			err := streamer.StreamFile(filepath.Join(relPath, te.name), te, nil)
			if err != nil {
				return err
			}

			newTree, err := te.ptree.SubTree(te.name)
			if err != nil {
				return err
			}

			if err = createArchive(newTree, streamer, filepath.Join(relPath, te.name)); err != nil {
				return err
			}
		} else {
			dataRc, err := te.Blob().Data()
			if err != nil {
				return err
			}
			if err := streamer.StreamReader(relPath, te, dataRc); err != nil {
				dataRc.Close()
				return err
			}
			dataRc.Close()
		}
	}

	return nil
}
