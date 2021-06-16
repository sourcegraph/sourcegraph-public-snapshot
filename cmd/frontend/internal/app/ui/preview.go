package ui

import (
	"fmt"
	"path"
	"strconv"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

var (
	singleLineRegexp     = lazyregexp.New(`^L(\d+)(?::\d+)?$`)
	multiLineRangeRegexp = lazyregexp.New(`^L(\d+)(?::\d+)?-(\d+)(?::\d+)?$`)
)

type lineRange struct {
	StartLine int
	EndLine   int
}

func findLineRangeInQueryParameters(queryParameters map[string][]string) *lineRange {
	for key := range queryParameters {
		if lineRange := getLineRange(key); lineRange != nil {
			return lineRange
		}
	}
	return nil
}

func getLineRange(value string) *lineRange {
	var startLine, endLine int
	if submatches := multiLineRangeRegexp.FindStringSubmatch(value); submatches != nil {
		startLine, _ = strconv.Atoi(submatches[1])
		endLine, _ = strconv.Atoi(submatches[2])
		return &lineRange{StartLine: startLine, EndLine: endLine}
	} else if submatches := singleLineRegexp.FindStringSubmatch(value); submatches != nil {
		startLine, _ = strconv.Atoi(submatches[1])
		return &lineRange{StartLine: startLine}
	}
	return nil
}

func formatLineRange(lineRange *lineRange) string {
	if lineRange == nil {
		return ""
	}

	formattedLineRange := ""
	if lineRange.StartLine != 0 && lineRange.EndLine != 0 {
		formattedLineRange = fmt.Sprintf("L%d-%d", lineRange.StartLine, lineRange.EndLine)
	} else if lineRange.StartLine != 0 {
		formattedLineRange = fmt.Sprintf("L%d", lineRange.StartLine)
	}
	return formattedLineRange
}

func getBlobPreviewImageURL(previewServiceURL string, blobURLPath string, lineRange *lineRange) string {
	blobPreviewImageURL := previewServiceURL + blobURLPath
	formattedLineRange := formatLineRange(lineRange)
	if formattedLineRange != "" {
		blobPreviewImageURL += "?range=" + formattedLineRange
	}
	return blobPreviewImageURL
}

func getBlobPreviewTitle(blobFilePath string, lineRange *lineRange) string {
	blobPreviewTitle := path.Base(blobFilePath)
	formattedLineRange := formatLineRange(lineRange)
	if formattedLineRange != "" {
		blobPreviewTitle += "#" + formattedLineRange
	}
	return blobPreviewTitle
}
