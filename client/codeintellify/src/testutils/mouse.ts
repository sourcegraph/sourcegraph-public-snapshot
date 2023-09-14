import type { Position } from '@sourcegraph/extension-api-types'

import { convertNode } from '../tokenPosition'

import type { CodeViewProps } from './dom'

interface Coordinates {
    x: number
    y: number
}

export const createMouseEvent = (type: string, coords: Coordinates): MouseEvent => {
    const event = new MouseEvent(type, {
        clientX: coords.x,
        clientY: coords.y,
        bubbles: true, // Must be true so that React can see it.
    })

    return event
}

const invalidPosition = ({ line, character }: Position, message: string): string =>
    `Invalid position L${line}:${character}. ${message}. Positions are 0-indexed.`

/**
 * Dispatch a mouse event at the given position in a code element. This is impure because the current hoverifier
 * implementation requires the click event to come from the already tokenized DOM elements. Ideally we would not
 * rely on this at all.
 *
 * @param eventType The type of event to dispatch.
 * @param codeViewProps the codeViewProps props from the test cases.
 * @param position the position of the event.
 */
export const dispatchMouseEventAtPositionImpure = (
    eventType: 'click' | 'mouseover' | 'mousemove',
    { codeView, getCodeElementFromLineNumber }: CodeViewProps,
    position: Position
): void => {
    const line = getCodeElementFromLineNumber(codeView, position.line)
    if (!line) {
        throw new Error(invalidPosition(position, 'Line not found'))
    }

    if (!line.classList.contains('annotated')) {
        convertNode(line)
    }

    let characterOffset = 0
    for (const child of line.childNodes) {
        const value = child.textContent
        if (!value) {
            continue
        }

        if (characterOffset <= position.character && characterOffset + value.length >= position.character) {
            const rectangle = line.getBoundingClientRect()
            const { top, height, left, width } = rectangle

            const event = createMouseEvent(eventType, {
                x: left + width / 2,
                y: top + height / 2,
            })

            child.dispatchEvent(event)

            return
        }

        characterOffset += value.length
    }
}
