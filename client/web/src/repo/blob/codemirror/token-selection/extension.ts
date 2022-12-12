import { Extension } from '@codemirror/state'
import { EditorView, hoverTooltip, keymap } from '@codemirror/view'

import { isInteractiveOccurrence, occurrenceAtMouseEvent } from '../occurrence-utils'
import { MOUSE_MAIN_BUTTON } from '../utils'

import { definitionCache, goToDefinitionOnMouseEvent, underlinedDefinitionFacet } from './definition'
import { documentHighlightsExtension } from './document-highlights'
import { getHoverTooltip, hoverCache, hoveredOccurrenceField, hoverField, setHoveredOccurrenceEffect } from './hover'
import { tokenSelectionKeyBindings } from './keybindings'
import { modifierClickFacet } from './modifier-click'
import { selectedOccurrence, syncSelectionWithURL, tokenSelectionTheme, warmupOccurrence } from './selections'

const LONG_CLICK_DURATION = 500

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

// Heuristic to approximate a click event between two mouse events (for
// exampole, mousedown and mouseup). The heuristic returns true based on the
// distance between the coordinates (clientY/clientX) of the two events.
function isClickDistance(event1: MouseEvent, event2: MouseEvent): boolean {
    const distanceX = Math.abs(event1.clientX - event2.clientX)
    const distanceY = Math.abs(event1.clientY - event2.clientY)
    const distance = distanceX + distanceY
    // The number is picked out of the blue, feel free to tweak this
    // heuristic if it's too lenient or too aggressive.
    return distance < 2
}

export function tokenSelectionExtension(): Extension {
    const events = new MouseEvents()
    return [
        documentHighlightsExtension(),
        modifierClickFacet.of(false),
        tokenSelectionTheme,
        selectedOccurrence.of(null),
        definitionCache,
        underlinedDefinitionFacet.of(null),
        hoveredOccurrenceField,
        hoverCache,
        hoverTooltip((view, position) => getHoverTooltip(view, position), { hoverTime: 200, hideOnChange: true }),
        hoverField,
        keymap.of(tokenSelectionKeyBindings),
        syncSelectionWithURL,
        EditorView.domEventHandlers({
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
            mousedown(event, view) {
                if (event.button !== MOUSE_MAIN_BUTTON) {
                    return
                }
                events.mousedown = event
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
            click(event: MouseEvent) {
                // Prevent click handlers because we handle events on mouseup.
                // Without the line below, Cmd+Click gets unpredicable behavior
                // where it sporadically opens goto-def links in a new tab.
                event.preventDefault()
            },
            mousemove(event, view) {
                events.mousemove = event
                const atEvent = occurrenceAtMouseEvent(view, event)
                const hoveredOccurrence = atEvent ? atEvent.occurrence : null
                if (hoveredOccurrence !== view.state.field(hoveredOccurrenceField)) {
                    view.dispatch({ effects: setHoveredOccurrenceEffect.of(hoveredOccurrence) })
                    // Only pre-fetch/warmup codeintel for interactive tokens.
                    // Users can still trigger code intel actions for
                    // non-interactive tokens, the result just won't be
                    // pre-fetched.
                    if (hoveredOccurrence && isInteractiveOccurrence(hoveredOccurrence)) {
                        warmupOccurrence(view, hoveredOccurrence)
                    }
                }
            },
        }),
    ]
}
