import { StateEffect, StateField } from '@codemirror/state'
import { Occurrence } from '@sourcegraph/shared/out/src/codeintel/scip'
import { EditorView, getTooltip, PluginValue, showTooltip, Tooltip, ViewPlugin, ViewUpdate } from '@codemirror/view'

import { positionToOffset, preciseOffsetAtCoords, uiPositionToOffset } from '../utils'
import { occurrenceAtPosition, positionAtCmPosition, rangeToCmSelection } from '../occurrence-utils'
import { showDocumentHighlightsForOccurrence } from './document-highlights'
import { getHoverTooltip, hoverCache } from './hover'
import { BehaviorSubject, from, fromEvent, of, Subject, Subscription } from 'rxjs'
import { catchError, debounceTime, filter, map, scan, switchMap, tap } from 'rxjs/operators'
import { computeMouseDirection, HOVER_DEBOUNCE_TIME, MOUSE_NO_BUTTON, pin } from '../hovercard'
import { blobPropsFacet } from '../index'
import { LoadingTooltip } from '../tooltips/LoadingTooltip'
import { formatSearchParameters, LineOrPositionOrRange } from '@sourcegraph/common/src'
import { CodeIntelTooltip } from '../tooltips/CodeIntelTooltip'
import { warmupOccurrence } from './selections'

type CodeIntelTooltipTrigger = 'focus' | 'hover' | 'pin'
type CodeIntelTooltipState = { occurrence: Occurrence; tooltip: LoadingTooltip | CodeIntelTooltip | null } | null

const setFocusedOccurrence = StateEffect.define<Occurrence | null>()
export const setFocusedOccurrenceTooltip = StateEffect.define<LoadingTooltip | CodeIntelTooltip | null>()
const setPinnedCodeIntelTooltipState = StateEffect.define<CodeIntelTooltipState>()
const setHoveredCodeIntelTooltipState = StateEffect.define<CodeIntelTooltipState>()

export const codeIntelTooltipsState = StateField.define<Record<CodeIntelTooltipTrigger, CodeIntelTooltipState>>({
    create() {
        return { hover: null, focus: null, pin: null }
    },
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setFocusedOccurrence)) {
                return {
                    ...value,
                    focus: effect.value ? { occurrence: effect.value, tooltip: null } : null,
                }
            }
            if (effect.is(setFocusedOccurrenceTooltip)) {
                // TODO: check if tooltip pos matches selected occurrence
                return {
                    ...value,
                    focus: value.focus?.occurrence ? { ...value.focus, tooltip: effect.value } : null,
                }
            }

            if (effect.is(setHoveredCodeIntelTooltipState)) {
                return { ...value, hover: effect.value }
            }

            if (effect.is(setPinnedCodeIntelTooltipState)) {
                return { ...value, pin: effect.value }
            }
        }
        if (transaction.selection) {
            console.log(transaction.selection.main)
        }
        return value
    },
    provide(field) {
        return [
            showTooltip.computeN([field], state => {
                const { hover, focus, pin } = state.field(field)
                const tooltips = []
                if (focus?.tooltip) {
                    tooltips.push(focus.tooltip)
                }
                if (pin?.tooltip) {
                    tooltips.push(pin.tooltip)
                }
                if (hover?.tooltip && hover.occurrence !== pin?.occurrence && hover.occurrence !== focus?.occurrence) {
                    tooltips.push(hover.tooltip)
                }
                return tooltips
            }),
        ]
    },
})

export const getCodeIntelTooltipState = (
    view: EditorView,
    key: CodeIntelTooltipTrigger
): { occurrence: Occurrence; tooltip: Tooltip | null } | null => view.state.field(codeIntelTooltipsState)?.[key]

const focusOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    const offset = positionToOffset(view.state.doc, occurrence.range.end)
    if (offset !== null) {
        const node = view.domAtPos(offset).node
        const el = node instanceof HTMLElement ? node : node.parentElement
        const lineEl = el?.matches('.cm-line') ? el : el?.closest('.cm-line')
        const interactiveOccurrenceEl = lineEl?.querySelector<HTMLElement>('.interactive-occurrence')
        if (interactiveOccurrenceEl) {
            interactiveOccurrenceEl?.focus()
        }
    }
}

export const selectOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    warmupOccurrence(view, occurrence)
    view.dispatch({
        effects: setFocusedOccurrence.of(occurrence),
        selection: rangeToCmSelection(view.state, occurrence.range),
    })
    showDocumentHighlightsForOccurrence(view, occurrence)
    focusOccurrence(view, occurrence)
}

