package tiktoken

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dlclark/regexp2"
)

type CoreBPE struct {
	encoder              map[string]int
	decoder              map[int]string
	specialTokensEncoder map[string]int
	specialTokensDecoder map[int]string
	tlRegex              *regexp2.Regexp
	tlSpecialRegex       *regexp2.Regexp
	sortedTokenBytes     [][]byte
}

func NewCoreBPE(encoder map[string]int, specialTokensEncoder map[string]int, pattern string) (*CoreBPE, error) {
	regex, err := regexp2.Compile(pattern, regexp2.None)
	if err != nil {
		return nil, fmt.Errorf("error compiling regex: %s", err)
	}

	specialRegexStrs := make([]string, 0, len(specialTokensEncoder))
	for k := range specialTokensEncoder {
		specialRegexStrs = append(specialRegexStrs, regexp.QuoteMeta(k))
	}
	specialRegex, err := regexp2.Compile(strings.Join(specialRegexStrs, "|"), regexp2.None)
	if err != nil {
		return nil, fmt.Errorf("error compiling special regex: %s", err)
	}

	decoder := make(map[int]string, len(encoder))
	for k, v := range encoder {
		decoder[v] = k
	}

	if len(encoder) != len(decoder) {
		return nil, errors.New("encoder and decoder map sizes are different")
	}

	specialTokensDecoder := make(map[int]string, len(specialTokensEncoder))
	for k, v := range specialTokensEncoder {
		specialTokensDecoder[v] = k
	}

	sortedTokenBytes := make([][]byte, 0, len(encoder))
	for k := range encoder {
		sortedTokenBytes = append(sortedTokenBytes, []byte(k))
	}
	sort.Slice(sortedTokenBytes, func(i, j int) bool {
		return bytes.Compare(sortedTokenBytes[i], sortedTokenBytes[j]) < 0
	})

	return &CoreBPE{
		encoder:              encoder,
		specialTokensEncoder: specialTokensEncoder,
		decoder:              decoder,
		specialTokensDecoder: specialTokensDecoder,
		tlRegex:              regex,
		tlSpecialRegex:       specialRegex,
		sortedTokenBytes:     sortedTokenBytes,
	}, nil
}

func (bp *CoreBPE) encodeNative(text string, allowedSpecial map[string]any) ([]int, int) {
	specialRegex := bp.tlSpecialRegex
	regex := bp.tlRegex
	ret := []int{}
	lastPieceTokenLen := 0
	textRunes := []rune(text)

	start := 0
	for {
		var nextSpecial []int
		startFind := start
		for {
			// Find the next allowed special token, if any
			temp := cutRunes(textRunes, startFind, len(textRunes))
			nextSpecial = findRegex2StringIndex(temp, specialRegex)
			if nextSpecial != nil {
				token := cutRunes(textRunes, startFind+nextSpecial[0], startFind+nextSpecial[1])
				if _, ok := allowedSpecial[token]; ok {
					break
				}
				startFind += nextSpecial[1]
			} else {
				break
			}
		}

		end := len([]rune(text))
		if nextSpecial != nil {
			end = start + nextSpecial[0]
		}

		// Okay, here we go, compare this logic to _encode_ordinary_native
		for _, mat := range findRegex2AllStringMatchIndex(cutRunes(textRunes, start, end), regex) {
			piece := cutRunes(textRunes, start+mat[0], start+mat[1])
			if token, ok := bp.encoder[piece]; ok {
				lastPieceTokenLen = 1
				ret = append(ret, token)
				continue
			}
			tokens := bytePairEncode([]byte(piece), bp.encoder)
			lastPieceTokenLen = len(tokens)
			ret = append(ret, tokens...)
		}

		if nextSpecial != nil {
			temp := cutRunes(textRunes, start+nextSpecial[0], start+nextSpecial[1])
			token := bp.specialTokensEncoder[temp]
			ret = append(ret, token)
			start = start + nextSpecial[1]
			lastPieceTokenLen = 0
		} else {
			break
		}
	}

	return ret, lastPieceTokenLen
}

func (bp *CoreBPE) encodeOrdinaryNative(text string) []int {
	ret := []int{}
	textRunes := []rune(text)
	for _, mat := range findRegex2AllStringMatchIndex(text, bp.tlRegex) {
		piece := cutRunes(textRunes, mat[0], mat[1])
		if token, ok := bp.encoder[piece]; ok {
			ret = append(ret, token)
			continue
		}
		tokens := bytePairEncode([]byte(piece), bp.encoder)
		ret = append(ret, tokens...)
	}
	return ret
}

func (bpe *CoreBPE) decodeNative(tokens []int) []byte {
	ret := make([]byte, 0, len(tokens)*2)
	for _, token := range tokens {
		tokenBytes, ok := bpe.decoder[token]
		if !ok {
			tokenBytes = bpe.specialTokensDecoder[token]
		}
		if len(tokenBytes) > 0 {
			ret = append(ret, tokenBytes...)
		}
	}
	return ret
}

func findRegex2StringIndex(text string, reg *regexp2.Regexp) []int {
	m, _ := reg.FindStringMatch(text)
	if m == nil {
		return nil
	}
	result := make([]int, 2)
	result[0] = m.Index
	result[1] = m.Index + m.Length
	return result
}

func findRegex2AllStringMatchIndex(text string, reg *regexp2.Regexp) [][]int {
	var matches [][]int
	m, _ := reg.FindStringMatch(text)
	for m != nil {
		result := make([]int, 2)
		result[0] = m.Index
		result[1] = m.Index + m.Length
		matches = append(matches, result)
		m, _ = reg.FindNextMatch(m)
	}
	return matches
}

func cutRunes(runes []rune, start, end int) string {
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	return string(runes[start:end])
}
