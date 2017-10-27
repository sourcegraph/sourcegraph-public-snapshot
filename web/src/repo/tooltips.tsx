import { highlightBlock, registerLanguage } from 'highlight.js/lib/highlight'
import unescape from 'lodash/unescape'
import marked from 'marked'
import { Hover, MarkedString } from 'vscode-languageserver-types'
import { getModeFromExtension } from '../util'
import { toAbsoluteBlobURL } from '../util/url'
import { AbsoluteRepoFilePosition, parseBrowserRepoURL } from './index'

registerLanguage('go', require('highlight.js/lib/languages/go'))
registerLanguage('javascript', require('highlight.js/lib/languages/javascript'))
registerLanguage('typescript', require('highlight.js/lib/languages/typescript'))
registerLanguage('java', require('highlight.js/lib/languages/java'))
registerLanguage('python', require('highlight.js/lib/languages/python'))
registerLanguage('php', require('highlight.js/lib/languages/php'))

let tooltip: HTMLElement
let loadingTooltip: HTMLElement
let tooltipActions: HTMLElement
let j2dAction: HTMLAnchorElement
let findRefsAction: HTMLAnchorElement
let moreContext: HTMLElement

// tslint:disable-next-line:max-line-length
const closeIconSVG = '<svg width="10px" height="10px"><path xmlns="http://www.w3.org/2000/svg" id="path0_fill" d="M 7.8565 7.86521C 7.66117 8.06054 7.3445 8.06054 7.14917 7.86521L 3.99917 4.71521L 0.851833 7.86254C 0.655167 8.05721 0.3385 8.05521 0.1445 7.85854C -0.0481667 7.66388 -0.0481667 7.34988 0.1445 7.15454L 3.29183 4.00721L 0.145167 0.860543C -0.0475001 0.663209 -0.0428332 0.346543 0.155167 0.153876C 0.349167 -0.0347905 0.6585 -0.0347905 0.8525 0.153876L 3.99917 3.30054L 7.1485 0.151209C 7.34117 -0.0467907 7.65783 -0.0507906 7.85583 0.141876C 8.05383 0.334543 8.05783 0.651209 7.86517 0.849209C 7.86183 0.852543 7.85917 0.855209 7.85583 0.858543L 4.7065 4.00788L 7.8565 7.15788C 8.05183 7.35321 8.0525 7.66988 7.8565 7.86521Z" /></svg >'
// tslint:disable-next-line:max-line-length
const referencesIconSVG = '<svg width="12px" height="8px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path15_fill" d="M 6.00625 8C 2.33125 8 0.50625 5.075 0.05625 4.225C -0.01875 4.075 -0.01875 3.9 0.05625 3.775C 0.50625 2.925 2.33125 0 6.00625 0C 9.68125 0 11.5063 2.925 11.9563 3.775C 12.0312 3.925 12.0312 4.1 11.9563 4.225C 11.5063 5.075 9.68125 8 6.00625 8ZM 6.00625 1.25C 4.48125 1.25 3.25625 2.475 3.25625 4C 3.25625 5.525 4.48125 6.75 6.00625 6.75C 7.53125 6.75 8.75625 5.525 8.75625 4C 8.75625 2.475 7.53125 1.25 6.00625 1.25ZM 6.00625 5.75C 5.03125 5.75 4.25625 4.975 4.25625 4C 4.25625 3.025 5.03125 2.25 6.00625 2.25C 6.98125 2.25 7.75625 3.025 7.75625 4C 7.75625 4.975 6.98125 5.75 6.00625 5.75Z"/></svg>'
const definitionIconSVG = '<svg width="11px" height="9px"><path fill="#FFFFFF" xmlns="http://www.w3.org/2000/svg" id="path10_fill" d="M 6.325 8.4C 6.125 8.575 5.8 8.55 5.625 8.325C 5.55 8.25 5.5 8.125 5.5 8L 5.5 6C 2.95 6 1.4 6.875 0.825 8.7C 0.775 8.875 0.6 9 0.425 9C 0.2 9 -4.44089e-16 8.8 -4.44089e-16 8.575C -4.44089e-16 8.575 -4.44089e-16 8.575 -4.44089e-16 8.55C 0.125 4.825 1.925 2.675 5.5 2.5L 5.5 0.5C 5.5 0.225 5.725 8.88178e-16 6 8.88178e-16C 6.125 8.88178e-16 6.225 0.05 6.325 0.125L 10.825 3.875C 11.025 4.05 11.075 4.375 10.9 4.575C 10.875 4.6 10.85 4.625 10.825 4.65L 6.325 8.4Z"/></svg>'

