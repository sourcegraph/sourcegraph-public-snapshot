import * as _ from 'lodash'

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

export function highlightNode(parentNode: HTMLElement, start: number, end: number): void {
    if (parentNode.classList.contains('annotated-selection-match')) {
        return
    }
    parentNode.classList.add('annotated-selection-match')
    highlightNodeHelper(parentNode, 0, start, end)
}

interface HighlightIteration {
    done: boolean
    consumed: number
    highlighted: number
}

function highlightNodeHelper(parentNode: HTMLElement, curr: number, start: number, length: number, currContainerNode?: HTMLElement): HighlightIteration {
    const origCurr = curr
    const numParentNodes = parentNode.childNodes.length

    if (length === 0) {
        return { done: true, consumed: 0, highlighted: 0}
    }

    let highlighted = 0

    for (let i = 0; i < numParentNodes; ++i) {
        if (curr >= start + length) {
            return { done: true, consumed: 0, highlighted: 0 }
        }
        const isLastNode = i === parentNode.childNodes.length - 1
        const node = parentNode.childNodes[i]
        if (node.nodeType === Node.TEXT_NODE) {
            const nodeText = _.unescape(node.textContent || '')

            if (curr <= start && curr + nodeText.length > start) {
                // Current node overlaps start of highlighting.
                parentNode.removeChild(node)

                // The characters beginning at the start of highlighting and extending to the end of the node.
                const rest = nodeText.substr(start - curr)

                const containerNode = document.createElement('span')
                if (nodeText.substr(0, start - curr) !== '') {
                    // If characters were consumed leading up to the start of highlighting, add them to the parent.
                    containerNode.appendChild(document.createTextNode(nodeText.substr(0, start - curr)))
                }

                if (rest.length >= length) {
                    // The highligted range is fully contained within the node.
                    const text = rest.substr(0, length)
                    const highlight = document.createElement('span')
                    highlight.className = 'selection-highlight'
                    highlight.appendChild(document.createTextNode(text))
                    containerNode.appendChild(highlight)
                    if (rest.substr(length)) {
                        containerNode.appendChild(document.createTextNode(rest.substr(length)))
                    }

                    if (parentNode.childNodes.length === 0 || isLastNode) {
                        parentNode.appendChild(containerNode)
                    } else {
                        parentNode.insertBefore(containerNode, parentNode.childNodes[i] || parentNode.firstChild)
                    }

                    return { done: true, consumed: nodeText.length, highlighted: length }
                } else {
                    // The highlighted range spans multiple nodes.
                    highlighted += rest.length

                    const highlight = document.createElement('span')
                    highlight.className = 'selection-highlight'
                    highlight.appendChild(document.createTextNode(rest))
                    containerNode.appendChild(highlight)

                    if (parentNode.childNodes.length === 0 || isLastNode) {
                        parentNode.appendChild(containerNode)
                    } else {
                        parentNode.insertBefore(containerNode, parentNode.childNodes[i] || parentNode.firstChild)
                    }
                }
            }

            curr += nodeText.length
        } else if (node.nodeType === Node.ELEMENT_NODE) {
            const elementNode = node as HTMLElement
            if (elementNode.classList.contains('selection-highlight')) {
                return { done: true, consumed: 0, highlighted: 0 }
            }
            const res = highlightNodeHelper(elementNode, curr, start + highlighted, length - highlighted)
            if (res.done) {
                return res
            } else {
                curr += res.consumed
                highlighted += res.highlighted
            }
        }
    }
    return { done: false, consumed: curr - origCurr, highlighted }
}
