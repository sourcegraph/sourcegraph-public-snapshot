import { type Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { MOUSE_MAIN_BUTTON, preciseOffsetAtCoords } from '../utils'

import { getCodeIntelAPI } from './api'
import { isModifierKey } from './modifier-click'

const LONG_CLICK_DURATION = 500

// Heuristic to approximate a click event between two mouse events (for
// example, mousedown and mouseup). The heuristic returns true based on the
// distance between the coordinates (clientY/clientX) of the two events.
function isClickDistance(event1: MouseEvent, event2: MouseEvent): boolean {
    const distanceX = Math.abs(event1.clientX - event2.clientX)
    const distanceY = Math.abs(event1.clientY - event2.clientY)
    const distance = distanceX + distanceY
    // The number is picked out of the blue, feel free to tweak this
    // heuristic if it's too lenient or too aggressive.
    return distance < 2
}

// Helper class to deal with mouse events for features like long-click.
class MouseEvents {
    // Last mousemove event.
    public mousemove = new MouseEvent('mouseover')
    // Last mousedown event.
    public mousedown = new MouseEvent('mousedown')
    // Timeout to cancel the long-click handler.
    public longClickTimeout: NodeJS.Timeout | undefined
    // Boolean to indicate that we triggered a long-click meaning we should
    // ignore the next mouseup event.
    public didLongClick = false
    public isLongClick(): boolean {
        return isClickDistance(this.mousedown, this.mousemove)
    }
    public isMouseupClick(mouseup: MouseEvent): boolean {
        return isClickDistance(mouseup, this.mousedown)
    }
}

function goToDefinitionOnMouseEvent(view: EditorView, event: MouseEvent, options?: { isLongClick?: boolean }): void {
    const api = getCodeIntelAPI(view.state)
    const position = preciseOffsetAtCoords(view, { x: event.clientX, y: event.clientY })
    if (position === null) {
        return
    }
    const occurrence = api.findOccurrenceRangeAt(position, view.state)
    if (!occurrence) {
        return
    }
    if (!isModifierKey(event) && !options?.isLongClick) {
        return
    }

    // TODO: show loading toolip
    api.goToDefinitionAt(view, occurrence.from)
}

/**
 * Extension which goes to a definition on long click or mod-click
 */
function goToDefinitionOnClick(): Extension {
    const events = new MouseEvents()

    return EditorView.domEventHandlers({
        // Approximate `click` with `mouseup` because `click` does not get
        // triggered in the scenario when the user holds down the meta-key,
        // waits for the underline decoration to get updated in the dom and
        // then clicks. We approximate the click handler by detecting
        // mouseup events that fire close to the last mousedown event.
        mouseup(event, view) {
            if (event.button !== MOUSE_MAIN_BUTTON) {
                return
            }
            if (events.longClickTimeout) {
                // Cancel the scheduled long-click handler.
                clearTimeout(events.longClickTimeout)
            }
            if (events.didLongClick) {
                // Cancel this mouseup event because a long-click was triggered.
                events.didLongClick = false
                return
            }
            if (events.isMouseupClick(event)) {
                goToDefinitionOnMouseEvent(view, event)
            }
        },
        mousemove(event, view) {
            events.mousemove = event
        },
        mousedown(event, view) {
            if (event.button !== MOUSE_MAIN_BUTTON) {
                return
            }
            events.mousedown = events.mousemove = event
            events.longClickTimeout = setTimeout(() => {
                if (!events.isLongClick()) {
                    // Cancel this long-click because the mouse has moved
                    // too far of a distance.
                    return
                }
                events.didLongClick = true // Cancel the next mouseup event.
                goToDefinitionOnMouseEvent(view, event, { isLongClick: true })
            }, LONG_CLICK_DURATION)
        },
    })
}

export function goToDefinitionExtension() {
    return [goToDefinitionOnClick()]
}
