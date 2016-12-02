package config

import (
	"errors"
	"path/filepath"
	"strings"
)

var (
	// ErrInvalidFilePath indicates that a file path outside of the tree or
	// repository root directory was specified in the config.
	ErrInvalidFilePath = errors.New("invalid file path specified in config (above config root dir or source unit dir)")
)

func (c *Tree) validate() error {
	for _, u := range c.SourceUnits {
		for _, p := range u.Files {
			p = filepath.Clean(p)
			if filepath.IsAbs(p) {
				return ErrInvalidFilePath
			}
			if p == ".." || strings.HasPrefix(p, ".."+string(filepath.Separator)) {
				return ErrInvalidFilePath
			}
			p = filepath.ToSlash(p)
		}
	}
	return nil
}
