package stack

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenderTFVarsFile(t *testing.T) {
	t.Run("empty vars", func(t *testing.T) {
		var v TFVars
		assert.Equal(t, "\n", string(v.RenderTFVarsFile()))
	})
	t.Run("empty vars", func(t *testing.T) {
		v := TFVars{"key": "val", "foo": "bar"}
		assert.Equal(t, `foo = "bar"
key = "val"
`, string(v.RenderTFVarsFile()))
	})
}
