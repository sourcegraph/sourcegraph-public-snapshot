import { type Extension } from '@codemirror/state'
import { EditorView, type PluginValue, ViewPlugin } from '@codemirror/view'
import { fromEvent, Subscription } from 'rxjs'
import { debounceTime, filter, map, scan } from 'rxjs/operators'

import { type Occurrence } from '@sourcegraph/shared/src/codeintel/scip'

import { preciseOffsetAtCoords } from '../utils'

import { getCodeIntelAPI } from './api'
import { getTooltipViewFor, hasTooltipAtOffset, hideTooltipForKey, showCodeIntelTooltipAtRange } from './tooltips'

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

const HOVER_DEBOUNCE_TIME = 25 // ms
/**
 * The MouseEvent uses numbers to indicate which button was pressed.
 * See https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/buttons#value
 */
const MOUSE_NO_BUTTON = 0

/**
 * Listens to mousemove events, determines whether the position under the mouse
 * cursor is a valid {@link Occurrence}, fetches hover information as necessary and updates {@link codeIntelTooltipsState}.
 */
const hoverManager = ViewPlugin.fromClass(
    class HoverManager implements PluginValue {
        private subscription: Subscription = new Subscription()

        constructor(private readonly view: EditorView) {
            this.subscription.add(
                fromEvent<MouseEvent>(this.view.dom, 'mousemove')
                    .pipe(
                        // Debounce events so that users can move over tokens without triggering hovercards immediately
                        debounceTime(HOVER_DEBOUNCE_TIME),

                        // Ignore some events
                        filter(event => {
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
                            return !hasTooltipAtOffset(view.state, offset, 'hover')
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
                                const tooltipView = getTooltipViewFor(view, 'hover')
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

                            const offset = preciseOffsetAtCoords(this.view, position)
                            if (offset) {
                                const range = getCodeIntelAPI(view.state).findOccurrenceRangeAt(offset, view.state)
                                if (range) {
                                    return range
                                }
                            }

                            // If there is no new occurrence and the mouse is moving away from an existin tooltip,
                            // we want the existing tooltip to hide
                            return position.direction !== 'towards' ? ('HIDE' as const) : null
                        })
                    )
                    .subscribe(next => {
                        if (next === 'HIDE') {
                            view.dispatch(hideTooltipForKey('hover'))
                        } else if (next) {
                            showCodeIntelTooltipAtRange(view, next, 'hover')
                        }
                    })
            )
        }

        public destroy(): void {
            this.subscription.unsubscribe()
        }
    }
)

/*
                const { navigate, location } = update.state.facet(blobPropsFacet)
                const params = new URLSearchParams(location.search)
                params.delete('popover')
                window.requestAnimationFrame(() =>
                    // Use `navigate(to)` instead of `navigate(to, { replace: true })` in case
                    // the user accidentally clicked somewhere without intending to
                    // dismiss the popover.
                    navigate({ search: formatSearchParameters(params) })
                )
                */

const tooltipStyles = EditorView.theme({
    // Tooltip styles is a combination of the default wildcard PopoverContent component (https://github.com/sourcegraph/sourcegraph/blob/5de30f6fa1c59d66341e4dfc0c374cab0ad17bff/client/wildcard/src/components/Popover/components/popover-content/PopoverContent.module.scss#L1-L10)
    // and the floating tooltip-like storybook usage example (https://github.com/sourcegraph/sourcegraph/blob/5de30f6fa1c59d66341e4dfc0c374cab0ad17bff/client/wildcard/src/components/Popover/story/Popover.story.module.scss#L54-L62)
    // ignoring the min/max width rules.
    '.cm-tooltip.tmp-tooltip': {
        fontSize: '0.875rem',
        backgroundClip: 'padding-box',
        backgroundColor: 'var(--dropdown-bg)',
        border: '1px solid var(--dropdown-border-color)',
        borderRadius: 'var(--popover-border-radius)',
        color: 'var(--body-color)',
        boxShadow: 'var(--dropdown-shadow)',
        padding: '0.5rem',
    },

    '.cm-tooltip.cm-tooltip-above.tmp-tooltip .cm-tooltip-arrow:before': {
        borderTopColor: 'var(--dropdown-border-color)',
    },
    '.cm-tooltip.cm-tooltip-above.tmp-tooltip .cm-tooltip-arrow:after': {
        borderTopColor: 'var(--dropdown-bg)',
    },
})

export const hoverExtension: Extension = [hoverManager, tooltipStyles]
