import { isEqual } from 'lodash'

import type { Position, Range } from '@sourcegraph/extension-api-types'

import { isPosition, type LineOrPositionOrRange } from './types'

/**
 * A collection of methods needed to tell codeintellify how to look at the DOM. These are required for
 * ensuring that we don't rely on any sort of specific DOM structure.
 */
export interface DOMFunctions {
    /**
     * Get the element containing the code for a line from an event target.
     *
     * @param target is the event target.
     * @returns the element containing the code for a line or null if it can't be found. For example, the second <td> inside a <tr> on Sourcegraph and Github.
     */
    getCodeElementFromTarget: (target: HTMLElement) => HTMLElement | null

    /**
     * Get the element containing the code for a line from a code view given a line number.
     *
     * @param codeView is the code view itself. For example, the <code> element on Sourcegraph or a <table> on GitHub.
     * @param part If the code view is a diff view, the part of the diff that the line number refers to.
     * @returns the element containing the code for the given line number or null if it can't be found.
     */
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number, part?: DiffPart) => HTMLElement | null

    /**
     * Gets the line number for a given element containing code for a line.
     * When this function is called, it will be passed the result of either `getCodeElementFromTarget` or `getCodeElementFromLineNumber`.
     *
     * @param codeElement The element containing code for a line.
     * @returns The line number.
     */
    getLineNumberFromCodeElement: (codeElement: HTMLElement) => number

    /**
     * If the code view is a diff view, must be provided to determine whether
     * a code element is from the base, head or unchanged part of the diff.
     * Must be `undefined` if the code view is not a diff view.
     *
     * @param codeElement is the element containing a line of code.
     * @returns The part of the diff `codeElement` belongs to
     */
    getDiffCodePart?: (codeElement: HTMLElement) => DiffPart

    /**
     * Must return `true` if the first character in the code element is always
     * the diff indicator (`+`, `-` or space), `false` otherwise.
     *
     * @param codeElement is the element containing a line of code.
     */
    isFirstCharacterDiffIndicator?(codeElement: HTMLElement): boolean
}

/**
 * Like `convertNode`, but idempotent.
 * The CSS class `annotated` is used to check if the cell is already converted.
 *
 * @param element The code `<td>` to convert.
 */
export function convertCodeElementIdempotent(element: HTMLElement): void {
    if (element && !element.classList.contains('annotated')) {
        convertNode(element)
        element.classList.add('annotated')
    }
}

/**
 * Helper function to determine whether the code view should be tokenized.
 */
export function shouldTokenize({
    tokenize,
    overrideTokenize,
}: {
    tokenize: boolean
    overrideTokenize: boolean | undefined
}): boolean {
    return typeof overrideTokenize === 'boolean' ? overrideTokenize : tokenize
}

/**
 * convertNode modifies a DOM node so that we can identify precisely token a user has clicked or hovered over.
 * On a code view, source code is typically wrapped in a HTML table cell. It may look like this:
 *
 * ```html
 * <td id="LC18" class="blob-code blob-code-inner js-file-line">
 *    <#textnode>\t</#textnode>
 *    <span class="pl-k">return</span>
 *    <#textnode>&amp;Router{namedRoutes: </#textnode>
 *    <span class="pl-c1">make</span>
 *    <#textnode>(</#textnode>
 *    <span class="pl-k">map</span>
 *    <#textnode>[</#textnode>
 *    <span class="pl-k">string</span>
 *    <#textnode>]*Route), KeepContext: </#textnode>
 *    <span class="pl-c1">false</span>
 *    <#textnode>}</#textnode>
 * </td>
 * ```
 *
 * The browser extension works by registering a hover event listeners on the <td> element. When the user hovers over
 * "return" (in the first <span> node) the event target will be the <span> node. We can use the event target to determine which line
 * and which character offset on that line to use to get tooltip data. But when the user hovers over "Router"
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
    for (let index = 0; index < parentNode.childNodes.length; ++index) {
        const node = parentNode.childNodes[index]
        const isLastNode = index === parentNode.childNodes.length - 1

        if (node.nodeType === Node.TEXT_NODE) {
            let nodeText = node.textContent ?? ''
            if (nodeText === '') {
                continue
            }
            node.remove()
            let insertBefore = index

            while (true) {
                const nextToken = consumeNextToken(nodeText)
                if (nextToken === '') {
                    break
                }
                const newTextNode = document.createTextNode(nextToken)
                const newTextNodeWrapper = document.createElement('SPAN')
                newTextNodeWrapper.append(newTextNode)
                if (isLastNode) {
                    parentNode.append(newTextNodeWrapper)
                } else {
                    // increment insertBefore as new span-wrapped text nodes are added
                    parentNode.insertBefore(newTextNodeWrapper, parentNode.childNodes[insertBefore++])
                }
                nodeText = nodeText.slice(nextToken.length)
            }
        } else if (node.nodeType === Node.ELEMENT_NODE) {
            const elementNode = node as HTMLElement
            if (elementNode.children.length > 0 || elementNode.textContent?.trim().length) {
                convertNode(elementNode)
            }
        }
    }
}

const VARIABLE_TOKENIZER = /(^\w+)/
const ASCII_CHARACTER_TOKENIZER = /(^[\u0021-\u002F\u003A-\u0040\u005B-\u0060\u007B-\u007E])/
const NONVARIABLE_TOKENIZER = /(^[^\u0021-\u007E]+)/

enum TokenType {
    /** Tokens that are alphanumeric, i.e. variable names, keywords */
    Alphanumeric,
    /** Tokens that are ascii characters but aren't in identies (i.e. {, }, [, ], |, ;,  etc) */
    ASCII,
    /** Every token we encounter that doesn't fall into the other two TokenTypes */
    Other,
}

