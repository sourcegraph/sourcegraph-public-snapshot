import { highlightBlock, registerLanguage } from 'highlight.js/lib/highlight'
import * as _ from 'lodash'
import marked from 'marked'
import { MarkupContent } from 'sourcegraph'
import { Key } from 'ts-key-enum'
import { AbsoluteRepoFile, AbsoluteRepoFilePosition, parseBrowserRepoURL } from '.'
import { HoverMerged } from '../../../../../shared/src/api/client/types/hover'
import { getModeFromPath } from '../../../../../shared/src/languages'
import { makeCloseIcon, makeSourcegraphIcon } from '../components/Icons'
import { sourcegraphUrl } from '../util/context'
import { toAbsoluteBlobURL } from '../util/url'

registerLanguage('go', require('highlight.js/lib/languages/go'))
registerLanguage('javascript', require('highlight.js/lib/languages/javascript'))
registerLanguage('typescript', require('highlight.js/lib/languages/typescript'))
registerLanguage('java', require('highlight.js/lib/languages/java'))
registerLanguage('python', require('highlight.js/lib/languages/python'))
registerLanguage('php', require('highlight.js/lib/languages/php'))
registerLanguage('bash', require('highlight.js/lib/languages/bash'))
registerLanguage('clojure', require('highlight.js/lib/languages/clojure'))
registerLanguage('cpp', require('highlight.js/lib/languages/cpp'))
registerLanguage('cs', require('highlight.js/lib/languages/cs'))
registerLanguage('css', require('highlight.js/lib/languages/css'))
registerLanguage('dockerfile', require('highlight.js/lib/languages/dockerfile'))
registerLanguage('elixir', require('highlight.js/lib/languages/elixir'))
registerLanguage('haskell', require('highlight.js/lib/languages/haskell'))
registerLanguage('html', require('highlight.js/lib/languages/xml'))
registerLanguage('lua', require('highlight.js/lib/languages/lua'))
registerLanguage('ocaml', require('highlight.js/lib/languages/ocaml'))
registerLanguage('r', require('highlight.js/lib/languages/r'))
registerLanguage('ruby', require('highlight.js/lib/languages/ruby'))
registerLanguage('rust', require('highlight.js/lib/languages/rust'))
registerLanguage('swift', require('highlight.js/lib/languages/swift'))

let tooltip: HTMLElement
let loadingTooltip: HTMLElement
let tooltipActions: HTMLElement
let j2dAction: HTMLAnchorElement
let findRefsAction: HTMLAnchorElement
let moreContext: HTMLElement

// tslint:disable-next-line:max-line-length prettier
const referencesIconSVG =
    '<svg width="12px" height="8px"><path fill="#24292e" xmlns="http://www.w3.org/2000/svg" id="path15_fill" d="M 6.00625 8C 2.33125 8 0.50625 5.075 0.05625 4.225C -0.01875 4.075 -0.01875 3.9 0.05625 3.775C 0.50625 2.925 2.33125 0 6.00625 0C 9.68125 0 11.5063 2.925 11.9563 3.775C 12.0312 3.925 12.0312 4.1 11.9563 4.225C 11.5063 5.075 9.68125 8 6.00625 8ZM 6.00625 1.25C 4.48125 1.25 3.25625 2.475 3.25625 4C 3.25625 5.525 4.48125 6.75 6.00625 6.75C 7.53125 6.75 8.75625 5.525 8.75625 4C 8.75625 2.475 7.53125 1.25 6.00625 1.25ZM 6.00625 5.75C 5.03125 5.75 4.25625 4.975 4.25625 4C 4.25625 3.025 5.03125 2.25 6.00625 2.25C 6.98125 2.25 7.75625 3.025 7.75625 4C 7.75625 4.975 6.98125 5.75 6.00625 5.75Z"/></svg>'
