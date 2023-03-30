import { registerHighlightContributions, renderMarkdown as renderMarkdownCommon } from '@sourcegraph/common'

/**
 * Render Markdown to safe HTML.
 *
 * NOTE: This only works when called in an environment with the DOM. In the VS Code extension, it
 * only works in the webview context, not in the extension host context, because the latter lacks a
 * DOM. We could use isomorphic-dompurify for that, but that adds needless complexity for now. If
 * that becomes necessary, we can add that.
 */
export function renderMarkdown(markdown: string, options: { speaker: 'human' | 'assistant' }): string {
    registerHighlightContributions()

    let html = renderMarkdownCommon(markdown)

    // Add Cody-specific Markdown for code blocks
    if (options.speaker === 'assistant') {
        html = html.replaceAll(
            '<pre><code',
            '<pre><button class="copy-code-button" onclick="navigator.clipboard.writeText(event.target.nextSibling.textContent).then(() => event.target.textContent = \'Copied\')">Copy</button><code'
        )
    }

    return html
}
