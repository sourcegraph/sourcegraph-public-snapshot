package git

import (
	"errors"
	"io/ioutil"
	"path/filepath"
)

func (repo *Repository) IsTagExist(tagName string) bool {
	tagPath := filepath.Join(repo.Path, "refs/tags", tagName)
	return isFile(tagPath)
}

func (repo *Repository) TagPath(tagName string) string {
	return filepath.Join(repo.Path, "refs/tags", tagName)
}

// GetTags returns all tags of given repository.
func (repo *Repository) GetTags() ([]string, error) {
	return repo.readRefDir("refs/tags", "")
}

func (repo *Repository) CreateTag(tagName, idStr string) error {
	return repo.createRef("tags", tagName, idStr)
}

func CreateTag(repoPath, tagName, id string) error {
	return CreateRef("tags", repoPath, tagName, id)
}

func (repo *Repository) GetTag(tagName string) (*Tag, error) {
	d, err := ioutil.ReadFile(repo.TagPath(tagName))
	if err != nil {
		return nil, err
	}

	id, err := NewIdFromString(string(d))
	if err != nil {
		return nil, err
	}

	tag, err := repo.getTag(id)
	if err != nil {
		return nil, err
	}
	tag.Name = tagName
	return tag, nil
}

func (repo *Repository) getTag(id sha1) (*Tag, error) {
	if repo.tagCache != nil {
		if c, ok := repo.tagCache[id]; ok {
			return c, nil
		}
	} else {
		repo.tagCache = make(map[sha1]*Tag, 10)
	}

	tp, _, dataRc, err := repo.getRawObject(id, false)
	if err != nil {
		return nil, err
	}

	defer func() {
		dataRc.Close()
	}()

	// tag with only reference to commit
	if tp == ObjectCommit {
		tag := new(Tag)
		tag.Id = id
		tag.Object = id
		tag.Type = "commit"
		tag.repo = repo
		repo.tagCache[id] = tag

		return tag, nil
	}

	// tag with message
	if tp != ObjectTag {
		return nil, errors.New("Expected tag type, read error.")
	}

	// TODO reader
	data, err := ioutil.ReadAll(dataRc)
	if err != nil {
		return nil, err
	}

	tag, err := parseTagData(data)
	if err != nil {
		return nil, err
	}

	tag.Id = id
	tag.repo = repo
	repo.tagCache[id] = tag

	return tag, nil
}
