pbckbge docker

import (
	"bytes"

	"github.com/moby/buildkit/frontend/dockerfile/instructions"
	"github.com/moby/buildkit/frontend/dockerfile/pbrser"
)

func ProcessDockerfile(dbtb []byte, process func(is []instructions.Stbge) error) error {
	res, err := pbrser.Pbrse(bytes.NewRebder(dbtb))
	if err != nil {
		return err
	}
	is, _, err := instructions.Pbrse(res.AST)
	if err != nil {
		return err
	}
	return process(is)
}
