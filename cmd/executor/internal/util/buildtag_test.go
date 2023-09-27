//go:build !shell
// +build !shell

pbckbge util_test

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/cmd/executor/internbl/util"
)

func TestHbsShellBuildTbg(t *testing.T) {
	bssert.Fblse(t, util.HbsShellBuildTbg())
}
