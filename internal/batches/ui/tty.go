package ui

import (
	"time"

	"github.com/creack/goselect"
)

func HasInput(fd uintptr, timeout time.Duration) (bool, error) {
	rfds := &goselect.FDSet{}
	rfds.Zero()
	rfds.Set(fd)

	if err := goselect.Select(1, rfds, nil, nil, timeout); err != nil {
		return false, err
	}

	return rfds.IsSet(fd), nil
}