/**
 * Listens to mousemove events, determines whether the position under the mouse
 * cursor is a valid {@link Occurrence}, fetches hover information as necessary and updates {@link hoverTooltip}.
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

                            const currentOccurrence = getCodeIntelTooltipState(view, 'hover')?.occurrence
                            if (!currentOccurrence) {
                                return true
                            }

                            return !isOffsetInHoverRange(
                                offset,
                                rangeToCmSelection(this.view.state, currentOccurrence.range)
                            )
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
                                const currentTooltip = getCodeIntelTooltipState(view, 'hover')?.tooltip
                                if (!currentTooltip) {
                                    return next
                                }

                                const tooltipView = getTooltip(view, currentTooltip)
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
                        switchMap(position => {
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
                                return of('HIDE' as const)
                            }

                            const offset = preciseOffsetAtCoords(this.view, position)
                            if (!offset) {
                                return of(null)
                            }
                            const pos = positionAtCmPosition(view, offset)
                            return of(occurrenceAtPosition(this.view.state, pos)).pipe(
                                tap(occurrence => {
                                    const current = getCodeIntelTooltipState(view, 'hover')?.occurrence
                                    if (current && occurrence && current !== occurrence) {
                                        view.dispatch({
                                            effects: setHoveredCodeIntelTooltipState.of(null),
                                        })
                                    }
                                }),
                                switchMap(occurrence => {
                                    if (!occurrence) {
                                        return of(null)
                                    }

                                    const offset = positionToOffset(this.view.state.doc, occurrence.range.start)
                                    if (offset === null) {
                                        return of(null)
                                    }

                                    // show loading tooltip
                                    view.dispatch({
                                        effects: setHoveredCodeIntelTooltipState.of({
                                            occurrence,
                                            tooltip: new LoadingTooltip(offset),
                                        }),
                                    })

                                    return from(getHoverTooltip(view, offset)).pipe(
                                        catchError(() => of(null)),

                                        // close loading tooltip
                                        tap(() => {
                                            const current = getCodeIntelTooltipState(view, 'hover')
                                            if (
                                                current?.tooltip instanceof LoadingTooltip &&
                                                current?.occurrence === occurrence
                                            ) {
                                                view.dispatch({ effects: setHoveredCodeIntelTooltipState.of(null) })
                                            }
                                        }),

                                        map(tooltip => (tooltip ? { tooltip, occurrence } : null))
                                    )
                                }),
                                map(hover => ({ position, hover }))
                            )
                        })
                    )
                    .subscribe(next => {
                        if (next === 'HIDE') {
                            view.dispatch({ effects: setHoveredCodeIntelTooltipState.of(null) })
                            return
                        }

                        // We only change the tooltip when
                        // a) There is a new tooltip at the position (tooltip !== null)
                        // b) there is no tooltip and the mouse is moving away from the tooltip
                        if (next?.hover || next?.position.direction !== 'towards') {
                            if (!next?.hover?.occurrence) {
                                view.dispatch({
                                    effects: setHoveredCodeIntelTooltipState.of(null),
                                })
                                return
                            }

                            view.dispatch({
                                effects: setHoveredCodeIntelTooltipState.of(next.hover),
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

function isOffsetInHoverRange(offset: number, range: { from: number; to: number }): boolean {
    return range.from <= offset && offset <= range.to
}

const getPinnedOccurrence = (view: EditorView, pin: LineOrPositionOrRange | null): Occurrence | null => {
    if (!pin || !pin.line || !pin.character) {
        return null
    }
    const offset = uiPositionToOffset(view.state.doc, { line: pin.line, character: pin.character })
    if (offset === null) {
        return null
    }
    return occurrenceAtPosition(view.state, positionAtCmPosition(view, offset)) ?? null
}

// Extension that automatically displays the code-intel popover when the URL has
// `popover=pinned`, and removed this URL parameter when the user clicks
// anywhere on the file to dismiss the pinned popover.
const pinManager = ViewPlugin.fromClass(
    class implements PluginValue {
        private nextPin: Subject<LineOrPositionOrRange | null>
        private subscription: Subscription

        constructor(view: EditorView) {
            this.nextPin = new BehaviorSubject(view.state.field(pin))
            this.subscription = this.nextPin
                .pipe(
                    map(pin => getPinnedOccurrence(view, pin)),
                    tap(occurrence => {
                        if (!occurrence) {
                            window.requestAnimationFrame(() =>
                                view.dispatch({ effects: setPinnedCodeIntelTooltipState.of(null) })
                            )
                        }
                    }),
                    switchMap(occurrence => {
                        if (!occurrence) {
                            return of(null)
                        }

                        return from(getHoverTooltip(view, rangeToCmSelection(view.state, occurrence.range).from)).pipe(
                            map(tooltip => ({ occurrence, tooltip }))
                        )
                    })
                )
                .subscribe(pin =>
                    // Scheduling the update for the next loop is necessary at the
                    // moment because we are triggering this effect in response to an
                    // editor update (pin field change) and you cannot synchronously
                    // trigger an update from an update.
                    window.requestAnimationFrame(() =>
                        view.dispatch({ effects: setPinnedCodeIntelTooltipState.of(pin) })
                    )
                )
        }

        public update(update: ViewUpdate): void {
            if (update.startState.field(pin) !== update.state.field(pin)) {
                this.nextPin.next(update.state.field(pin))
            }

            if (update.selectionSet && update.state.field(pin)) {
                // Remove `popover=pinned` from the URL when the user updates the selection.
                const history = update.state.facet(blobPropsFacet).history
                const params = new URLSearchParams(history.location.search)
                params.delete('popover')
                window.requestAnimationFrame(() =>
                    // Use `history.push` instead of `history.replace` in case
                    // the user accidentally clicked somewhere without intending to
                    // dismiss the popover.
                    history.push({ ...history.location, search: formatSearchParameters(params) })
                )
            }
        }

        public destroy(): void {
            this.subscription.unsubscribe()
        }
    }
)

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

    '.cm-tooltip-above:not(.tmp-tooltip), .cm-tooltip-below:not(.tmp-tooltip)': {
        border: 'unset',
    },

    '.cm-tooltip.cm-tooltip-above.tmp-tooltip .cm-tooltip-arrow:before': {
        borderTopColor: 'var(--dropdown-border-color)',
    },
    '.cm-tooltip.cm-tooltip-above.tmp-tooltip .cm-tooltip-arrow:after': {
        borderTopColor: 'var(--dropdown-bg)',
    },
})

export function codeIntelTooltipsExtension() {
    return [codeIntelTooltipsState, hoverCache, hoverManager, pinManager, tooltipStyles]
}
