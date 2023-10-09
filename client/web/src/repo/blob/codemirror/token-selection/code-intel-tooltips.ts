import {
    countColumn,
    EditorSelection,
    type Extension,
    Prec,
    StateEffect,
    StateField,
    type TransactionSpec,
} from '@codemirror/state'
import {
    EditorView,
    getTooltip,
    type PluginValue,
    showTooltip,
    type Tooltip,
    ViewPlugin,
    type ViewUpdate,
} from '@codemirror/view'
import { BehaviorSubject, from, fromEvent, of, type Subject, Subscription } from 'rxjs'
import { debounceTime, filter, map, scan, switchMap, tap } from 'rxjs/operators'

import type { HoverMerged, TextDocumentPositionParameters } from '@sourcegraph/client-api/src'
import { formatSearchParameters, type LineOrPositionOrRange } from '@sourcegraph/common/src'
import { type Occurrence, Position } from '@sourcegraph/shared/src/codeintel/scip'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { getCodeIntelAPI } from '../codeintel'
import { blobPropsFacet } from '../index'
import {
    isInteractiveOccurrence,
    occurrenceAtMouseEvent,
    occurrenceAtPosition,
    positionAtCmPosition,
    rangeToCmSelection,
} from '../occurrence-utils'
import { BLOB_SEARCH_CONTAINER_ID } from '../search'
import { CodeIntelTooltip, type HoverResult } from '../tooltips/CodeIntelTooltip'
import { positionToOffset, preciseOffsetAtCoords, uiPositionToOffset } from '../utils'

import { preloadDefinition } from './definition'
import { showDocumentHighlightsForOccurrence } from './document-highlights'
import { languageSupport } from './languageSupport'

type CodeIntelTooltipTrigger = 'focus' | 'hover' | 'pin'
type CodeIntelTooltipState = { occurrence: Occurrence; tooltip: Tooltip | null } | null

export const setFocusedOccurrence = StateEffect.define<Occurrence | null>()
export const setFocusedOccurrenceTooltip = StateEffect.define<Tooltip | null>()
const setPinnedCodeIntelTooltipState = StateEffect.define<CodeIntelTooltipState>()
const setHoveredCodeIntelTooltipState = StateEffect.define<CodeIntelTooltipState>()

/**
 * {@link StateField} storing focused (selected), hovered and pinned {@link Occurrence}s and {@link Tooltip}s associate with them.
 */
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
                return {
                    ...value,
                    focus: value.focus?.occurrence ? { ...value.focus, tooltip: effect.value } : null,
                }
            }

            if (effect.is(setHoveredCodeIntelTooltipState)) {
                return { ...value, hover: effect.value }
            }

            if (effect.is(setPinnedCodeIntelTooltipState)) {
                // If the pinned occurrence is the same as the hovered or focused one, use pin the existing one
                for (const trigger of ['hover', 'focus'] as const) {
                    if (effect.value?.occurrence === value[trigger]?.occurrence) {
                        return { ...value, pin: value[trigger], [trigger]: null }
                    }
                }

                return { ...value, pin: effect.value }
            }
        }

        return value
    },
    provide(field) {
        return [
            showTooltip.computeN([field], state => {
                const { hover, focus, pin } = state.field(field)
                const isLanguageSupported = state.facet(languageSupport)

                // Only show one tooltip for the occurrence at a time
                const uniqueTooltips = [pin, focus, hover]
                    .reduce((acc, current) => {
                        if (current?.tooltip && acc.every(({ occurrence }) => occurrence !== current.occurrence)) {
                            acc.push(current)
                        }
                        return acc
                    }, [] as NonNullable<CodeIntelTooltipState>[])
                    .map(({ tooltip }) => tooltip)
                    .filter(tooltip =>
                        tooltip instanceof CodeIntelTooltip ? (isLanguageSupported ? tooltip : null) : tooltip
                    )

                return uniqueTooltips
            }),

            /**
             * If there is a focused occurrence set editor's tabindex to -1, so that pressing Shift+Tab moves the focus
             * outside the editor instead of focusing the editor itself.
             *
             * Explicitly define extension precedence to override the [default tabindex value](https://sourcegraph.com/github.com/sourcegraph/sourcegraph@728ea45d1cc063cd60cbd552e00929c09cb8ced8/-/blob/client/web/src/repo/blob/CodeMirrorBlob.tsx?L47&).
             */
            Prec.high(
                EditorView.contentAttributes.compute([field], state => ({
                    tabindex: state.field(field).focus?.occurrence ? '-1' : '0',
                }))
            ),
        ]
    },
})

export const getCodeIntelTooltipState = (
    view: EditorView,
    key: CodeIntelTooltipTrigger
): { occurrence: Occurrence; tooltip: Tooltip | null } | null => view.state.field(codeIntelTooltipsState)[key]

const focusOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    const offset = positionToOffset(view.state.doc, occurrence.range.end)
    if (offset !== null) {
        const node = view.domAtPos(offset).node
        const element = node instanceof HTMLElement ? node : node.parentElement
        const lineEl = element?.matches('.cm-line') ? element : element?.closest('.cm-line')
        const interactiveOccurrenceEl = lineEl?.querySelector<HTMLElement>('.interactive-occurrence')
        if (interactiveOccurrenceEl) {
            interactiveOccurrenceEl.focus()
        }
    }
}

