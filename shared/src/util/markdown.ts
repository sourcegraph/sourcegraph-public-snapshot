import { highlight, highlightAuto } from 'highlight.js/lib/highlight'
import { without } from 'lodash'
// This is the only file allowed to import this module, all other modules must use renderMarkdown() exported from here
// eslint-disable-next-line no-restricted-imports
import marked from 'marked'
import sanitize from 'sanitize-html'
import { Overwrite } from 'utility-types'

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
    options:
        | {
              /** Strip off any HTML and return a plain text string, useful for previews */
              plainText: true
          }
        | {
              /** Following options apply to non-plaintext output only. */
              plainText?: false

              /** Allow links to data: URIs and that download files. */
              allowDataUriLinksAndDownloads?: boolean
          } = {}
): string => {
    const rendered = marked(markdown, {
        gfm: true,
        breaks: false,
        sanitize: false,
        highlight: (code, language) => highlightCodeSafe(code, language),
    })

    let opt: Overwrite<sanitize.IOptions, sanitize.IDefaults>
    if (options.plainText) {
        opt = { allowedAttributes: {}, allowedSchemes: [], allowedSchemesByTag: {}, allowedTags: [], selfClosing: [] }
    } else {
        opt = {
            ...sanitize.defaults,
            // Ensure <object> must have type attribute set
            exclusiveFilter: ({ tag, attribs }) => tag === 'object' && !attribs.type,

            // Allow highligh.js styles, e.g.
            // <span class="hljs-keyword">
            // <code class="language-javascript">
            allowedTags: [
                ...without(sanitize.defaults.allowedTags, 'iframe'),
                'h1',
                'h2',
                'span',
                'img',
                'object',
                'svg',
                'rect',
                'title',
            ],
            allowedAttributes: {
                ...sanitize.defaults.allowedAttributes,
                a: [
                    ...sanitize.defaults.allowedAttributes.a,
                    'title',
                    'data-tooltip', // TODO support fancy tooltips through native titles
                ],
                object: ['data', { name: 'type', values: ['image/svg+xml'] }, 'width'],
                svg: ['width', 'height', 'viewbox', 'version'],
                rect: ['x', 'y', 'width', 'height', 'fill', 'stroke', 'stroke-width'],
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
        if (options.allowDataUriLinksAndDownloads) {
            opt.allowedAttributes.a = [...opt.allowedAttributes.a, 'download']
            opt.allowedSchemesByTag = {
                ...opt.allowedSchemesByTag,
                a: [...(opt.allowedSchemesByTag.a || opt.allowedSchemes), 'data'],
            }
        }
    }

    return sanitize(rendered, opt)
}
