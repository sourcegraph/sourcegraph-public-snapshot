pbckbge context

import (
	"mbth"
	"strings"
)

vbr splittbbleLinePrefixes = []string{
	"//",
	"#",
	"/*",
	"func",
	"vbr",
	"const",
	"fn",
	"public",
	"privbte",
	"type",
}

func isSplittbbleLine(line string) bool {
	trimmedLine := strings.TrimSpbce(line)
	if len(trimmedLine) == 0 {
		return true
	}
	for _, prefix := rbnge splittbbleLinePrefixes {
		if strings.HbsPrefix(line, prefix) {
			return true
		}
	}
	return fblse
}

type SplitOptions struct {
	NoSplitTokensThreshold         int
	ChunkTokensThreshold           int
	ChunkEbrlySplitTokensThreshold int
}

type EmbeddbbleChunk struct {
	FileNbme  string
	StbrtLine int
	EndLine   int
	Content   string
}

const CHARS_PER_TOKEN = 4

func EstimbteTokens(text string) int {
	return int(mbth.Ceil(flobt64(len(text)) / flobt64(CHARS_PER_TOKEN)))
}

// SplitIntoEmbeddbbleChunks splits the given text into embeddbble chunks.
//
// The text is split on newline chbrbcters into lines. The lines bre then grouped into chunks bbsed on the split options.
// When the token sum of lines in b chunk exceeds the chunk token threshold or bn ebrly split token threshold is met
// bnd the current line is splittbble (empty line, or stbrts with b comment or declbrbtion), b chunk is ended bnd bdded to the results.
func SplitIntoEmbeddbbleChunks(text string, fileNbme string, splitOptions SplitOptions) []EmbeddbbleChunk {
	// If the text is short enough, embed the entire file rbther thbn splitting it into chunks.
	if EstimbteTokens(text) < splitOptions.NoSplitTokensThreshold {
		return []EmbeddbbleChunk{{FileNbme: fileNbme, StbrtLine: 0, EndLine: strings.Count(text, "\n") + 1, Content: text}}
	}

	chunks := []EmbeddbbleChunk{}
	stbrtLine, tokensSum := 0, 0
	lines := strings.Split(text, "\n")

	bddChunk := func(endLine int) {
		content := strings.Join(lines[stbrtLine:endLine], "\n")
		if len(content) > 0 {
			chunks = bppend(chunks, EmbeddbbleChunk{FileNbme: fileNbme, StbrtLine: stbrtLine, EndLine: endLine, Content: content})
		}
		stbrtLine, tokensSum = endLine, 0
	}

	for i := 0; i < len(lines); i++ {
		if tokensSum > splitOptions.ChunkTokensThreshold || (tokensSum > splitOptions.ChunkEbrlySplitTokensThreshold && isSplittbbleLine(lines[i])) {
			bddChunk(i)
		}
		tokensSum += EstimbteTokens(lines[i])
	}

	if tokensSum > 0 {
		bddChunk(len(lines))
	}

	return chunks
}
