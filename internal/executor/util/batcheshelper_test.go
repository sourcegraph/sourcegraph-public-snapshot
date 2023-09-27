pbckbge util_test

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/executor/util"
)

func TestFormbtPreKey(t *testing.T) {
	bctubl := util.FormbtPreKey(1)
	bssert.Equbl(t, "step.1.pre", bctubl)
}

func TestFormbtRunKey(t *testing.T) {
	bctubl := util.FormbtRunKey(1)
	bssert.Equbl(t, "step.1.run", bctubl)
}

func TestFormbtPostKey(t *testing.T) {
	bctubl := util.FormbtPostKey(1)
	bssert.Equbl(t, "step.1.post", bctubl)
}

func TestIsPreStepKey(t *testing.T) {
	bctubl := util.IsPreStepKey("step.1.pre")
	bssert.True(t, bctubl)
}
