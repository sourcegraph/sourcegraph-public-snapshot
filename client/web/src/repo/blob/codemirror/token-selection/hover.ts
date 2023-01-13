import { countColumn, Extension, SelectionRange, StateEffect, StateField } from '@codemirror/state'
import {
    closeHoverTooltips,
    Decoration,
    DecorationSet,
    EditorView,
    getTooltip,
    PluginValue,
    showTooltip,
    Tooltip,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'
import { isEqual } from 'lodash'
import { BehaviorSubject, from, fromEvent, of, Subject, Subscription } from 'rxjs'
import { catchError, debounceTime, filter, map, scan, switchMap, tap } from 'rxjs/operators'

import { HoverMerged, TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { formatSearchParameters, LineOrPositionOrRange } from '@sourcegraph/common'
import { getOrCreateCodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'
import { Occurrence, Position } from '@sourcegraph/shared/src/codeintel/scip'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import {
    computeMouseDirection,
    HOVER_DEBOUNCE_TIME,
    MOUSE_NO_BUTTON,
    pin,
    selectionHighlightDecoration,
} from '../hovercard'
import {
    isInteractiveOccurrence,
    occurrenceAtPosition,
    positionAtCmPosition,
    rangeToCmSelection,
} from '../occurrence-utils'
import { CodeIntelTooltip, HoverResult } from '../tooltips/CodeIntelTooltip'
import { preciseOffsetAtCoords, preciseWordAtCoords, uiPositionToOffset } from '../utils'

export function hoverExtension(): Extension {
    return [hoverCache, hoveredOccurrenceField, hoverTooltip, hoverManager, tooltipStyles, hoverField, pinManager]
}
export const hoverCache = StateField.define<Map<Occurrence, Promise<HoverResult>>>({
    create: () => new Map(),
    update: value => value,
})

export const closeHover = (view: EditorView): void =>
    // Always emit `closeHoverTooltips` alongside `setHoverEffect.of(null)` to
    // fix an issue where the tooltip could get stuck if you rapidly press Space
    // before the tooltip finishes loading.
    view.dispatch({ effects: [setHoverEffect.of(null), closeHoverTooltips] })
export const showHover = (view: EditorView, tooltip: Tooltip): void =>
    view.dispatch({ effects: setHoverEffect.of(tooltip) })

// intentionally not exported because clients should use the close/open hover
// helpers above.
const setHoverEffect = StateEffect.define<Tooltip | null>()
export const hoverField = StateField.define<Tooltip | null>({
    create: () => null,
    update(tooltip, transactions) {
        if (transactions.docChanged || transactions.selection) {
            // Close hover when the selection moves and when the document
            // changes (although that should not happen because we only support
            // read-only mode).
            return null
        }
        for (const effect of transactions.effects) {
            if (effect.is(setHoverEffect)) {
                tooltip = effect.value
            }
        }
        return tooltip
    },
    provide: field => showTooltip.from(field),
})
export const setHoveredOccurrenceEffect = StateEffect.define<Occurrence | null>()
export const hoveredOccurrenceField = StateField.define<Occurrence | null>({
    create: () => null,
    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setHoveredOccurrenceEffect)) {
                value = effect.value
            }
        }
        return value
    },
})

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

export async function getHoverTooltip(view: EditorView, pos: number): Promise<Tooltip | null> {
    const cmLine = view.state.doc.lineAt(pos)
    const line = cmLine.number - 1
    const character = countColumn(cmLine.text, 1, pos - cmLine.from)
    const occurrence = occurrenceAtPosition(view.state, new Position(line, character))
    if (!occurrence) {
        return null
    }
    const result = await hoverAtOccurrence(view, occurrence)
    if (!result.markdownContents) {
        return null
    }
    const pinnedOccurrence = getPinnedOccurrence(view, view.state.field(pin))
    return new CodeIntelTooltip(view, occurrence, result, occurrence === pinnedOccurrence)
}

export function hoverAtOccurrence(view: EditorView, occurrence: Occurrence): Promise<HoverResult> {
    const cache = view.state.field(hoverCache)
    const fromCache = cache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const uri = toURIWithPath(view.state.facet(blobPropsFacet).blobInfo)
    const contents = hoverRequest(view, occurrence, {
        position: occurrence.range.start,
        textDocument: { uri },
    })
    cache.set(occurrence, contents)
    return contents
}

