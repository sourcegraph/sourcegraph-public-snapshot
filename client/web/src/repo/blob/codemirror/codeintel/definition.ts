import { type Extension, Facet, type Range as CodeMirrorRange } from '@codemirror/state'
import { Decoration, EditorView, type Tooltip } from '@codemirror/view'
import { from } from 'rxjs'

import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { MOUSE_MAIN_BUTTON, preciseOffsetAtCoords } from '../utils'

import { findOccurrenceRangeAt, goToDefinitionAt, hasDefinitionAt } from './api'
import { codeIntelDecorations } from './decorations'
import { showTooltip } from './tooltips'
import { type UpdateableValue, createLoaderExtension, isModifierKey } from './utils'

interface Range {
    from: number
    to: number
}

const LONG_CLICK_DURATION = 500

// Ranges with definitions are underlined when the user holds the Mod key.
// The styles are definied in modifier-click.
const hasDefinitionDeco = Decoration.mark({ class: 'sg-definition-available' })

/**
 * HasDefinition keeps definition-availability state for a given range.
 */
class HasDefinition implements UpdateableValue<boolean, HasDefinition> {
    /**
     * Passing `null` to {@param hasDefinition} (default) indicates that
     * definition information for this range needs to be fetched.
     */
    constructor(public range: Range, public hasDefinition: boolean | null = null) {}

    public update(hasDefinition: boolean): HasDefinition {
        return new HasDefinition(this.range, hasDefinition)
    }

    public get key(): Range {
        return this.range
    }

    public get isPending(): boolean {
        return this.hasDefinition === null
    }
}

/**
 * Position to check whether or not they have a definition associated with it.
 * If yes they are marked with the .sg-definition-available class.
 */
export const showHasDefinition = Facet.define<Range>({
    enables: self => [
        createLoaderExtension({
            input(state) {
                return state.facet(self)
            },
            create(range) {
                return new HasDefinition(range, null)
            },
            load(request, state) {
                return from(hasDefinitionAt(state, request.range.from))
            },
            provide: self => [
                // Style ranges that have definitions available
                codeIntelDecorations.compute([self], state => {
                    const decorations: CodeMirrorRange<Decoration>[] = []

                    for (const { range, hasDefinition } of state.field(self)) {
                        if (hasDefinition) {
                            decorations.push(hasDefinitionDeco.range(range.from, range.to))
                        }
                    }

                    return Decoration.set(decorations, true)
                }),
            ],
        }),
    ],
})

const [tooltip, setTooltip] = createUpdateableField<Tooltip | null>(null, field => showTooltip.from(field))

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
    const position = preciseOffsetAtCoords(view, { x: event.clientX, y: event.clientY })
    if (position === null) {
        return
    }
    const occurrence = findOccurrenceRangeAt(view.state, position)
    if (!occurrence) {
        return
    }
    if (!isModifierKey(event) && !options?.isLongClick) {
        return
    }

    setTooltip(view, new LoadingTooltip(occurrence.from, occurrence.to))
    void goToDefinitionAt(view, occurrence.from).finally(() => {
        setTooltip(view, null)
    })
}

/**
 * Extension which goes to a definition on long click or mod-click
 */
export function goToDefinitionOnClick(): Extension {
    const events = new MouseEvents()

    return [
        tooltip,
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
            mousemove(event) {
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
        }),
    ]
}
