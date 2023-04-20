import { registerHighlightContributions, renderMarkdown as renderMarkdownCommon } from '@sourcegraph/common'

/**
 * Render Markdown to safe HTML.
 *
 * NOTE: This only works when called in an environment with the DOM. In the VS Code extension, it
 * only works in the webview context, not in the extension host context, because the latter lacks a
 * DOM. We could use isomorphic-dompurify for that, but that adds needless complexity for now. If
 * that becomes necessary, we can add that.
 */
export function renderMarkdown(markdown: string): string {
    registerHighlightContributions()

    // Add Cody-specific Markdown rendering if needed.
    return renderMarkdownCommon(markdown, { breaks: true })
}