/**
 * Get the type of token we are looking at.
 *
 * @param node The node containing the token.
 */
function getTokenType(node: Node): TokenType {
    const text = node.textContent ?? ''
    if (text.length === 0) {
        return TokenType.Other
    }
    const variableMatch = text.match(VARIABLE_TOKENIZER)
    if (variableMatch) {
        return TokenType.Alphanumeric
    }
    const asciiMatch = text.match(ASCII_CHARACTER_TOKENIZER)
    if (asciiMatch) {
        return TokenType.ASCII
    }
    return TokenType.Other
}

/**
 * Checks to see if the TokenType of node is the same as the provided token type.
 *
 * When tokenizing the DOM, alphanumeric characters are grouped because they are identities.
 *
 * We also group whitespace just in case. See `consumeNextToken` comments for more information.
 * This is a helper function for making sure the node is the same type of a token and if we care
 * about grouping the type of token together.
 */
function isSameTokenType(tokenType: TokenType, node: Node): boolean {
    // We don't care about grouping things like :=, ===, etc
    if (tokenType === TokenType.ASCII) {
        return false
    }

    return tokenType === getTokenType(node)
}

/**
 * consumeNextToken parses the text content of a text node and returns the next "distinct"
 * code token. It handles edge case #1 from convertNode(). The tokenization scheme is
 * heuristic-based and uses simple regular expressions.
 *
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

/**
 * Get the all of the text nodes under a given node in the DOM tree.
 *
 * @param node is the node in which you want to get all of the text nodes from it's children
 */
export const getTextNodes = (node: Node): Node[] => {
    if (node.childNodes.length === 0 && node.TEXT_NODE === node.nodeType && node.nodeValue) {
        return [node]
    }

    const nodes: Node[] = []

    for (const child of node.childNodes) {
        nodes.push(...getTextNodes(child))
    }

    return nodes
}

/**
 * Returns the <span> (descendent of a <td> containing code) which contains text beginning
 * at the specified character offset, or within the specified offsets (1-indexed).
 * Will convert tokens in the code cell if needed.
 *
 * @param codeElement the element containing syntax highlighted code
 * @param offset object that specifies start character offset, and the optional end character offset (1-indexed)
 */