// tslint:disable-next-line:max-line-length prettier
const definitionIconSVG =
    '<svg width="11px" height="9px"><path fill="#24292e" xmlns="http://www.w3.org/2000/svg" id="path10_fill" d="M 6.325 8.4C 6.125 8.575 5.8 8.55 5.625 8.325C 5.55 8.25 5.5 8.125 5.5 8L 5.5 6C 2.95 6 1.4 6.875 0.825 8.7C 0.775 8.875 0.6 9 0.425 9C 0.2 9 -4.44089e-16 8.8 -4.44089e-16 8.575C -4.44089e-16 8.575 -4.44089e-16 8.575 -4.44089e-16 8.55C 0.125 4.825 1.925 2.675 5.5 2.5L 5.5 0.5C 5.5 0.225 5.725 8.88178e-16 6 8.88178e-16C 6.125 8.88178e-16 6.225 0.05 6.325 0.125L 10.825 3.875C 11.025 4.05 11.075 4.375 10.9 4.575C 10.875 4.6 10.85 4.625 10.825 4.65L 6.325 8.4Z"/></svg>'

export interface TooltipData extends Partial<HoverMerged> {
    target: HTMLElement
    ctx: AbsoluteRepoFilePosition
    defUrl?: string
    asyncDefUrl?: boolean
    loading?: boolean
}

/**
 * createTooltips initializes the DOM elements used for the hover
 * tooltip and "Loading..." text indicator, adding the former
 * to the DOM (but hidden). It is idempotent.
 */
export function createTooltips(): void {
    const mountedTooltip = document.querySelector('.sg-tooltip')
    if (mountedTooltip) {
        if (tooltip) {
            return
        }
        mountedTooltip.remove()
    }
    tooltip = document.createElement('DIV')
    tooltip.classList.add('sg-tooltip')
    tooltip.style.visibility = 'hidden'
    document.body.appendChild(tooltip)

    loadingTooltip = document.createElement('DIV')
    loadingTooltip.appendChild(document.createTextNode('Loading...'))
    loadingTooltip.className = 'sg-tooltip__loading'

    tooltipActions = document.createElement('DIV')
    tooltipActions.className = 'sg-tooltip__actions'

    moreContext = document.createElement('DIV')
    moreContext.className = 'sg-tooltip__more-actions'
    moreContext.appendChild(document.createTextNode('Click for more actions...'))

    const definitionIcon = document.createElement('svg')
    definitionIcon.innerHTML = definitionIconSVG
    definitionIcon.className = 'sg-tooltip__definition-icon'

    j2dAction = document.createElement('A') as HTMLAnchorElement
    j2dAction.appendChild(definitionIcon)
    j2dAction.appendChild(document.createTextNode('Go to definition'))
    j2dAction.className = 'btn btn-sm BtnGroup-item sg-tooltip__action'

    const referencesIcon = document.createElement('svg')
    referencesIcon.innerHTML = referencesIconSVG
    referencesIcon.className = 'sg-tooltip__references-icon'

    findRefsAction = document.createElement('A') as HTMLAnchorElement
    findRefsAction.appendChild(referencesIcon)
    findRefsAction.appendChild(document.createTextNode('Find references'))
    findRefsAction.className = 'btn btn-sm BtnGroup-item sg-tooltip__action'

    tooltipActions.appendChild(j2dAction)
    tooltipActions.appendChild(findRefsAction)
}

function constructBaseTooltip(): void {
    tooltip.appendChild(loadingTooltip)
    tooltip.appendChild(moreContext)
    tooltip.appendChild(tooltipActions)
}

// This global refers to the element the tooltip should be positioned relative to.
// We keep a reference b/c we must update the x-coordinate of the tooltip on window resize.
let tooltipTarget: HTMLElement | undefined

/**
 * hideTooltip makes the tooltip on the DOM invisible.
 */
export function hideTooltip(): void {
    const previousTarget = tooltipTarget
    tooltipTarget = undefined
    if (!tooltip) {
        return
    }
    if (previousTarget) {
        previousTarget.classList.remove('selection-highlight-sticky')
    }

    while (tooltip.firstChild) {
        tooltip.removeChild(tooltip.firstChild)
    }
    tooltip.style.visibility = 'hidden' // prevent black dot of empty content
    tooltip.classList.remove('docked')

    document.dispatchEvent(new CustomEvent<{}>('sourcegraph:dismissTooltip', {}))
}

export function isTooltipVisible(ctx: AbsoluteRepoFile, isBase: boolean): boolean {
    if (!tooltip) {
        return false
    }
    if (isBase) {
        return (
            tooltip.classList.contains('tooltip-base') &&
            tooltip.classList.contains(ctx.filePath) &&
            tooltip.style.visibility === 'visible'
        )
    }
    return (
        !tooltip.classList.contains('tooltip-base') &&
        tooltip.classList.contains(ctx.filePath) &&
        tooltip.style.visibility === 'visible'
    )
}

