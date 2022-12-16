import { countColumn, Extension, StateEffect, StateField } from '@codemirror/state'
import {
    closeHoverTooltips,
    EditorView,
    hoverTooltip,
    PluginValue,
    showTooltip,
    Tooltip,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'
import { BehaviorSubject, from, of, Subject, Subscription } from 'rxjs'
import { map, switchMap } from 'rxjs/operators'

import { HoverMerged, TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { formatSearchParameters, LineOrPositionOrRange } from '@sourcegraph/common'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { Occurrence, Position } from '@sourcegraph/shared/src/codeintel/scip'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { pin } from '../hovercard'
import { isInteractiveOccurrence, occurrenceAtPosition } from '../occurrence-utils'
import { CodeIntelTooltip, emptyHoverResult, HoverResult } from '../tooltips/CodeIntelTooltip'
import { uiPositionToOffset } from '../utils'

export function hoverExtension(): Extension {
    return [
        hoverCache,
        hoveredOccurrenceField,
        hoverTooltip((view, position) => getHoverTooltip(view, position), { hoverTime: 200, hideOnChange: true }),
        hoverField,
        pinManager,
    ]
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

export async function getHoverTooltip(view: EditorView, pos: number): Promise<Tooltip | null> {
    const cmLine = view.state.doc.lineAt(pos)
    const line = cmLine.number - 1
    const character = countColumn(cmLine.text, 1, pos - cmLine.from) - 1
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
    const api = await view.state.facet(blobPropsFacet).extensionsController?.extHostAPI
    if (!api) {
        return emptyHoverResult
    }
    const hover = await api.getHover(params)
    const result = await wrapRemoteObservable(hover).toPromise()
    let markdownContents =
        result === undefined || result.isLoading || result.result === null || result.result.contents.length === 0
            ? ''
            : result.result.contents
                  .map(({ value }) => value)
                  .join('\n\n----\n\n')
                  .trimEnd()
    if (markdownContents === '' && isInteractiveOccurrence(occurrence)) {
        markdownContents = 'No hover information available'
    }
    return { markdownContents, hoverMerged: result?.result, isPrecise: isPrecise(result?.result) }
}

function isPrecise(hover: HoverMerged | null | undefined): boolean {
    for (const badge of hover?.aggregatedBadges || []) {
        if (badge.text === 'precise') {
            return true
        }
    }
    return false
}
