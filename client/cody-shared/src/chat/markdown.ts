import { registerHighlightContributions, renderMarkdown as renderMarkdownCommon } from '@sourcegraph/common'

/**
 * Render Markdown to safe HTML.

 * NOTE: This only works when called in an environment with the DOM. In the VS
 * Code extension, it only works in the webview context, not in the extension
 * host context, because the latter lacks a DOM. We could use
 * isomorphic-dompurify for that, but that adds needless complexity for now. If
 * that becomes necessary, we can add that.
 */
export function renderCodyMarkdown(markdown: string): string {
    registerHighlightContributions()

    // Add Cody-specific Markdown rendering if needed.
    return renderMarkdownCommon(markdown, {
        breaks: true,
    })
}

/**
 * Since Cody can write HTML in non-markdown blocks as well and we still want to
 * present this output to the user, we manually scan over all non code block
 * lines and replace HTML starting tags so that they are still visible in the
 * resulting HTML. The final output will still be escaped by DOMPurify so this
 * is only for visual purposes.
 *
 * 🚨 SECURITY: The final output of this should still be piped through
 * `renderCodyMarkdown` to ensure all HTML is sanitized.
 */
export function escapeCodyMarkdown(markdown: string): string {
    const markdownLines = parseMarkdown(markdown)
    const escapedMarkdownLines = markdownLines.map(line => {
        if (line.isCodeBlock) {
            return line.line
        }

        return line.line.replaceAll(/&/g, '&amp;').replaceAll(/</g, '&lt;').replaceAll(/>/g, '&gt;')
    })
    return escapedMarkdownLines.join('\n')
}

export interface MarkdownLine {
    line: string
    isCodeBlock: boolean
}

export function parseMarkdown(text: string): MarkdownLine[] {
    const markdownLines: MarkdownLine[] = []
    let isCodeBlock = false

    for (const line of text.split('\n')) {
        if (line.trim().startsWith('```') || line.trim().startsWith('~~~')) {
            markdownLines.push({ line, isCodeBlock: true })
            isCodeBlock = !isCodeBlock
        } else {
            markdownLines.push({ line, isCodeBlock })
        }
    }
    return markdownLines
}