export function isOtherFileTooltipVisible(ctx: AbsoluteRepoFile): boolean {
    return !tooltip.classList.contains(ctx.filePath) && tooltip.style.visibility === 'visible'
}

interface Actions {
    definition: (ctx: AbsoluteRepoFilePosition) => (e: MouseEvent) => void
    references: (ctx: AbsoluteRepoFilePosition) => (e: MouseEvent) => void
    dismiss: () => void
}

/**
 * updateTooltip displays the appropriate tooltip given current state (and may hide
 * the tooltip if no text is available).
 */
export function updateTooltip(data: TooltipData, docked: boolean, actions: Actions, isBase: boolean): void {
    if (isTooltipVisible(data.ctx, isBase)) {
        hideTooltip() // hide before updating tooltip text
    }

    const { loading, target, ctx } = data
    if (!target) {
        // no target to show hover for; tooltip is hidden
        return
    }
    tooltipTarget = target
    tooltip.className = ''
    tooltip.classList.add('sg-tooltip')

    tooltip.classList.add(data.ctx.filePath)

    if (isBase) {
        tooltip.classList.add('tooltip-base')
    }

    constructBaseTooltip()
    loadingTooltip.style.display = loading ? 'block' : 'none'
    moreContext.style.display = docked || loading ? 'none' : 'flex'
    tooltipActions.style.display = docked ? 'flex' : 'none'

    j2dAction.style.display = 'block'
    j2dAction.href = data.defUrl ? new URL(data.defUrl, sourcegraphUrl).href : ''

    if (data.asyncDefUrl) {
        j2dAction.style.cursor = 'pointer'
        j2dAction.onclick = actions.definition(data.ctx)
    } else if (data.defUrl && j2dAction.href !== window.location.href) {
        j2dAction.style.cursor = 'pointer'
        j2dAction.onclick = actions.definition(parseBrowserRepoURL(data.defUrl) as AbsoluteRepoFilePosition)
    } else {
        j2dAction.style.cursor = 'not-allowed'
        j2dAction.onclick = () => false
    }

    findRefsAction.style.display = 'block'
    findRefsAction.href = ctx ? toAbsoluteBlobURL({ ...ctx, referencesMode: 'local' }) : ''
    findRefsAction.onclick = actions.references(ctx)

    if (!data.loading) {
        loadingTooltip.style.visibility = 'hidden'

        if (!data.contents) {
            return
        }
        type MarkedString =
            | string
            | {
                  language: string
                  value: string
              }
        const contentsArray: (MarkupContent | MarkedString)[] = Array.isArray(data.contents)
            ? data.contents
            : ([data.contents] as (MarkupContent | MarkedString)[])
        if (contentsArray.length === 0) {
            return
        }
        const firstContent = contentsArray[0]
        const title: string = typeof firstContent === 'string' ? firstContent : firstContent.value
        let doc: string | undefined
        for (const content of contentsArray.slice(1)) {
            if (typeof content === 'string') {
                doc = content
            } else if ('language' in content && content.language === 'markdown') {
                doc = content.value
            } else if ('kind' in content && content.kind === 'markdown') {
                doc = content.value
            }
        }
        if (!title) {
            // no tooltip text / search context; tooltip is hidden
            return
        }

        const container = document.createElement('DIV')
        container.className = 'sg-tooltip__title-container'

        const tooltipText = document.createElement('DIV')
        tooltipText.className = `${getModeFromPath(ctx.filePath)} sg-tooltip__title`
        tooltipText.appendChild(document.createTextNode(title))

        const icon = makeSourcegraphIcon()
        icon.className = icon.className + ' sg-tooltip__sg-icon'

        container.appendChild(icon)
        container.appendChild(tooltipText)
        tooltip.insertBefore(container, moreContext)

        const closeContainer = document.createElement('span')
        closeContainer.className = 'sg-tooltip__close-icon'
        closeContainer.onclick = actions.dismiss

        if (docked) {
            const closeIcon = makeCloseIcon()
            closeContainer.appendChild(closeIcon)
            container.appendChild(closeContainer)
            tooltip.classList.add('docked')
        } else {
            tooltip.classList.remove('docked')
        }

        highlightBlock(tooltipText)

        if (doc) {
            const tooltipDoc = document.createElement('DIV')
            tooltipDoc.className = 'sg-tooltip__doc'
            tooltipDoc.innerHTML = marked(doc, { gfm: true, breaks: true, sanitize: true })
            tooltip.insertBefore(tooltipDoc, moreContext)

            // Handle scrolling ourselves so that scrolling to the bottom of
            // the tooltip documentation does not cause the page to start
            // scrolling (which is a very jarring experience).
            tooltip.addEventListener(
                'wheel',
                (e: WheelEvent) => {
                    tooltipDoc.scrollTop += e.deltaY
                },
                { passive: true } as any
            )
        }
    } else {
        loadingTooltip.style.visibility = 'visible'
    }

    // Anchor it horizontally, prior to rendering to account for wrapping
    // changes to vertical height if the tooltip is at the edge of the viewport.
    const targetBound = target.getBoundingClientRect()
    tooltip.style.left = targetBound.left + window.scrollX + 'px'

    // Anchor the tooltip vertically.
    const tooltipBound = tooltip.getBoundingClientRect()
    const relTop = targetBound.top + window.scrollY
    const margin = 5
    let tooltipTop = relTop - (tooltipBound.height + margin)
    if (tooltipTop - window.scrollY < 0) {
        // Tooltip wouldn't be visible from the top, so display it at the
        // bottom.
        const relBottom = targetBound.bottom + window.scrollY
        tooltipTop = relBottom + margin
    }
    tooltip.style.top = tooltipTop + 'px'
    // Make it all visible to the user.
    tooltip.style.visibility = 'visible'
}

