/**
 * Inserts an element after the reference node.
 * @param el The element to be rendered.
 * @param referenceNode The node to render the element after.
 */
export function insertAfter(el: HTMLElement, referenceNode: Node): void {
    if (referenceNode.parentNode) {
        referenceNode.parentNode.insertBefore(el, referenceNode.nextSibling)
    }
}

export function isMouseEventWithModifierKey(e: MouseEvent): boolean {
    return e.altKey || e.shiftKey || e.ctrlKey || e.metaKey || e.which === 2
}

/**
 * Highlights a node using recursive node walking.
 *
 * @param node the node to highlight
 * @param start the current character position (starts at 0).
 * @param lenght the number of characters to highlight.
 */
export function highlightNode(node: HTMLElement, start: number, length: number, w: Window = window): void {
    if (start < 0 || length <= 0 || start >= node.textContent!.length) {
        return
    }
    if (node.classList.contains('annotated-selection-match')) {
        return
    }
    node.classList.add('annotated-selection-match')
    highlightNodeHelper(node, 0, start, length, w)
}

interface HighlightResult {
    highlightingCompleted: boolean
    charsConsumed: number
    charsHighlighted: number
}

/**
 * Highlights a node using recursive node walking.
 *
 * @param currNode the current node being walked.
 * @param currOffset the current character position (starts at 0).
 * @param start the offset character where highlting starts.
 * @param lenght the number of characters to highlight.
 */
function highlightNodeHelper(
    currNode: HTMLElement,
    currOffset: number,
    start: number,
    length: number,
    w: Window
): HighlightResult {
    // typescript doesn't define Node on the Window, although it exists (and must be used e.g. when we are passed a Window from jsdom)
    const Node = (w as any).Node as Node

    if (length === 0) {
        return { highlightingCompleted: true, charsConsumed: 0, charsHighlighted: 0 }
    }

    const origOffset = currOffset
    const numChildNodes = currNode.childNodes.length

    let charsHighlighted = 0

    for (let i = 0; i < numChildNodes; ++i) {
        if (currOffset >= start + length) {
            return { highlightingCompleted: true, charsConsumed: 0, charsHighlighted: 0 }
        }
        const isLastNode = i === currNode.childNodes.length - 1
        const child = currNode.childNodes[i]

        switch (child.nodeType) {
            case Node.TEXT_NODE: {
                const nodeText = child.textContent!

                if (currOffset <= start && currOffset + nodeText.length > start) {
                    // Current node overlaps start of highlighting.
                    currNode.removeChild(child)

                    // The characters beginning at the start of highlighting and extending to the end of the node.
                    const rest = nodeText.substr(start - currOffset)

                    const containerNode = w.document.createElement('span')
                    if (nodeText.substr(0, start - currOffset) !== '') {
                        // If characters were consumed leading up to the start of highlighting, add them to the parent.
                        containerNode.appendChild(w.document.createTextNode(nodeText.substr(0, start - currOffset)))
                    }

                    if (rest.length >= length) {
                        // The highligted range is fully contained within the node.
                        const text = rest.substr(0, length)
                        const highlight = w.document.createElement('span')
                        highlight.className = 'selection-highlight'
                        highlight.appendChild(w.document.createTextNode(text))
                        containerNode.appendChild(highlight)
                        if (rest.length > length) {
                            // There is more in the span than the highlighted chars.
                            containerNode.appendChild(w.document.createTextNode(rest.substr(length)))
                        }

                        if (currNode.childNodes.length === 0 || isLastNode) {
                            currNode.appendChild(containerNode)
                        } else {
                            currNode.insertBefore(containerNode, currNode.childNodes[i] || currNode.firstChild)
                        }

                        return { highlightingCompleted: true, charsConsumed: nodeText.length, charsHighlighted: length }
                    }

                    // Else the highlighted range spans multiple nodes.
                    charsHighlighted += rest.length

                    const highlight = w.document.createElement('span')
                    highlight.className = 'selection-highlight'
                    highlight.appendChild(w.document.createTextNode(rest))
                    containerNode.appendChild(highlight)

                    if (currNode.childNodes.length === 0 || isLastNode) {
                        currNode.appendChild(containerNode)
                    } else {
                        currNode.insertBefore(containerNode, currNode.childNodes[i] || currNode.firstChild)
                    }
                }

                currOffset += nodeText.length
                break
            }

            case Node.ELEMENT_NODE: {
                const elementNode = child as HTMLElement
                const res = highlightNodeHelper(
                    elementNode,
                    currOffset,
                    start + charsHighlighted,
                    length - charsHighlighted,
                    w
                )
                if (res.highlightingCompleted) {
                    return res
                } else {
                    currOffset += res.charsConsumed
                    charsHighlighted += res.charsHighlighted
                }
                break
            }
        }
    }

    return { highlightingCompleted: false, charsConsumed: currOffset - origOffset, charsHighlighted }
}