export interface TooltipData extends Partial<Hover> {
    target: HTMLElement
    ctx: AbsoluteRepoFilePosition
    defUrl?: string
    loading?: boolean
}

/**
 * createTooltips initializes the DOM elements used for the hover
 * tooltip and "Loading..." text indicator, adding the former
 * to the DOM (but hidden). It is idempotent.
 */
export function createTooltips(): void {
    if (document.querySelector('.sg-tooltip')) {
        return // idempotence
    }

    tooltip = document.createElement('DIV')
    tooltip.className = 'tooltip'
    tooltip.classList.add('sg-tooltip')
    tooltip.style.visibility = 'hidden'

    document.querySelector('.blob')!.appendChild(tooltip)

    loadingTooltip = document.createElement('DIV')
    loadingTooltip.appendChild(document.createTextNode('Loading...'))
    loadingTooltip.className = 'tooltip__loading'

    tooltipActions = document.createElement('DIV')
    tooltipActions.className = 'tooltip__actions'

    moreContext = document.createElement('DIV')
    moreContext.className = 'tooltip__more-actions'
    moreContext.appendChild(document.createTextNode('Click for more actions'))

    const definitionIcon = document.createElement('svg')
    definitionIcon.innerHTML = definitionIconSVG
    definitionIcon.className = 'tooltip__definition-icon'

    j2dAction = document.createElement('A') as HTMLAnchorElement
    j2dAction.appendChild(definitionIcon)
    j2dAction.appendChild(document.createTextNode('Go to definition'))
    j2dAction.className = `btn btn-sm BtnGroup-item`
    j2dAction.className = 'tooltip__action'

    const referencesIcon = document.createElement('svg')
    referencesIcon.innerHTML = referencesIconSVG
    referencesIcon.className = 'tooltip__references-icon'

    findRefsAction = document.createElement('A') as HTMLAnchorElement
    findRefsAction.appendChild(referencesIcon)
    findRefsAction.appendChild(document.createTextNode('Find references'))
    findRefsAction.className = 'tooltip__action'

    tooltipActions.appendChild(j2dAction)
    tooltipActions.appendChild(findRefsAction)
}

function constructBaseTooltip(): void {
    tooltip.appendChild(loadingTooltip)
    tooltip.appendChild(moreContext)
    tooltip.appendChild(tooltipActions)
}

/**
 * hideTooltip makes the tooltip on the DOM invisible.
 */
