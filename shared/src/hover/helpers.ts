import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import marked from 'marked'
import * as React from 'react'
import sanitize from 'sanitize-html'

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
 * @return Safe HTML
 */
export const highlightCodeSafe = (code: string, language?: string): string => {
    try {
        if (language === 'plaintext' || language === 'text') {
            return escapeHTML(code)
        }
        if (language) {
            return highlight(language, code, true).value
        }
        return highlightAuto(code).value
    } catch (err) {
        console.warn('Error syntax-highlighting hover markdown code block', err)
        return escapeHTML(code)
    }
}

/**
 * Renders the given markdown to HTML, highlighting code and sanitizing dangerous HTML.
 * Can throw an exception on parse errors.
 */
export const renderMarkdown = (markdown: string): string => {
    const rendered = marked(markdown, {
        gfm: true,
        breaks: true,
        sanitize: false,
        highlight: (code, language) => highlightCodeSafe(code, language),
    })
    return sanitize(rendered, {
        // Allow highligh.js styles, e.g.
        // <span class="hljs-keyword">
        // <code class="language-javascript">
        allowedTags: [...sanitize.defaults.allowedTags, 'span'],
        allowedAttributes: {
            span: ['class'],
            code: ['class'],
        },
    })
}

/**
 * Converts a synthetic React event to a persisted, native Event object.
 *
 * @param event The synthetic React event object
 */
export const toNativeEvent = <E extends React.SyntheticEvent<T>, T>(event: E): E['nativeEvent'] => {
    event.persist()
    return event.nativeEvent
}