export function findElementWithOffset(
    codeElement: HTMLElement,
    { offsetStart, offsetEnd }: { offsetStart: number; offsetEnd?: number },
    tokenize = true
): HTMLElement | undefined {
    if (tokenize) {
        // Without being converted first, finding the position is inaccurate
        convertCodeElementIdempotent(codeElement)
    }

    const textNodes = getTextNodes(codeElement)

    let startNode: Node
    let endNode: Node
    if (typeof offsetEnd === 'number' && offsetEnd > offsetStart) {
        // Find elements in RANGE

        /** How far forward we have looked so far. Starting at one because codeintellify treats positions as being 1-indexed. */
        let offsetStep = 1

        // Find the text nodes at the given start and end offsets
        let targetStartNode: Node | undefined
        let targetEndNode: Node | undefined
        for (const node of textNodes) {
            const text = node.textContent ?? ''

            if (!targetStartNode) {
                if (offsetStep <= offsetStart && offsetStep + text.length > offsetStart) {
                    targetStartNode = node
                }
            }
            // offsetEnd should be greater than offsetStart, so only check for it after targetStartNode has been found
            if (targetStartNode) {
                // End position of range is exclusive
                if (offsetStep < offsetEnd && offsetStep + text.length >= offsetEnd) {
                    targetEndNode = node
                    break
                }
            }

            offsetStep += text.length
        }

        if (!targetStartNode || !targetEndNode) {
            return undefined
        }
        // Since the whole range has been specified, don't look for token edges (trust the input)

        startNode = targetStartNode
        endNode = targetEndNode
    } else {
        // Find elements at POSITION

        /** How far forward we have looked so far. Starting at one because codeintellify treats positions as being 1-indexed. */
        let offsetStep = 1
        let nodeIndex = 0

        // Find the text node that is at the given offset.
        let targetNode: Node | undefined
        for (const [index, node] of textNodes.entries()) {
            const text = node.textContent ?? ''
            if (offsetStep <= offsetStart && offsetStep + text.length > offsetStart) {
                targetNode = node
                nodeIndex = index
                break
            }

            offsetStep += text.length
        }

        if (!targetNode) {
            return undefined
        }

        const tokenType = getTokenType(targetNode)

        /**
         * Walk forwards or backwards to find the edge of the actual token, not the DOM element.
         * This is needed because tokens can span different elements. In diffs, tokens can be colored
         * differently based if just part of the token changed.
         *
         * In other words, its not unexpexted to find a token that looks like: My<span>Token</span>.
         * Without doing this, just "My" or "Token" will be highlighted depending on where you hover.
         *
         * @param idx the index to start at
         * @param delta the direction we are walking
         */
        const findTokenEdgeIndex = (index: number, delta: -1 | 1): number => {
            let current = index

            while (textNodes[current + delta] && isSameTokenType(tokenType, textNodes[current + delta])) {
                current += delta
            }

            return current
        }

        startNode = textNodes[findTokenEdgeIndex(nodeIndex, -1)]
        endNode = textNodes[findTokenEdgeIndex(nodeIndex, 1)]
    }

    // Create a range spanning from the beginning of the token to the end.
    const tokenRange = document.createRange()
    tokenRange.setStartBefore(startNode)
    tokenRange.setEndAfter(endNode)

    // If the text nodes are the same, its safe to return the common ancestor which is the container element.
    if (startNode === endNode || (tokenRange.commonAncestorContainer as HTMLElement).classList.contains('wrapped')) {
        return tokenRange.commonAncestorContainer as HTMLElement
    }

    // Otherwise, we can't guarantee that the common ancestor container doesn't contain
    // whitespace or other characters around it. To solve for this case, we'll just
    // surround the contents of the range with a new span.
    const wrapper = document.createElement('span')
    wrapper.classList.add('wrapped')

    // NOTE: We can't use tokenRange.surroundContents(wrapper) because(from https://developer.mozilla.org/en-US/docs/Web/API/Range/surroundContents):
    //
    // An exception will be thrown, however, if the Range splits a non-Text node with only one of its
    // boundary points. That is, unlike the alternative above, if there are partially selected nodes,
    // they will not be cloned and instead the operation will fail.
    wrapper.append(tokenRange.extractContents())
    tokenRange.insertNode(wrapper)

    return wrapper
}

/**
 * Whether a line belongs to the base rev of the diff (removed), the head (added) or `null` if either (not changed).
 */
export type DiffPart = 'base' | 'head' | null

/**
 * Returned when only the line is known.
 *
 * 1-indexed
 */
export interface Line {
    line: number
}

export interface HoveredToken {
    /** 1-indexed */
    line: number
    /** 1-indexed */
    character: number
    part?: DiffPart
}

/**
 * Determines the line and character offset for some source code, identified by its HTMLElement wrapper.
 * It works by traversing the DOM until the HTMLElement's TD ancestor. Once the ancestor is found, we traverse the DOM again
 * (this time the opposite direction) counting characters until the original target is found.
 * Returns undefined if line/char cannot be determined for the provided target.
 *
 * @param target The element to compute line & character offset for.
 * @param ignoreFirstChar Whether to ignore the first character on a line when computing character offset.
 */