const preloadHoverData = (view: EditorView, occurrence: Occurrence): void => {
    if (!view.state.field(hoverCache).has(occurrence)) {
        hoverAtOccurrence(view, occurrence).then(
            () => {},
            () => {}
        )
    }
}

const warmupOccurrence = (view: EditorView, occurrence: Occurrence): void => {
    preloadHoverData(view, occurrence)
    preloadDefinition(view, occurrence)
}

/**
 * Sets given occurrence to {@link codeIntelTooltipsState}, sets editor selection to the occurrence range start,
 * fetches hover, definition data and document highlights for occurrence, and focuses the selected occurrence DOM node.
 */
export const selectOccurrence = (view: EditorView, occurrence: Occurrence, isClickEvent?: boolean): void => {
    warmupOccurrence(view, occurrence)
    const spec: TransactionSpec = { effects: setFocusedOccurrence.of(occurrence) }
    if (!isClickEvent) {
        /**
         * Set editor selection cursor to the occurrence start.
         * Ignore click events, they update editor selection by default.
         */
        const selection = rangeToCmSelection(view.state, occurrence.range)
        spec.selection = EditorSelection.cursor(selection.from)
    }
    view.dispatch(spec)
    showDocumentHighlightsForOccurrence(view, occurrence)
    focusOccurrence(view, occurrence)
}

const hoverCache = StateField.define<Map<Occurrence, Promise<HoverResult>>>({
    create: () => new Map(),
    update: value => value,
})

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

async function hoverRequest(
    view: EditorView,
    occurrence: Occurrence,
    params: TextDocumentPositionParameters
): Promise<HoverResult> {
    const api = getCodeIntelAPI(view.state)
    const hover = await api.getHover(params)

    let markdownContents: string =
        hover === null || hover === undefined || hover.contents.length === 0
            ? ''
            : hover.contents
                  .map(({ value }) => value)
                  .join('\n\n----\n\n')
                  .trimEnd()
    const precise = isPrecise(hover)
    if (!precise && markdownContents.length > 0 && !isInteractiveOccurrence(occurrence)) {
        // Don't show search-based results for non-interactive tokens. For example, we don't
        // want to show results for keyword tokens or string literals.
        return {
            isPrecise: false,
            markdownContents: '',
            hoverMerged: null,
        }
    }
    if (markdownContents === '' && isInteractiveOccurrence(occurrence)) {
        markdownContents = 'No hover information available'
    }
    return { markdownContents, hoverMerged: hover, isPrecise: precise }
}

function isPrecise(hover: HoverMerged | null | undefined): boolean {
    for (const badge of hover?.aggregatedBadges || []) {
        if (badge.text === 'precise') {
            return true
        }
    }
    return false
}

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
                                tap(occurrence => {
                                    if (occurrence) {
                                        preloadDefinition(view, occurrence)
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

                                    return from(getHoverTooltip(view, offset)).pipe(
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

/**
 * This field is used by the blob component to sync the position from the URL to
 * the editor.
 */
export const [pin, updatePin] = createUpdateableField<LineOrPositionOrRange | null>(null)

const getPinnedOccurrence = (view: EditorView, pin: LineOrPositionOrRange | null): Occurrence | null => {
    if (!pin?.line || !pin?.character) {
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
                const { navigate, location } = update.state.facet(blobPropsFacet)
                const params = new URLSearchParams(location.search)
                params.delete('popover')
                window.requestAnimationFrame(() =>
                    // Use `navigate(to)` instead of `navigate(to, { replace: true })` in case
                    // the user accidentally clicked somewhere without intending to
                    // dismiss the popover.
                    navigate({ search: formatSearchParameters(params) })
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

export function codeIntelTooltipsExtension(): Extension {
    return [
        codeIntelTooltipsState,
        hoverCache,
        hoverManager,
        pinManager,
        tooltipStyles,

        ViewPlugin.define(view => ({
            update(update: ViewUpdate) {
                if (update.selectionSet) {
                    /**
                     * Selection change may result in the focused occurrence being outside the viewport
                     * (e.g. selecting text from current position to the end of the document).
                     * When focused occurrence is outside the viewport, it is removed from the DOM and editor loses focus.
                     * Ensure the editor remains focused when this happens for keyboard navigation to work.
                     * Ignore cases when viewport change is caused by navigating to next/previous search result
                     * (e.g., by clicking 'Enter' when the search input field is focused).
                     */
                    view.requestMeasure({
                        read(view: EditorView) {
                            if (
                                !view.contentDOM.contains(document.activeElement) &&
                                !document.activeElement?.closest(`#${BLOB_SEARCH_CONTAINER_ID}`)
                            ) {
                                view.contentDOM.focus()
                            }
                        },
                    })
                }
            },
        })),

        EditorView.domEventHandlers({
            click(event, view) {
                // Close selected (focused) code-intel tooltip on click outside.
                const atEvent = occurrenceAtMouseEvent(view, event)
                const current = getCodeIntelTooltipState(view, 'focus')
                if (atEvent?.occurrence !== current?.occurrence && current?.tooltip instanceof CodeIntelTooltip) {
                    view.dispatch({ effects: setFocusedOccurrenceTooltip.of(null) })
                }
            },
        }),
    ]
}
