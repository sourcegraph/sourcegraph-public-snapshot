package discussions

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"strings"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/pkg/highlight"
)

func formatTargetRepoLinesText(tr *types.DiscussionThreadTargetRepo) string {
	if !tr.HasSelection() {
		return ""
	}

	// Determine the starting line number.
	lineNumber := 1 + int(*tr.StartLine) - len(*tr.LinesBefore)

	var b bytes.Buffer
	padding := len(fmt.Sprint(*tr.EndLine))
	addLines := func(lines []string, prefix string) {
		for _, line := range lines {
			lineNumber++
			fmt.Fprintf(&b, "%s%"+fmt.Sprint(padding)+"d %s\n", prefix, lineNumber, line)
		}
	}
	addLines(*tr.LinesBefore, "  ")
	addLines(*tr.Lines, "> ")
	addLines(*tr.LinesAfter, "  ")
	return b.String()
}

var dataLineMatch = regexp.MustCompile(`\<tr\>\<td class\=\"line\" data\-line\=\"(\d+)\"\>`)

func formatTargetRepoLinesHTML(ctx context.Context, tr *types.DiscussionThreadTargetRepo) (template.HTML, error) {
	if !tr.HasSelection() || tr.Path == nil {
		return "", nil
	}
	fontStyle := `font-family: SFMono-Regular, Consolas, Menlo, DejaVu Sans Mono, monospace; `
	fontStyle += `font-size: 12px; `
	fontStyle += `line-height: 16px; `
	fontStyle += `tab-size: 4; `

	baseStyle := `white-space: pre-wrap; `
	baseStyle += `margin: 0; `
	baseStyle += `padding-left: 0.5rem; `
	baseStyle += `padding-right: 0.5rem; `

	beforeStyle := baseStyle
	beforeStyle += `background-color: #0e121b; `
	beforeStyle += `padding-top: 0.5rem; `
	beforeStyle += `border-top-left-radius: 4px; `
	beforeStyle += `border-top-right-radius: 4px; `

	mainStyle := baseStyle
	mainStyle += `background-color: #1d2535; `

	afterStyle := baseStyle
	afterStyle += `background-color: #0e121b; `
	afterStyle += `padding-bottom: 0.5rem; `
	afterStyle += `border-bottom-left-radius: 4px; `
	afterStyle += `border-bottom-right-radius: 4px; `

	var b bytes.Buffer
	addLines := func(style string, codeLines []string, baseLineNumber int) error {
		if len(codeLines) == 0 {
			return nil
		}
		code := strings.Join(codeLines, "\n")

		// TODO(slimsag): When the file exists (e.g. hasn't been deleted from
		// Git), we can get better highlighting by passing the entire file
		// contents along here.
		isLightTheme := false
		disableTimeout := false
		html, _, err := highlight.Code(ctx, code, *tr.Path, disableTimeout, isLightTheme)
		if err != nil {
			return err
		}

		// TODO(slimsag:discussions): HACK: Replace strings like
		// `<tr><td class="line" data-line="3">` with an actual line number and
		// <tr> element styling. This would be better done with proper HTML
		// parsing -- obviously -- but this works OK for now.
		result := string(html)
		for _, match := range dataLineMatch.FindAllStringSubmatch(string(html), -1) {
			fullMatch, lineNumberText := match[0], match[1]
			lineNumber, _ := strconv.Atoi(lineNumberText)
			newHTML := fmt.Sprintf(`<tr><td style="color: #32405d; padding-right: 1rem; %s">%d</td>`, fontStyle, baseLineNumber+lineNumber)
			result = strings.Replace(result, fullMatch, newHTML, 1)
		}
		result = strings.Replace(result, `<td class="code">`, fmt.Sprintf(`<td style="%s" class="code">`, fontStyle), -1)

		_, err = fmt.Fprintf(&b, `<div style="%s">%s</div>`, style, result)
		return err
	}

	baseLine := 1 + int(*tr.StartLine) - len(*tr.LinesBefore)
	if err := addLines(beforeStyle, *tr.LinesBefore, baseLine); err != nil {
		return "", err
	}
	baseLine += len(*tr.LinesBefore) - 1
	if err := addLines(mainStyle, *tr.Lines, baseLine); err != nil {
		return "", err
	}
	baseLine += len(*tr.Lines) - 1
	if err := addLines(afterStyle, *tr.LinesAfter, baseLine); err != nil {
		return "", err
	}
	return template.HTML(b.String()), nil
}
