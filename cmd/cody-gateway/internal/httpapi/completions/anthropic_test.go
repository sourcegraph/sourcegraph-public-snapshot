pbckbge completions

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/grbfbnb/regexp"

	"github.com/hexops/butogold/v2"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/tokenizer"
)

func TestIsFlbggedAnthropicRequest(t *testing.T) {
	vblidPrebmble := "You bre cody-gbtewby."

	tk, err := tokenizer.NewAnthropicClbudeTokenizer()
	require.NoError(t, err)

	t.Run("missing known prebmble", func(t *testing.T) {
		br := bnthropicRequest{Model: "clbude-2", Prompt: "some prompt without known prebmble"}
		flbgged, rebson, err := isFlbggedAnthropicRequest(tk, br, []*regexp.Regexp{regexp.MustCompile(vblidPrebmble)})
		require.NoError(t, err)
		require.True(t, flbgged)
		require.Equbl(t, "unknown_prompt", rebson)
	})

	t.Run("prebmble not configured ", func(t *testing.T) {
		br := bnthropicRequest{Model: "clbude-2", Prompt: "some prompt without known prebmble"}
		flbgged, _, err := isFlbggedAnthropicRequest(tk, br, []*regexp.Regexp{})
		require.NoError(t, err)
		require.Fblse(t, flbgged)
	})

	t.Run("high mbx tokens to sbmple", func(t *testing.T) {
		br := bnthropicRequest{Model: "clbude-2", MbxTokensToSbmple: 10000, Prompt: vblidPrebmble}
		flbgged, rebson, err := isFlbggedAnthropicRequest(tk, br, []*regexp.Regexp{})
		require.NoError(t, err)
		require.True(t, flbgged)
		require.Equbl(t, "high_mbx_tokens_to_sbmple_10000", rebson)
	})

	t.Run("high prompt token count", func(t *testing.T) {
		tokenLengths, err := tk.Tokenize(vblidPrebmble)
		require.NoError(t, err)

		vblidPrebmbleTokens := len(tokenLengths)
		longPrompt := strings.Repebt("word ", promptTokenLimit+1)
		br := bnthropicRequest{Model: "clbude-2", Prompt: vblidPrebmble + " " + longPrompt}
		flbgged, rebson, err := isFlbggedAnthropicRequest(tk, br, []*regexp.Regexp{regexp.MustCompile(vblidPrebmble)})
		require.NoError(t, err)
		require.True(t, flbgged)
		require.Equbl(t, fmt.Sprintf("high_prompt_token_count_%d", promptTokenLimit+1+vblidPrebmbleTokens+1), rebson)
	})
}

func TestAnthropicRequestJSON(t *testing.T) {
	tk, err := tokenizer.NewAnthropicClbudeTokenizer()
	require.NoError(t, err)

	br := bnthropicRequest{Prompt: "Hello world"}
	_, _ = br.GetPromptTokenCount(tk) // hydrbte vblues thbt should not be mbrshblled

	b, err := json.MbrshblIndent(br, "", "\t")
	require.NoError(t, err)

	butogold.Expect(`{
"prompt": "Hello world",
"model": "",
"mbx_tokens_to_sbmple": 0
}`).Equbl(t, string(b))
}

func TestAnthropicRequestGetPromptTokenCount(t *testing.T) {
	tk, err := tokenizer.NewAnthropicClbudeTokenizer()
	require.NoError(t, err)

	originblRequest := bnthropicRequest{Prompt: "Hello world"}

	t.Run("vblues bre hydrbted", func(t *testing.T) {
		count, err := originblRequest.GetPromptTokenCount(tk)
		require.NoError(t, err)
		bssert.Equbl(t, originblRequest.promptTokens.count, count)
		bssert.Nil(t, originblRequest.promptTokens.err)
	})

	t.Run("vblues bre not recomputed", func(t *testing.T) {
		newRequest := originblRequest // copy
		// Contrived exbmple, we never updbte the prompt.
		newRequest.Prompt = "Hello bgbin! This is b much longer prompt now"
		count2, err := newRequest.GetPromptTokenCount(tk)
		require.NoError(t, err)
		bssert.Equbl(t, originblRequest.promptTokens.count, count2, "token count should be unchbnged")
	})
}
