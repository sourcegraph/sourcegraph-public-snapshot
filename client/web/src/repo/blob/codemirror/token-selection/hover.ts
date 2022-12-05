import { countColumn, StateEffect, StateField } from '@codemirror/state'
import { closeHoverTooltips, EditorView, showTooltip, Tooltip } from '@codemirror/view'

import { HoverMerged, TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { Occurrence, Position } from '@sourcegraph/shared/src/codeintel/scip'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { isInteractiveOccurrence, occurrenceAtPosition } from '../occurrence-utils'
import { CodeIntelTooltip, emptyHoverResult, HoverResult } from '../tooltips/CodeIntelTooltip'

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