window.addEventListener('keyup', (e: KeyboardEvent) => {
    if (e.key === Key.Escape) {
        hideTooltip()
    }
})

if (document.readyState === 'complete' || document.readyState === 'interactive') {
    addDocumentClickListener()
} else {
    window.addEventListener('load', () => {
        addDocumentClickListener()
    })
}

function addDocumentClickListener(): void {
    document.body.addEventListener(
        'click',
        (e: MouseEvent) => {
            if (!isInsideCodeContainer(e.target as HTMLElement)) {
                hideTooltip()
            }
        },
        { passive: true } as any
    )
}

window.addEventListener(
    'resize',
    () => {
        if (tooltipTarget) {
            const targetBound = tooltipTarget.getBoundingClientRect()
            tooltip.style.left = targetBound.left + window.scrollX + 'px'
        }
    },
    { passive: true } as any
)

/**
 * convertNode modifies a DOM node so that we can identify precisely token a user has clicked or hovered over.
 * On a code view, source code is typically wrapped in a HTML table cell. It may look like this:
 *
 *     <td id="LC18" class="blob-code blob-code-inner js-file-line">
 *        <#textnode>\t</#textnode>
 *        <span class="pl-k">return</span>
 *        <#textnode>&amp;Router{namedRoutes: </#textnode>
 *        <span class="pl-c1">make</span>
 *        <#textnode>(</#textnode>
 *        <span class="pl-k">map</span>
 *        <#textnode>[</#textnode>
 *        <span class="pl-k">string</span>
 *        <#textnode>]*Route), KeepContext: </#textnode>
 *        <span class="pl-c1">false</span>
 *        <#textnode>}</#textnode>
 *     </td>
 *
 * The browser extension works by registering a hover event listeners on the <td> element. When the user hovers over
 * "return" (in the first <span> node) the event target will be the <span> node. We can use the event target to determine which line
 * and which character offset on that line to use to fetch tooltip data. But when the user hovers over "Router"
 * (in the second text node) the event target will be the <td> node, which lacks the appropriate specificity to request
 * tooltip data. To circumvent this, all we need to do is wrap every free text node in a <span> tag.
 *
 * In summary, convertNode effectively does this: https://gist.github.com/lebbe/6464236
 *
 * There are three additional edge cases we handle:
 *   1. some text nodes contain multiple discrete code tokens, like the second text node in the example above; by wrapping
 *     that text node in a <span> we lose the ability to distinguish whether the user is hovering over "Router" or "namedRoutes".
 *   2. there may be arbitrary levels of <span> nesting; in the example above, every <span> node has only one (text node) child, but
 *     in reality a <span> node could have multiple children, both text and element nodes
 *   3. on GitHub diff views (e.g. pull requests) the table cell contains an additional prefix character ("+" or "-" or " ", representing
 *     additions, deletions, and unchanged code, respectively); we want to make sure we don't count that character when computing the
 *     character offset for the line
 *   4. TODO(john) some code hosts transform source code before rendering; in the example above, the first text node may be a tab character
 *     or multiple spaces
 *
 * @param parentNode The node to convert.
 */
