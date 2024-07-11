import type { Extension } from '@codemirror/state'
import { type EditorView, type PluginValue, ViewPlugin, getTooltip, type Tooltip } from '@codemirror/view'
import { from, fromEvent, Subscription, timer, race } from 'rxjs'
import { debounceTime, filter, map, scan, switchMap, tap } from 'rxjs/operators'

import { createUpdateableField } from '@sourcegraph/shared/src/components/codemirror/utils'

import { preciseOffsetAtCoords } from '../utils'

import { findOccurrenceRangeAt, getHoverTooltip } from './api'
import { showHasDefinition } from './definition'
import { type TooltipSource, showTooltip } from './tooltips'

/**
 * This field stores various information about the currently hovered range, which
 * is used to provide tooltips via {@link showTooltip} and definition highlighting
 * via {@link showHasDefinition}.
 */
const [hoverTooltip, setHoverTooltip] = createUpdateableField<{
    range: { from: number; to: number }
    source: TooltipSource
} | null>(null, self => [
    showTooltip.computeN([self], state => {
        const field = state.field(self)
        return field ? [field.source] : []
    }),
    showHasDefinition.computeN([self], state => {
        const range = state.field(self)?.range
        return range ? [range] : []
    }),
])

/**
 * This function computes whether the mouse moves towards or away
 * from the rectable {@param rect}.
 */
function computeMouseDirection(
    rect: DOMRect,
    position1: { x: number; y: number },
    position2: { x: number; y: number }
): 'towards' | 'away' {
    if (
        // Moves away from the top
        (position2.y < position1.y && position2.y < rect.top) ||
        // Moves away from the bottom
        (position2.y > position1.y && position2.y > rect.bottom) ||
        // Moves away from the left
        (position2.x < position1.x && position2.x < rect.left) ||
        // Moves away from the right
        (position2.x > position1.x && position2.x > rect.right)
    ) {
        return 'away'
    }

    return 'towards'
}

const MOUSEMOVE_DEBOUNCE_TIME = 25 // ms
const MOUSELEAVE_DEBOUNCE_TIME = 250 // ms
/**
 * The MouseEvent uses numbers to indicate which button was pressed.
 * See https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/buttons#value
 */
const MOUSE_NO_BUTTON = 0

/**
 * Listens to mousemove events, determines whether the position under the mouse
 * cursor is a valid {@link Occurrence}, fetches hover information as necessary and
 * updates {@link hoverTooltip}.
 */
