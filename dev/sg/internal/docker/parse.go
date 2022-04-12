package docker

import (
	"bytes"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

func ProcessDockerfile(data []byte, process func(is []instructions.Stage) error) error {
	res, err := parser.Parse(bytes.NewReader(data))
	if err != nil {
		return err
	}
	is, _, err := instructions.Parse(res.AST)
	if err != nil {
		return err
	}
	return process(is)
}
