package graph

import "errors"

var ErrDefNotExist = errors.New("def does not exist")

func IsNotExist(err error) bool {
	return err == ErrDefNotExist
}