export function convertNode(parentNode: HTMLElement): void {
    for (let i = 0; i < parentNode.childNodes.length; ++i) {
        const node = parentNode.childNodes[i]
        const isLastNode = i === parentNode.childNodes.length - 1
        if (node.nodeType === Node.TEXT_NODE) {
            let nodeText = _.unescape(node.textContent || '')
            if (nodeText === '') {
                continue
            }
            parentNode.removeChild(node)
            let insertBefore = i

            while (true) {
                const nextToken = consumeNextToken(nodeText)
                if (nextToken === '') {
                    break
                }
                const newTextNode = document.createTextNode(nextToken)
                const newTextNodeWrapper = document.createElement('SPAN')
                newTextNodeWrapper.classList.add('wrapped-node')
                newTextNodeWrapper.appendChild(newTextNode)
                if (isLastNode) {
                    parentNode.appendChild(newTextNodeWrapper)
                } else {
                    // increment insertBefore as new span-wrapped text nodes are added
                    parentNode.insertBefore(newTextNodeWrapper, parentNode.childNodes[insertBefore++])
                }
                nodeText = nodeText.substr(nextToken.length)
            }
        } else if (node.nodeType === Node.ELEMENT_NODE) {
            const elementNode = node as HTMLElement
            if (elementNode.children.length > 0 || (elementNode.textContent && elementNode.textContent.trim().length)) {
                convertNode(elementNode)
            }
        }
    }
}

const VARIABLE_TOKENIZER = /(^\w+)/
const ASCII_CHARACTER_TOKENIZER = /(^[\x21-\x2F|\x3A-\x40|\x5B-\x60|\x7B-\x7E])/
const NONVARIABLE_TOKENIZER = /(^[^\x21-\x7E]+)/

/**
 * consumeNextToken parses the text content of a text node and returns the next "distinct"
 * code token. It handles edge case #1 from convertNode(). The tokenization scheme is
 * heuristic-based and uses simple regular expressions.
 * @param txt Aribitrary text to tokenize.
 */
function consumeNextToken(txt: string): string {
    if (txt.length === 0) {
        return ''
    }

    // first, check for real stuff, i.e. sets of [A-Za-z0-9_]
    const variableMatch = txt.match(VARIABLE_TOKENIZER)
    if (variableMatch) {
        return variableMatch[0]
    }
    // next, check for tokens that are not variables, but should stand alone
    // i.e. {}, (), :;. ...
    const asciiMatch = txt.match(ASCII_CHARACTER_TOKENIZER)
    if (asciiMatch) {
        return asciiMatch[0]
    }
    // finally, the remaining tokens we can combine into blocks, since they are whitespace
    // or UTF8 control characters. We had better clump these in case UTF8 control bytes
    // require adjacent bytes
    const nonVariableMatch = txt.match(NONVARIABLE_TOKENIZER)
    if (nonVariableMatch) {
        return nonVariableMatch[0]
    }
    return txt[0]
}

function getPreDataContainer(target: HTMLElement): HTMLPreElement | undefined {
    if (target.tagName === 'PRE') {
        return target as HTMLPreElement
    }
    while (target && target.tagName !== 'PRE' && target.tagName !== 'BODY') {
        // Find ancestor which wraps the whole line of code, not just the target token.
        target = target.parentNode as HTMLElement
    }
    if (!target) {
        return undefined
    }
    if (target.tagName === 'PRE') {
        return target as HTMLPreElement
    }
}

export function getTableDataCell(target: HTMLElement): HTMLTableDataCellElement | undefined {
    if (target.tagName === 'TD') {
        return target as HTMLTableDataCellElement
    }
    while (target && target.tagName !== 'TD' && target.tagName !== 'BODY') {
        // Find ancestor which wraps the whole line of code, not just the target token.
        target = target.parentNode as HTMLElement
    }
    if (!target) {
        return undefined
    }
    if (target.tagName === 'TD') {
        return target as HTMLTableDataCellElement
    }
}

function isInsideCodeContainer(target: HTMLElement): boolean {
    return Boolean(getPreDataContainer(target) || getTableDataCell(target))
}
