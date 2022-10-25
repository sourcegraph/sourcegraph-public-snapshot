package ui

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

var (
	singleLineRegexp     = lazyregexp.New(`^L(\d+)(:\d+)?$`)
	multiLineRangeRegexp = lazyregexp.New(`^L(\d+)(:\d+)?-(\d+)(:\d+)?$`)
)

type lineRange struct {
	StartLine          int
	StartLineCharacter int
	EndLine            int
	EndLineCharacter   int
}

func FindLineRangeInQueryParameters(queryParameters map[string][]string) *lineRange {
	for key := range queryParameters {
		if lineRange := getLineRange(key); lineRange != nil {
			return lineRange
		}
	}
	return nil
}

func parseCharacterMatch(characterMatch string) int {
	var character int
	if characterMatch != "" {
		character, _ = strconv.Atoi(strings.TrimLeft(characterMatch, ":"))
	}
	return character
}

func getLineRange(value string) *lineRange {
	var startLine, startLineCharacter, endLine, endLineCharacter int
	if submatches := multiLineRangeRegexp.FindStringSubmatch(value); submatches != nil {
		startLine, _ = strconv.Atoi(submatches[1])
		startLineCharacter = parseCharacterMatch(submatches[2])
		endLine, _ = strconv.Atoi(submatches[3])
		endLineCharacter = parseCharacterMatch(submatches[4])
		return &lineRange{StartLine: startLine, StartLineCharacter: startLineCharacter, EndLine: endLine, EndLineCharacter: endLineCharacter}
	} else if submatches := singleLineRegexp.FindStringSubmatch(value); submatches != nil {
		startLine, _ = strconv.Atoi(submatches[1])
		startLineCharacter = parseCharacterMatch(submatches[2])
		return &lineRange{StartLine: startLine, StartLineCharacter: startLineCharacter}
	}
	return nil
}

func formatCharacter(character int) string {
	if character == 0 {
		return ""
	}
	return fmt.Sprintf(":%d", character)
}

func FormatLineRange(lineRange *lineRange) string {
	if lineRange == nil {
		return ""
	}

	formattedLineRange := ""
	if lineRange.StartLine != 0 && lineRange.EndLine != 0 {
		formattedLineRange = fmt.Sprintf("L%d%s-%d%s", lineRange.StartLine, formatCharacter(lineRange.StartLineCharacter), lineRange.EndLine, formatCharacter(lineRange.EndLineCharacter))
	} else if lineRange.StartLine != 0 {
		formattedLineRange = fmt.Sprintf("L%d%s", lineRange.StartLine, formatCharacter(lineRange.StartLineCharacter))
	}
	return formattedLineRange
}

func getBlobPreviewImageURL(previewServiceURL string, blobURLPath string, lineRange *lineRange) string {
	blobPreviewImageURL := previewServiceURL + blobURLPath
	formattedLineRange := FormatLineRange(lineRange)

	queryValues := url.Values{}
	if formattedLineRange != "" {
		queryValues.Add("range", formattedLineRange)
	}

	encodedQueryValues := queryValues.Encode()
	if encodedQueryValues != "" {
		encodedQueryValues = "?" + encodedQueryValues
	}

	return blobPreviewImageURL + encodedQueryValues
}

func getBlobPreviewTitle(blobFilePath string, lineRange *lineRange, symbolResult *result.Symbol) string {
	formattedLineRange := FormatLineRange(lineRange)
	formattedBlob := path.Base(blobFilePath)
	if formattedLineRange != "" {
		formattedBlob += "?" + formattedLineRange
	}
	if symbolResult != nil {
		return fmt.Sprintf("%s %s (%s)", symbolResult.LSPKind().String(), symbolResult.Name, formattedBlob)
	}
	return formattedBlob
}