// Extension that automatically displays the code-intel popover when the URL has
// `popover=pinned`, and removed this URL parameter when the user clicks
// anywhere on the file to dismiss the pinned popover.
const pinManager = ViewPlugin.fromClass(
    class implements PluginValue {
        public decorations: DecorationSet
        private nextPin: Subject<LineOrPositionOrRange | null>
        private subscription: Subscription

        constructor(view: EditorView) {
            this.decorations = Decoration.none
            this.nextPin = new BehaviorSubject(view.state.field(pin))
            this.subscription = this.nextPin
                .pipe(
                    map(pin => {
                        const occurrence = getPinnedOccurrence(view, pin)
                        return occurrence ? rangeToCmSelection(view.state, occurrence.range) : null
                    }),
                    tap(range => {
                        if (!range) {
                            this.computeDecorations(null)
                        }
                    }),
                    switchMap(range =>
                        range
                            ? from(getHoverTooltip(view, range.from)).pipe(
                                  tap(tooltip => this.computeDecorations(tooltip ? range : null))
                              )
                            : of(null)
                    )
                )
                .subscribe(tooltip =>
                    // Scheduling the update for the next loop is necessary at the
                    // moment because we are triggering this effect in response to an
                    // editor update (pin field change) and you cannot synchronously
                    // trigger an update from an update.
                    window.requestAnimationFrame(() => view.dispatch({ effects: setHoverEffect.of(tooltip) }))
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

        private computeDecorations(range: SelectionRange | null): void {
            this.decorations = range
                ? Decoration.set(selectionHighlightDecoration.range(range.from, range.to))
                : Decoration.none
        }
    },
    {
        decorations: ({ decorations }) => decorations,
    }
)

async function hoverRequest(
    view: EditorView,
    occurrence: Occurrence,
    params: TextDocumentPositionParameters
): Promise<HoverResult> {
    const api = await getOrCreateCodeIntelAPI(view.state.facet(blobPropsFacet).platformContext)
    const hover = await api.getHover(params).toPromise()

    let markdownContents: string =
        hover === null || hover.contents.length === 0
            ? ''
            : hover.contents
                  .map(({ value }) => value)
                  .join('\n\n----\n\n')
                  .trimEnd()
    if (markdownContents === '' && isInteractiveOccurrence(occurrence)) {
        markdownContents = 'No hover information available'
    }
    return { markdownContents, hoverMerged: hover, isPrecise: isPrecise(hover) }
}

function isPrecise(hover: HoverMerged | null): boolean {
    for (const badge of hover?.aggregatedBadges || []) {
        if (badge.text === 'precise') {
            return true
        }
    }
    return false
}

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

/**
 * Field for storing visible hovered occurrence range and visible code-intel tooltip for this occurrence.
 */
const [hoverTooltip, setHoverTooltip] = createUpdateableField<{
    tooltip: Tooltip
    range: { from: number; to: number }
} | null>(null, field => [
    // show code-intel tooltip
    showTooltip.computeN([field], state => [state.field(field)?.tooltip ?? null]),

    // highlight occurrence with tooltip
    EditorView.decorations.compute([field], state => {
        const value = state.field(field)

        if (!value?.tooltip || !value?.range) {
            return Decoration.none
        }

        return Decoration.set(selectionHighlightDecoration.range(value.range.from, value.range.to))
    }),
])

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

                            const currentTooltip = view.state.field(hoverTooltip)
                            if (!currentTooltip) {
                                return true
                            }

                            return !isOffsetInHoverRange(offset, currentTooltip.range)
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
                                const currentTooltip = view.state.field(hoverTooltip)?.tooltip
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

                            return of(preciseWordAtCoords(this.view, position)).pipe(
                                // if the hovered token changed, hide the existing tooltip and proceed with fetching of the new one
                                tap(range => {
                                    const currentRange = view.state.field(hoverTooltip)?.range
                                    if (range && currentRange && !isEqual(currentRange, range)) {
                                        setHoverTooltip(view, null)
                                    }
                                }),
                                switchMap(range =>
                                    range
                                        ? from(getHoverTooltip(view, range.from)).pipe(
                                              catchError(() => of(null)),
                                              map(tooltip => (tooltip ? { tooltip, range } : null))
                                          )
                                        : of(null)
                                ),
                                map(tooltip => ({ position, tooltip }))
                            )
                        })
                    )
                    .subscribe(next => {
                        if (next === 'HIDE') {
                            setHoverTooltip(view, null)
                            return
                        }

                        // We only change the tooltip when
                        // a) There is a new tooltip at the position (tooltip !== null)
                        // b) there is no tooltip and the mouse is moving away from the tooltip
                        if (next.tooltip || next.position.direction !== 'towards') {
                            setHoverTooltip(view, next.tooltip)
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
