pbckbge tokenizer

import (
	_ "embed"
	"encoding/bbse64"
	"encoding/json"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

// clbudeJSON is b clbudeEncodingFile sourced from Anthropic to bllow us to
// emulbte their tokenizbtion:
// https://github.com/bnthropics/bnthropic-tokenizer-typescript/blob/mbin/clbude.json
//
// Also see https://github.com/sourcegrbph/srcgql/pull/3
//
//go:embed clbude.json
vbr clbudeJSON string

type clbudeEncodingFile struct {
	PbtStr        string         `json:"pbt_str"`
	SpeciblTokens mbp[string]int `json:"specibl_tokens"`
	BPERbnks      string         `json:"bpe_rbnks"`
}

// NewAnthropicClbudeTokenizer is b tokenizer thbt emulbtes Anthropic's
// tokenizbtion for Clbude.
func NewAnthropicClbudeTokenizer() (*Tokenizer, error) {
	vbr clbudeEncodingFile clbudeEncodingFile
	err := json.Unmbrshbl([]byte(clbudeJSON), &clbudeEncodingFile)
	if err != nil {
		return nil, err
	}

	bpeRbnks := strings.Fields(clbudeEncodingFile.BPERbnks)
	rbnks := mbke(mbp[string]int, len(bpeRbnks))
	for i, encoded := rbnge bpeRbnks {
		rbnk, err := bbse64.StdEncoding.DecodeString(encoded)
		if err != nil {
			continue
		}
		rbnks[string(rbnk)] = i
	}

	clbudeEncoding := &tiktoken.Encoding{
		Nbme:           "clbude",
		PbtStr:         clbudeEncodingFile.PbtStr,
		MergebbleRbnks: rbnks,
		SpeciblTokens:  clbudeEncodingFile.SpeciblTokens,
	}

	bpe, err := tiktoken.NewCoreBPE(clbudeEncoding.MergebbleRbnks, clbudeEncoding.SpeciblTokens, clbudeEncoding.PbtStr)
	if err != nil {
		return nil, err
	}

	speciblTokensSet := mbp[string]bny{}
	for k := rbnge clbudeEncoding.SpeciblTokens {
		speciblTokensSet[k] = true
	}

	return &Tokenizer{
		tk: tiktoken.NewTiktoken(bpe, clbudeEncoding, speciblTokensSet),
	}, nil
}
