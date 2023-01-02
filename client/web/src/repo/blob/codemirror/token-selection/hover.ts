import { countColumn, Extension, StateEffect, StateField } from '@codemirror/state'
import {
    closeHoverTooltips,
    Decoration,
    EditorView,
    getTooltip,
    PluginValue,
    showTooltip,
    Tooltip,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'
import { BehaviorSubject, from, fromEvent, of, Subject, Subscription } from 'rxjs'
import { catchError, debounceTime, filter, map, scan, switchMap, tap } from 'rxjs/operators'

import { HoverMerged, TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { formatSearchParameters, LineOrPositionOrRange } from '@sourcegraph/common'
import { getOrCreateCodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'
import { Occurrence, Position } from '@sourcegraph/shared/src/codeintel/scip'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { pin } from '../hovercard'
import {
    isInteractiveOccurrence,
    occurrenceAtMouseEvent,
    occurrenceAtPosition,
    rangeToCmSelection,
} from '../occurrence-utils'
import { CodeIntelTooltip, HoverResult } from '../tooltips/CodeIntelTooltip'
import { preciseOffsetAtCoords, uiPositionToOffset } from '../utils'
import { warmupOccurrence } from './selections'

export function hoverExtension(): Extension {
    return [hoverCache, hoveredOccurrenceField, hoverTooltipField, hoverManager, tooltipStyles, hoverField, pinManager]
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

/**
 * Effect for setting the hovercard for the currently hovered token.
 */
const setHoverTooltip = StateEffect.define<CodeIntelTooltip | null>()
const hoverTooltipField = StateField.define<CodeIntelTooltip | null>({
    create() {
        return null
    },

    update(state, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setHoverTooltip)) {
                return effect.value
            }
        }
        return state
    },

    provide(field) {
        return [
            showTooltip.computeN([field], state => [state.field(field)]),

            EditorView.decorations.compute([field], state => {
                const tooltip = state.field(field)
                if (tooltip === null) {
                    return Decoration.none
                }

                const range = rangeToCmSelection(state, tooltip.occurrence.range)

                return Decoration.set(Decoration.mark({ class: 'selection-highlight' }).range(range.from, range.to))
            }),
        ]
    },
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
    return new CodeIntelTooltip(view, occurrence, result)
}

export function hoverAtOccurrence(view: EditorView, occurrence: Occurrence): Promise<HoverResult> {
    const cache = view.state.field(hoverCache)
    const fromCache = cache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const uri = toURIWithPath(view.state.facet(blobPropsFacet).blobInfo)
    const contents = hoverRequest(view, occurrence, {
        position: { line: occurrence.range.start.line, character: occurrence.range.start.character + 1 },
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
        private nextPin: Subject<LineOrPositionOrRange | null>
        private subscription: Subscription

        constructor(view: EditorView) {
            this.nextPin = new BehaviorSubject(view.state.field(pin))
            this.subscription = this.nextPin
                .pipe(
                    map(pin => {
                        if (!pin || !pin.line || !pin.character) {
                            return null
                        }

                        return uiPositionToOffset(view.state.doc, { line: pin.line, character: pin.character })
                    }),
                    switchMap(pos => (pos ? from(getHoverTooltip(view, pos)) : of(null)))
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
 * The MouseEvent uses numbers to indicate which button was pressed.
 * See https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/buttons#value
 */
const MOUSE_NO_BUTTON = 0

const HOVER_DEBOUNCE_TIME = 25 // ms

/**
 * Listens to mousemove events, determines whether the position under the mouse
 * cursor is eligible (whether a "word" is under the mouse cursor), fetches
 * hover information as necessary and updates {@link hoverTooltipField}.
 */
const hoverManager = ViewPlugin.fromClass(
    class HoverManager implements PluginValue {
        private subscription: Subscription

        constructor(private readonly view: EditorView) {
            this.subscription = fromEvent<MouseEvent>(this.view.dom, 'mousemove')
                .pipe(
                    tap(event => {
                        const atEvent = occurrenceAtMouseEvent(view, event)
                        const hoveredOccurrence = atEvent ? atEvent.occurrence : null
                        if (hoveredOccurrence && isInteractiveOccurrence(hoveredOccurrence)) {
                            warmupOccurrence(view, hoveredOccurrence)
                        }
                    }),

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
                        const tooltip = view.state.field(hoverTooltipField)
                        if (!tooltip) {
                            return true
                        }
                        const wordWithTooltip = view.state.wordAt(tooltip.pos)
                        if (!wordWithTooltip) {
                            return true
                        }
                        return !isOffsetInHoverRange(offset, wordWithTooltip)
                    }),

                    // Determine the precise location of the word under the cursor.
                    switchMap(position => {
                        // Hide any hovercard when
                        // - the mouse is over an element that is not part of
                        //   the content. This seems necessary to make hovercards
                        //   not appear and hide open hovercards when the mouse
                        //   moves over the editor's search panel.
                        // - the user starts to select text
                        if (
                            position.buttons !== MOUSE_NO_BUTTON ||
                            !position.target ||
                            !this.view.contentDOM.contains(position.target as Node)
                        ) {
                            return of(null)
                        }

                        return of(position).pipe(
                            map(position => preciseOffsetAtCoords(view, position)),
                            switchMap(pos =>
                                typeof pos === 'number'
                                    ? from(getHoverTooltip(view, pos)).pipe(catchError(() => of(null)))
                                    : of(null)
                            )
                        )
                    })
                )
                .subscribe(value => {
                    console.log({ value })
                    view.dispatch({ effects: setHoverTooltip.of(value) })
                })
        }

        public destroy(): void {
            this.subscription.unsubscribe()
        }
    }
)

function isOffsetInHoverRange(offset: number, range: { from: number; to: number }): boolean {
    return range.from <= offset && offset <= range.to
}
