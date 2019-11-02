import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import { without } from 'lodash'
// This is the only file allowed to import this module, all other modules must use renderMarkdown() exported from here
// eslint-disable-next-line no-restricted-imports
import marked from 'marked'
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
 * @returns Safe HTML
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
 *
 * @param markdown The markdown to render
 * @param options Options object for passing additional parameters
 */
export const renderMarkdown = (
    markdown: string,
    options: {
        /** Strip off any HTML and return a plain text string, useful for previews */

        plainText?: boolean
    } = {}
): string => {
    const rendered = marked(markdown, {
        gfm: true,
        breaks: true,
        sanitize: false,
        highlight: (code, language) => highlightCodeSafe(code, language),
    })
    return sanitize(
        rendered,
        options.plainText
            ? { allowedTags: [], allowedAttributes: {} }
            : {
                  // Defaults: https://sourcegraph.com/github.com/punkave/sanitize-html@90aac2665011be6fa21a8864d21c604ee984294f/-/blob/src/index.js#L571-589

                  // Allow highligh.js styles, e.g.
                  // <span class="hljs-keyword">
                  // <code class="language-javascript">
                  allowedTags: [...without(sanitize.defaults.allowedTags, 'iframe'), 'h1', 'h2', 'span', 'img'],
                  allowedAttributes: {
                      ...sanitize.defaults.allowedAttributes,
                      span: ['class'],
                      code: ['class'],
                      h1: ['id'],
                      h2: ['id'],
                      h3: ['id'],
                      h4: ['id'],
                      h5: ['id'],
                      h6: ['id'],
                  },
              }
    )
}