export function hideTooltip(): void {
    if (!tooltip) {
        return
    }

    while (tooltip.firstChild) {
        tooltip.removeChild(tooltip.firstChild)
    }
    tooltip.style.visibility = 'hidden' // prevent black dot of empty content
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
export function updateTooltip(data: TooltipData, docked: boolean, actions: Actions): void {
    hideTooltip() // hide before updating tooltip text

    const { loading, target, ctx } = data

    if (!target) {
        // no target to show hover for; tooltip is hidden
        return
    }

    constructBaseTooltip()
    loadingTooltip.style.display = loading ? 'block' : 'none'
    moreContext.style.display = docked || loading ? 'none' : 'flex'
    tooltipActions.style.display = docked ? 'flex' : 'none'

    j2dAction.style.display = 'block'
    j2dAction.href = data.defUrl ? data.defUrl : ''

    if (data.defUrl && j2dAction.href !== window.location.href) {
        j2dAction.style.cursor = 'pointer'
        j2dAction.onclick = actions.definition(parseBrowserRepoURL(data.defUrl) as AbsoluteRepoFilePosition)
    } else {
        j2dAction.style.cursor = 'not-allowed'
        j2dAction.onclick = () => false
    }

    findRefsAction.style.display = 'block'
    findRefsAction.onclick = actions.references(ctx)

    if (ctx) {
        let revString = ''
        if (ctx.commitID) {
            revString = `@${ctx.commitID}`
        }
        findRefsAction.href = toAbsoluteBlobURL({ ...ctx, referencesMode: 'local' })
    } else {
        findRefsAction.href = ''
    }

    if (!loading) {
        loadingTooltip.style.visibility = 'hidden'

        if (!data.contents) {
            return
        }
        const contentsArray: MarkedString[] = Array.isArray(data.contents) ? data.contents : [data.contents]
        if (contentsArray.length === 0) {
            return
        }
        const firstContent = contentsArray[0]
        const title: string = typeof firstContent === 'string' ? firstContent : firstContent.value
        let doc: string | undefined
        for (const markedString of contentsArray.slice(1)) {
            if (typeof markedString === 'string') {
                doc = markedString
            } else if (markedString.language === 'markdown') {
                doc = markedString.value
            }
        }
        if (!title) {
            // no tooltip text / search context; tooltip is hidden
            return
        }

        const container = document.createElement('DIV')
        container.className = 'tooltip__divider'

        const tooltipText = document.createElement('DIV')
        tooltipText.className = `tooltip__title ${getModeFromExtension(ctx.filePath)}`
        tooltipText.appendChild(document.createTextNode(title))

        container.appendChild(tooltipText)
        tooltip.insertBefore(container, moreContext)

        const closeContainer = document.createElement('a')
        closeContainer.className = 'tooltip__close-icon'
        closeContainer.onclick = actions.dismiss

        if (docked) {
            const closeButton = document.createElement('svg')
            closeButton.innerHTML = closeIconSVG
            closeContainer.appendChild(closeButton)
            container.appendChild(closeContainer)
        }

        highlightBlock(tooltipText)

        if (doc) {
            const tooltipDoc = document.createElement('DIV')
            tooltipDoc.className = 'tooltip__doc'
            tooltipDoc.innerHTML = marked(doc, { gfm: true, breaks: true, sanitize: true })
            tooltip.insertBefore(tooltipDoc, moreContext)

            // Handle scrolling ourselves so that scrolling to the bottom of
            // the tooltip documentation does not cause the page to start
            // scrolling (which is a very jarring experience).
            tooltip.addEventListener('wheel', (e: WheelEvent) => {
                e.preventDefault()
                tooltipDoc.scrollTop += e.deltaY
            })
        }
    } else {
        loadingTooltip.style.visibility = 'visible'
    }

    const scrollingElement = document.querySelector('.blob')!
    const scrollingElementBound = scrollingElement.getBoundingClientRect()
    const blobTable = document.querySelector('.blob > table')! // table that we're positioning tooltips relative to.
    const tableBound = blobTable.getBoundingClientRect() // tables bounds
    const targetBound = target.getBoundingClientRect() // our target elements bounds

    // Anchor it horizontally, prior to rendering to account for wrapping
    // changes to vertical height if the tooltip is at the edge of the viewport.
    const relLeft = targetBound.left - tableBound.left
    tooltip.style.left = relLeft + 'px'

    // Anchor the tooltip vertically.
    const tooltipBound = tooltip.getBoundingClientRect()
    const relTop = (targetBound.top + scrollingElement.scrollTop) - scrollingElementBound.top
    const margin = 5
    let tooltipTop = relTop - (tooltipBound.height + margin)
    if ((tooltipTop - scrollingElement.scrollTop) < 0) {
        // Tooltip wouldn't be visible from the top, so display it at the
        // bottom.
        const relBottom = (targetBound.bottom + scrollingElement.scrollTop) - scrollingElementBound.top
        tooltipTop = relBottom + margin
    }
    tooltip.style.top = tooltipTop + 'px'

    // Make it all visible to the user.
    tooltip.style.visibility = 'visible'
}

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
            let nodeText = unescape(node.textContent || '')
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
            // if (elementNode.children.length > 0) {
            // The element is something more complicated than <span>text</span>; recurse.
            convertNode(elementNode)
            // }
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

export function getTableDataCell(target: HTMLElement): HTMLTableDataCellElement | undefined {
    if (target.tagName === 'TD') {
        // Short-circuit; we are hovering over a line of code, but no token in particular.
        return undefined
    }
    while (target && target.tagName !== 'TD' && target.tagName !== 'BODY') {
        // Find ancestor which wraps the whole line of code, not just the target token.
        target = (target.parentNode as HTMLElement)
    }
    if (target.tagName === 'TD') {
        return target as HTMLTableDataCellElement
    }
    return undefined
}

/**
 * Returns the <span> (descendent of a <td> containing code) which contains text beginning
 * at the specified character offset (1-indexed)
 * @param cell the <td> containing syntax highlighted code
 * @param offset character offset
 */
export function findElementWithOffset(cell: HTMLElement, offset: number): HTMLElement | undefined {
    let currOffset = 0
    const walkNode = (currNode: HTMLElement): HTMLElement | undefined => {
        const numChildNodes = currNode.childNodes.length
        for (let i = 0; i < numChildNodes; ++i) {
            const child = currNode.childNodes[i]
            switch (child.nodeType) {
                case Node.TEXT_NODE:
                    if (currOffset + child.textContent!.length >= offset) {
                        return currNode
                    }
                    currOffset += child.textContent!.length
                    continue

                case Node.ELEMENT_NODE:
                    const found = walkNode(child as HTMLElement)
                    if (found) {
                        return found
                    }
                    continue
            }
        }
        return undefined
    }
    return walkNode(cell)
}

/**
 * getTargetLineAndOffset determines the line and character offset for some source code, identified by its HTMLElement wrapper.
 * It works by traversing the DOM until the HTMLElement's TD ancestor. Once the ancestor is found, we traverse the DOM again
 * (this time the opposite direction) counting characters until the original target is found.
 * Returns undefined if line/char cannot be determined for the provided target.
 * @param target The element to compute line & character offset for.
 * @param ignoreFirstChar Whether to ignore the first character on a line when computing character offset.
 */
export function getTargetLineAndOffset(target: HTMLElement, ignoreFirstChar = false): { line: number, character: number, word: string } | undefined {
    const origTarget = target
    if (target.tagName === 'TD') {
        // Short-circuit; we are hovering over a line of code, but no token in particular.
        return undefined
    }
    while (target && target.tagName !== 'TD' && target.tagName !== 'BODY') {
        // Find ancestor which wraps the whole line of code, not just the target token.
        target = (target.parentNode as HTMLElement)
    }
    if (!target || target.tagName !== 'TD') {
        // Make sure we're looking at an element we've annotated line number for (otherwise we have no idea )
        return undefined
    }
    let line: number
    if (target.classList.contains('line')) {
        line = parseInt(target.getAttribute('data-line')!, 10)
    } else {
        line = parseInt((target.previousElementSibling as HTMLTableDataCellElement).getAttribute('data-line')!, 10)
    }

    let character = 1
    // Iterate recursively over the current target's children until we find the original target;
    // count characters along the way. Return true if the original target is found.
    function findOrigTarget(root: HTMLElement): boolean {
        // tslint:disable-next-line
        for (let i = 0; i < root.childNodes.length; ++i) {
            const child = root.childNodes[i] as HTMLElement
            if (child === origTarget) {
                return true
            }
            if (child.children === undefined) {
                character += child.textContent!.length
                continue
            }
            if (child.children.length > 0 && findOrigTarget(child)) {
                // Walk over nested children, then short-circuit the loop to avoid double counting children.
                return true
            }
            if (child.children.length === 0) {
                // Child is not the original target, but has no chidren to recurse on. Add to character offset.
                character += (child.textContent as string).length // TODO(john): I think this needs to be escaped before we add its length...
                if (ignoreFirstChar) {
                    character -= 1 // make sure not to count weird diff prefix
                    ignoreFirstChar = false
                }
            }
        }
        return false
    }
    // Start recursion.
    if (findOrigTarget(target)) {
        return { line, character, word: origTarget.innerText }
    }
    return undefined
}
