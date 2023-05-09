import { marked } from 'marked'

import { registerHighlightContributions, renderMarkdown as renderMarkdownCommon } from '@sourcegraph/common'

const LEXER_OPTIONS = { gfm: true }

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
    })
}

/**
 * Since Cody can write HTML in non-markdown blocks as well and we still want to
 * present this output to the user, we manually scan over all non code block
 * lines and replace HTML starting tags so that they are still visible in the
 * resulting HTML. The finale output will still be escaped by DOMPurify so this
 * is only for visual purpose.
 *
 * 🚨 SECURITY: The final output of this should still be piped through
 * `renderCodyMarkdown` and in turn DOMPurify to ensure all HTML is sanitized
 * from malicious code.
 */
export function escapeCodyMarkdown(markdown: string, isStreaming: boolean): string {
    // The marked lexer seems to behave unexpectedly when the full message is
    // not received yet (e.g. when a code block starts but it's not closed yet).
    //
    // Hence we only escape the markdown if we have the full message.
    if (isStreaming) {
        return markdown
    }

    const tokens = parseMarkdown(markdown)
    const escapedTokens = tokens.map(token => {
        switch (token.type) {
            case 'code':
            case 'codespan':
                return token
            default:
                return {
                    ...token,
                    // To detect inline code blocks, we need to parse every block
                    // level token again and check if it contains inline
                    raw: parseMarkdownInline(token.raw)
                        .map(token => {
                            if (token.type === 'codespan') {
                                return token.raw
                            }

                            return token.raw.replaceAll(/&/g, '&amp;').replaceAll(/</g, '&lt;').replaceAll(/>/g, '&gt;')
                        })
                        .join(''),
                }
        }
    })
    return escapedTokens.map(token => token.raw).join('')
}

/**
 * Returns the parsed markdown at block level.
 */
export function parseMarkdown(text: string): marked.Token[] {
    return marked.Lexer.lex(text, LEXER_OPTIONS)
}

/**
 * Returns the parsed markdown at the inline level. This is useful for finding inline code blocks
 */
function parseMarkdownInline(text: string): marked.Token[] {
    return marked.Lexer.lexInline(text, LEXER_OPTIONS)
}
