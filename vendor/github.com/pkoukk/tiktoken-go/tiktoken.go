package tiktoken

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dlclark/regexp2"
)

var bpeLoader BpeLoader = NewDefaultBpeLoader()

func SetBpeLoader(loader BpeLoader) {
	bpeLoader = loader
}

func GetEncoding(encodingName string) (*Tiktoken, error) {
	enc, err := getEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	pbe, err := NewCoreBPE(enc.MergeableRanks, enc.SpecialTokens, enc.PatStr)
	if err != nil {
		return nil, err
	}
	specialTokensSet := map[string]any{}
	for k := range enc.SpecialTokens {
		specialTokensSet[k] = true
	}
	return NewTiktoken(pbe, enc, specialTokensSet), nil
}

func EncodingForModel(modelName string) (*Tiktoken, error) {
	if encodingName, ok := MODEL_TO_ENCODING[modelName]; ok {
		return GetEncoding(encodingName)
	} else {
		for prefix, encodingName := range MODEL_PREFIX_TO_ENCODING {
			if strings.HasPrefix(modelName, prefix) {
				return GetEncoding(encodingName)
			}
		}
	}
	return nil, fmt.Errorf("no encoding for model %s", modelName)
}

type Tiktoken struct {
	bpe              *CoreBPE
	pbeEncoding      *Encoding
	specialTokensSet map[string]any
}

func (t *Tiktoken) Encode(text string, allowedSpecial []string, disallowedSpecial []string) []int {
	var allowedSpecialSet map[string]any
	if len(allowedSpecial) == 0 {
		allowedSpecialSet = map[string]any{}
	} else if len(allowedSpecial) == 1 && allowedSpecial[0] == "all" {
		allowedSpecialSet = t.specialTokensSet
	} else {
		allowedSpecialSet = map[string]any{}
		for _, v := range allowedSpecial {
			allowedSpecialSet[v] = nil
		}
	}

	disallowedSpecialSet := map[string]any{}
	for _, v := range disallowedSpecial {
		disallowedSpecialSet[v] = nil
	}
	if len(disallowedSpecial) == 1 && disallowedSpecial[0] == "all" {
		disallowedSpecialSet = difference(t.specialTokensSet, allowedSpecialSet)
	}

	if len(disallowedSpecialSet) > 0 {
		specialRegex := t.SpecialTokenRegex(disallowedSpecialSet)
		m := findRegex2StringMatch(text, specialRegex)
		if m != "" {
			panic(fmt.Sprintf("text contains disallowed special token %s", m))
		}
	}

	tokens, _ := t.bpe.encodeNative(text, allowedSpecialSet)
	return tokens
}

func (t *Tiktoken) EncodeOrdinary(text string) []int {
	return (t.bpe.encodeOrdinaryNative(text))
}

func (t *Tiktoken) Decode(tokens []int) string {
	return string(t.bpe.decodeNative(tokens))
}

func (t *Tiktoken) SpecialTokenRegex(disallowedSpecialSet map[string]any) *regexp2.Regexp {
	specialRegexStrs := make([]string, 0, len(disallowedSpecialSet))
	for k := range disallowedSpecialSet {
		specialRegexStrs = append(specialRegexStrs, regexp.QuoteMeta(k))
	}
	specialRegex := regexp2.MustCompile(strings.Join(specialRegexStrs, "|"), regexp2.None)
	return specialRegex
}

func findRegex2StringMatch(text string, reg *regexp2.Regexp) string {
	m, _ := reg.FindStringMatch(text)
	if m == nil {
		return ""
	}

	return m.String()
}

func difference(setA, setB map[string]any) map[string]any {
	result := make(map[string]any)
	for k := range setA {
		if _, ok := setB[k]; !ok {
			result[k] = true
		}
	}
	return result
}

// NewTiktoken can be used to create a *Tiktoken with custom parameters.
func NewTiktoken(bpe *CoreBPE, encoding *Encoding, specialTokensSet map[string]any) *Tiktoken {
	return &Tiktoken{
		bpe:              bpe,
		pbeEncoding:      encoding,
		specialTokensSet: specialTokensSet,
	}
}
