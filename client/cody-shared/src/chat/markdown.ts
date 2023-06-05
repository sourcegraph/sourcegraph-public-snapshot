import { marked } from 'marked'

import { registerHighlightContributions, renderMarkdown as renderMarkdownCommon } from '@sourcegraph/common'

const DOMPURIFY_CONFIG = {
    ALLOWED_TAGS: [
        'p',
        'div',
        'span',
        'pre',
        'h1',
        'h2',
        'h3',
        'h4',
        'h5',
        'h6',
        'i',
        'em',
        'b',
        'strong',
        'code',
        'pre',
        'blockquote',
        'ul',
        'li',
        'ol',
        'a',
        'table',
        'tr',
        'th',
        'td',
        'thead',
        'tbody',
        'tfoot',
        's',
        'u',
    ],
    ALLOWED_URI_REGEXP: /^vscode:?\/\/[^\s#$./?].\S*$/i,
}

/**
 * Render Markdown to safe HTML.
 *
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
        dompurifyConfig: DOMPURIFY_CONFIG,
    })
}

/**
 * Returns the parsed markdown at block level.
 */
export function parseMarkdown(text: string): marked.Token[] {
    return marked.Lexer.lex(text, { gfm: true })
}
