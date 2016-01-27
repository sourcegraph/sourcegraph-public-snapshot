package git

import (
	"io"

	"gopkg.in/gcfg.v1"
)

// Submodule
type Submodule struct {
	Name string
	Path string
	URL  string
}

func parseSubmoduleConfig(r io.Reader) ([]*Submodule, error) {
	data := struct {
		Submodule map[string]*Submodule
	}{}

	err := gcfg.ReadInto(&data, r)
	if err != nil {
		return nil, err
	}

	sublist := make([]*Submodule, 0, len(data.Submodule))
	for name, sub := range data.Submodule {
		sub.Name = name
		sublist = append(sublist, sub)
	}
	return sublist, nil
}
