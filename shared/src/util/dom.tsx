/**
 * Highlights a node using recursive node walking.
 *
 * @param node the node to highlight
 * @param start the current character position (starts at 0).
 * @param length the number of characters to highlight.
 */
export function highlightNode(node: HTMLElement, start: number, length: number): void {
    if (start < 0 || length <= 0 || start >= node.textContent!.length) {
        return
    }

    // return if length is invalid/longer than remaining characters between start and end
    if (length > node.textContent!.length - start) {
        return
    }

    // TODO!(sqs): prevent double-highlighting
    if (node.querySelector('.selection-highlight')) {
        return
    }

    // We want to treat text nodes as walkable so they can be highlighted. Wrap these in a span and
    // replace them in the DOM.
    if (node.nodeType === Node.TEXT_NODE && node.textContent !== null) {
        const sp = document.createElement('span')
        sp.innerHTML = node.textContent
        node.parentNode!.replaceChild(sp, node)
        node = sp
    }
    node.classList.add('annotated-selection-match')
    highlightNodeHelper(node, 0, start, length)
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
 * @param length the number of characters to highlight.
 */
function highlightNodeHelper(
    currNode: HTMLElement,
    currOffset: number,
    start: number,
    length: number
): HighlightResult {
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

                    const containerNode = document.createElement('span')
                    if (nodeText.substr(0, start - currOffset) !== '') {
                        // If characters were consumed leading up to the start of highlighting, add them to the parent.
                        containerNode.appendChild(document.createTextNode(nodeText.substr(0, start - currOffset)))
                    }

                    if (rest.length >= length) {
                        // The highlighted range is fully contained within the node.
                        if (currNode.classList.contains('selection-highlight')) {
                            // Nothing to do; it's already highlighted.
                            currNode.appendChild(child)
                        } else {
                            const text = rest.substr(0, length)
                            const highlight = document.createElement('span')
                            highlight.className = 'selection-highlight'
                            highlight.appendChild(document.createTextNode(text))
                            containerNode.appendChild(highlight)
                            if (rest.length > length) {
                                // There is more in the span than the highlighted chars.
                                containerNode.appendChild(document.createTextNode(rest.substr(length)))
                            }

                            if (currNode.childNodes.length === 0 || isLastNode) {
                                currNode.appendChild(containerNode)
                            } else {
                                currNode.insertBefore(containerNode, currNode.childNodes[i] || currNode.firstChild)
                            }
                        }

                        return { highlightingCompleted: true, charsConsumed: nodeText.length, charsHighlighted: length }
                    }

                    // Else the highlighted range spans multiple nodes.
                    charsHighlighted += rest.length

                    const highlight = document.createElement('span')
                    highlight.className = 'selection-highlight'
                    highlight.appendChild(document.createTextNode(rest))
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
                    length - charsHighlighted
                )
                if (res.highlightingCompleted) {
                    return res
                }
                currOffset += res.charsConsumed
                charsHighlighted += res.charsHighlighted
                break
            }
        }
    }

    return { highlightingCompleted: false, charsConsumed: currOffset - origOffset, charsHighlighted }
}