export function locateTarget(
    target: HTMLElement,
    {
        getCodeElementFromTarget,
        getLineNumberFromCodeElement,
        getDiffCodePart,
        isFirstCharacterDiffIndicator,
    }: DOMFunctions
): Line | HoveredToken | undefined {
    const codeElement = getCodeElementFromTarget(target)

    if (!codeElement) {
        // Make sure we're looking at an element we've annotated line number for (otherwise we have no idea )
        return undefined
    }

    const line = getLineNumberFromCodeElement(codeElement)

    // If the hovered target was the code element itself or a parent,
    // make sure to not return the last character
    if (target === codeElement) {
        return { line }
    }

    const part = getDiffCodePart?.(codeElement)
    let ignoreFirstCharacter = !!isFirstCharacterDiffIndicator && isFirstCharacterDiffIndicator(codeElement)

    let character = 1
    // Iterate recursively over the current target's children until we find the original target;
    // count characters along the way. Return true if the original target is found.
    function findOrigTarget(root: HTMLElement): boolean {
        if (root === target) {
            return true
        }
        for (const child of root.childNodes) {
            if (child === target) {
                return true
            }
            if (!(child instanceof HTMLElement)) {
                character += child.textContent!.length
                continue
            }
            if (child.children.length > 0 && findOrigTarget(child)) {
                // Walk over nested children, then short-circuit the loop to avoid double counting children.
                return true
            }
            if (child.children.length === 0) {
                // Child is not the original target, but has no children to recurse on. Add to character offset.
                character += (child.textContent as string).length // TODO(john): I think this needs to be escaped before we add its length...
                if (ignoreFirstCharacter) {
                    character -= 1
                    ignoreFirstCharacter = false
                }
            }
        }
        return false
    }
    // Start recursion.
    if (findOrigTarget(codeElement)) {
        return { line, character, part }
    }
    return { line }
}

export interface GetCodeElementsInRangeOptions extends Pick<DOMFunctions, 'getCodeElementFromLineNumber'> {
    codeView: HTMLElement
    position?: LineOrPositionOrRange
    part?: DiffPart
}

export const getCodeElementsInRange = ({
    codeView,
    position,
    part,
    getCodeElementFromLineNumber,
}: GetCodeElementsInRangeOptions): {
    /** 1-indexed line number */
    line: number
    /** The element containing the code */
    element: HTMLElement
}[] => {
    if (!position || position.line === undefined) {
        return []
    }

    const elements: { line: number; element: HTMLElement }[] = []
    for (let line = position.line; line <= (position.endLine || position.line); line++) {
        const element = getCodeElementFromLineNumber(codeView, line, part)
        if (!element) {
            break
        }
        elements.push({ line, element })
    }
    return elements
}

/**
 * Returns the token `<span>` element in a code view for a given 1-indexed position or range.
 *
 * @param codeView The code view
 * @param positionOrRange 1-indexed position or range
 * @param domOptions Code-host specific implementations of DOM retrieval functions
 * @param part If the code view is a diff view, the part of the diff that the position refers to
 */
export const getTokenAtPositionOrRange = (
    codeView: HTMLElement,
    positionOrRange: Position | Range,
    {
        getCodeElementFromLineNumber,
        isFirstCharacterDiffIndicator,
    }: Pick<DOMFunctions, 'getCodeElementFromLineNumber' | 'isFirstCharacterDiffIndicator'>,
    part?: DiffPart,
    tokenize = true
): HTMLElement | undefined => {
    if (
        isPosition(positionOrRange) ||
        isEqual(positionOrRange.start, positionOrRange.end) ||
        positionOrRange.start.line !== positionOrRange.end.line
    ) {
        // In this branch, `positionOrRange` is either a position, or an invalid range that we treat as a position
        let { line, character } = isPosition(positionOrRange) ? positionOrRange : positionOrRange.start

        const codeElement = getCodeElementFromLineNumber(codeView, line, part)
        if (!codeElement) {
            return undefined
        }

        // On diff pages, account for the +/- indicator
        if (isFirstCharacterDiffIndicator?.(codeElement)) {
            character++
        }

        return findElementWithOffset(codeElement, { offsetStart: character }, tokenize)
    }

    const { start, end } = positionOrRange

    let { character: startCharacter, line } = start
    let { character: endCharacter } = end

    const codeElement = getCodeElementFromLineNumber(codeView, line, part)
    if (!codeElement) {
        return undefined
    }

    if (isFirstCharacterDiffIndicator?.(codeElement)) {
        startCharacter++
        endCharacter++
    }

    // Ensure that `startCharacter` is always smaller than `endCharacter`
    return findElementWithOffset(
        codeElement,
        {
            offsetStart: startCharacter < endCharacter ? startCharacter : endCharacter,
            offsetEnd: startCharacter < endCharacter ? endCharacter : startCharacter,
        },
        tokenize
    )
}
