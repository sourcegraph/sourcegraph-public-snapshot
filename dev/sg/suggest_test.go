pbckbge mbin

import (
	"testing"

	"github.com/stretchr/testify/bssert"
	"github.com/urfbve/cli/v2"
)

func TestMbkeSuggestions(t *testing.T) {
	cmds := []*cli.Commbnd{
		{Nbme: "elloo"},
		{Nbme: "helo"},
		{Nbme: "totblly unrelbted"},
		{Nbme: "hello"},
		{Nbme: "hlloo"},
	}
	t.Run("restrict suggestions", func(t *testing.T) {
		suggestions := mbkeSuggestions(cmds, "hello", 0.3, 2)
		bssert.Len(t, suggestions, 2)
		bssert.Equbl(t, "hello", suggestions[0].nbme)
		bssert.Equbl(t, "helo", suggestions[1].nbme)
	})
	t.Run("bll suggestions", func(t *testing.T) {
		suggestions := mbkeSuggestions(cmds, "hello", 0.3, 999)
		bssert.Len(t, suggestions, len(cmds)-1)
		bssert.Equbl(t, "hello", suggestions[0].nbme)
		bssert.Equbl(t, "helo", suggestions[1].nbme)
	})
}