const hoverManager = ViewPlugin.fromClass(
    class HoverManager implements PluginValue {
        private subscription: Subscription = new Subscription()
        /**
         * A reference to the currently shown tooltip. This is needed to be able to determine
         * whether the mouse moves towards or away from the currently open tooltip.
         */
        private tooltip: (Tooltip & { end: number }) | null = null

        constructor(private readonly view: EditorView) {
            let mouseHasLeft = false

            // The 'mousemove' event handler below hides the popover when the mouse
            // moves away from the tooltip, but it doesn't work in the case when the
            // mouse "leaves" the file view via a popover ('mousemove' doesn't trigger):
            //
            //             ▲      Popover
            //         ┌───│────────────┐
            //         │███│████████████│     Editor
            //    ┌────│███│█Mouse██████│──────────┐
            //    │    │███│████████████│          │
            //    │    └───│────────────┘          │
            //    │ const █│█ = ...                │
            //    │                                │
            //    └────────────────────────────────┘
            //
            // To handle this case we listen to 'mouseleave' events on the editor view,
            // and hide the popover when the event is triggered.
            this.subscription.add(
                fromEvent(this.view.dom, 'mouseleave')
                    .pipe(
                        // We are debouncing 'mousemove' events. To ensure that
                        // 'mousemove' events are ignored after the mouse has left
                        // the editor view we are setting a flag here.
                        tap(() => (mouseHasLeft = true)),

                        switchMap(() =>
                            race(
                                // By delaying the 'mouseleave' event here we are allowing the
                                // user to move the mouse back to the popover to keep it open.
                                // This is helpful when the user accidentally moves the mouse
                                // out of the popover.
                                timer(MOUSELEAVE_DEBOUNCE_TIME).pipe(map(() => true)),
                                fromEvent(this.view.dom, 'mouseenter').pipe(map(() => false))
                            )
                        )
                    )
                    .subscribe(hide => {
                        if (hide) {
                            this.tooltip = null
                            setHoverTooltip(view, null)
                        }
                    })
            )

            this.subscription.add(
                fromEvent<MouseEvent>(this.view.dom, 'mousemove')
                    .pipe(
                        // Debounce events so that users can move over tokens without triggering hovercards immediately
                        debounceTime(MOUSEMOVE_DEBOUNCE_TIME),

                        // Ignore some events
                        filter(event => {
                            // Ignore events when the mouse has left the editor view.
                            // Keeping this flag ensures that debounced 'mousemove' events are ignored
                            // after the mouse has left the editor view.
                            // We are also resetting the flag here to ensure that the hover manager isn't "stuck"
                            // in a state where it ignores all events.
                            if (mouseHasLeft) {
                                mouseHasLeft = false
                                return false
                            }

                            // Ignore events when hovering over an existing hovercard.
                            // This causes existing hovercards to stay open.
                            if (
                                (event.target as HTMLElement | null)?.closest(
                                    '.cm-code-intel-hovercard:not(.cm-code-intel-hovercard-pinned)'
                                )
                            ) {
                                return false
                            }

                            // We have to forward any move events that also have a
                            // button pressed. User is probably selecting text and
                            // hovercards should be hidden.
                            if (event.buttons !== MOUSE_NO_BUTTON) {
                                return true
                            }

                            // Ignore events inside the current hover range. Without this
                            // hovercards flicker when the active range is wider than the
                            // word-under-cursor range. For example, hovering over
                            //
                            // import ( "io/fs" )
                            //
                            // will detect `io` and `fs` as separate words (and would
                            // therefore trigger two individual word lookups), but the
                            // hover information returned by the server is for the whole
                            // `io/fs` range.
                            const offset = preciseOffsetAtCoords(view, event)
                            if (offset === null) {
                                return true
                            }

                            // Process event if the current hover location is outside the current hover
                            // occurrence (if any)
                            return !(this.tooltip && this.tooltip.pos <= offset && offset <= this.tooltip.end)
                        }),

                        // To make it easier to reach the tooltip with the mouse, we determine
                        // in which direction the mouse moves and only hide the tooltip when
                        // the mouse moves away from it.
                        scan(
                            (
                                previous: {
                                    x: number
                                    y: number
                                    target: EventTarget | null
                                    buttons: number
                                    direction?: 'towards' | 'away' | undefined
                                },
                                next
                            ) => {
                                const tooltipView = this.tooltip && getTooltip(this.view, this.tooltip)
                                if (!tooltipView) {
                                    return next
                                }

                                const direction = computeMouseDirection(
                                    tooltipView.dom.getBoundingClientRect(),
                                    previous,
                                    next
                                )
                                return { x: next.x, y: next.y, buttons: next.buttons, target: next.target, direction }
                            }
                        ),

                        // Determine the precise location of the word under the cursor.
                        map(position => {
                            // Hide any tooltip when
                            // - the mouse is over an element that is not part of
                            //   the content. This seems necessary to make tooltips
                            //   not appear and hide open tooltips when the mouse
                            //   moves over the editor's search panel.
                            // - the user starts to select text
                            if (
                                position.buttons !== MOUSE_NO_BUTTON ||
                                !position.target ||
                                !this.view.contentDOM.contains(position.target as Node)
                            ) {
                                return 'HIDE' as const
                            }

                            // Always show the tooltip when user hovers over new token
                            const offset = preciseOffsetAtCoords(this.view, position)
                            if (offset) {
                                const range = findOccurrenceRangeAt(view.state, offset)
                                if (range) {
                                    return range
                                }
                            }

                            // If there is no new occurrence and the mouse is moving away from an existing tooltip,
                            // we want the existing tooltip to hide
                            return position.direction !== 'towards' ? ('HIDE' as const) : null
                        })
                    )
                    .subscribe(next => {
                        if (next === 'HIDE') {
                            this.tooltip = null
                            setHoverTooltip(view, null)
                        } else if (next) {
                            setHoverTooltip(view, {
                                range: next,
                                source: from(getHoverTooltip(view.state, next.from)).pipe(
                                    // We need to retain a reference to the created tooltip so that
                                    // we can compute whether the pointer moves away or towards it.
                                    tap(tooltip => (this.tooltip = tooltip))
                                ),
                            })
                        }
                    })
            )
        }

        public destroy(): void {
            this.subscription.unsubscribe()
        }
    }
)

/**
 * This extension will track the token underneath the cursor and show
 * code intel tooltips when available.
 */
export const hoverExtension: Extension = [hoverManager, hoverTooltip]
