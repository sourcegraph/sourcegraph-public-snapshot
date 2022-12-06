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

const LONGPRESS_DURATION = 500

export function tokenSelectionExtension(): Extension {
    let clientY = 0
    let clientX = 0
    let longPressTimeout: NodeJS.Timeout | undefined
    let didLongpress = false
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
            // We use `mouseup` instead of `click` because `click` does not get
            // triggered in the scenario when the user holds down the meta-key,
            // waits for the underline decoration to get updated in the dom and
            // then clicks. We approximate the click handler by detecting
            // mouseup events that fire close to the last mousedown event.
            mouseup(event, view) {
                if (longPressTimeout) {
                    clearTimeout(longPressTimeout)
                }
                if (event.button !== MOUSE_MAIN_BUTTON) {
                    return
                }
                if (didLongpress) {
                    didLongpress = false
                    return
                }
                const distanceX = event.clientX - clientX
                const distanceY = event.clientY - clientY
                const distance = distanceX + distanceY
                // Feel free to tweak this heuristic if it's too lenient or too
                // aggressive.
                const isClick = distance < 10
                if (isClick) {
                    goToDefinitionOnMouseEvent(view, event)
                }
            },
            mousedown(event, view) {
                if (event.button !== MOUSE_MAIN_BUTTON) {
                    return
                }
                clientX = event.clientX
                clientY = event.clientY
                longPressTimeout = setTimeout(() => {
                    didLongpress = true
                    goToDefinitionOnMouseEvent(view, event, { isLongPress: true })
                }, LONGPRESS_DURATION)
            },
            click(event: MouseEvent) {
                // Prevent click handlers because we handle events on mouseup.
                // Without the line below, Cmd+Click gets unpredicable behavior
                // where it sporadically opens goto-def links in a new tab.
                event.preventDefault()
            },
            mousemove(event, view) {
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
