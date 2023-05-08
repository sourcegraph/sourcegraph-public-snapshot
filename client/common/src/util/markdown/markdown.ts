import DOMPurify, { Config as DOMPurifyConfig } from 'dompurify'
import { highlight, highlightAuto } from 'highlight.js/lib/core'
// This is the only file allowed to import this module, all other modules must use renderMarkdown() exported from here
// eslint-disable-next-line no-restricted-imports
import { marked } from 'marked'

import { logger } from '../logger'

/**
 * Escapes HTML by replacing characters like `<` with their HTML escape sequences like `&lt;`
 */
const escapeHTML = (html: string): string => {
    const span = document.createElement('span')
    span.textContent = html
    return span.innerHTML
}

/**
 * Attempts to syntax-highlight the given code.
 * If the language is not given, it is auto-detected.
 * If an error occurs, the code is returned as plain text with escaped HTML entities
 *
 * @param code The code to highlight
 * @param language The language of the code, if known
 * @returns Safe HTML
 */
export const highlightCodeSafe = (code: string, language?: string): string => {
    try {
        if (language === 'plaintext' || language === 'text') {
            return escapeHTML(code)
        }
        if (language === 'sourcegraph') {
            return code
        }
        if (language) {
            return highlight(code, { language, ignoreIllegals: true }).value
        }
        return highlightAuto(code).value
    } catch (error) {
        logger.warn('Error syntax-highlighting hover markdown code block', error)
        return escapeHTML(code)
    }
}

/**
 * Renders the given markdown to HTML, highlighting code and sanitizing dangerous HTML.
 * Can throw an exception on parse errors.
 *
 * @param markdown The markdown to render
 */
export const renderMarkdown = (
    markdown: string,
    options: {
        /** Whether to render line breaks as HTML `<br>`s */
        breaks?: boolean
        /** Whether to disable autolinks. Explicit links using `[text](url)` are still allowed. */
        disableAutolinks?: boolean
        renderer?: marked.Renderer
        headerPrefix?: string
        /** Strip off any HTML and return a plain text string, useful for previews */
        plainText?: boolean
    } = {}
): string => {
    const tokenizer = new marked.Tokenizer()
    if (options.disableAutolinks) {
        // Why the odd double-casting below?
        // Because returning undefined is the recommended way to easily disable autolinks
        // but the type definition does not allow it.
        // More context here: https://github.com/markedjs/marked/issues/882
        tokenizer.url = () => undefined as unknown as marked.Tokens.Link
    }

    const rendered = marked(markdown, {
        gfm: true,
        breaks: options.breaks,
        highlight: (code, language) => highlightCodeSafe(code, language),
        renderer: options.renderer,
        headerPrefix: options.headerPrefix ?? '',
        tokenizer,
    })

    const dompurifyConfig: DOMPurifyConfig & { RETURN_DOM_FRAGMENT?: false; RETURN_DOM?: false } = options.plainText
        ? {
              ALLOWED_TAGS: [],
              ALLOWED_ATTR: [],
              KEEP_CONTENT: true,
          }
        : {
              USE_PROFILES: { html: true },
              FORBID_ATTR: ['rel', 'style'],
          }

    return DOMPurify.sanitize(rendered, dompurifyConfig).trim()
}

export const markdownLexer = (markdown: string): marked.TokensList => marked.lexer(markdown)
