import * as sourcegraph from '@sourcegraph/extension-api-types'
import { countColumn, StateEffect, StateField } from '@codemirror/state'
import { EditorView, getTooltip, showTooltip, Tooltip, TooltipView } from '@codemirror/view'
import { Remote } from 'comlink'

import { HoverMerged, TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { Occurrence, Position } from '@sourcegraph/shared/src/codeintel/scip'
import { BlobViewState, toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { isInteractiveOccurrence } from './tokens-as-links'
import { HovercardView, HoverData } from './hovercard'
import { of } from 'rxjs'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { occurrenceAtPosition, rangeToSelection } from './positions'
import { blobInfoFacet, codeintelFacet, textDocumentImplemenationSupport } from './textdocument-facets'

export const hoverField = StateField.define<Tooltip | null>({
    create: () => null,
    update(tooltip, transactions) {
        for (const effect of transactions.effects) {
            if (effect.is(setHoverEffect)) {
                tooltip = effect.value
            }
        }
        return tooltip
    },
    provide: field => showTooltip.from(field),
})
export const setHoverEffect = StateEffect.define<Tooltip | null>()
export const hoverCache = StateField.define<Map<Occurrence, Promise<HoverResult>>>({
    create: () => new Map(),
    update: value => value,
})

export function showTemporaryTooltip(
    view: EditorView,
    message: string,
    position: sourcegraph.Position,
    clearTimeout: number
): void {
    const line = view.state.doc.line(position.line + 1)
    const tooltip: Tooltip = {
        pos: line.from + position.character + 1,
        above: true,
        create() {
            const div = document.createElement('div')
            div.textContent = message
            return {
                dom: div,
            }
        },
    }
    view.dispatch({ effects: setHoverEffect.of(tooltip) })
    setTimeout(() => {
        const tooltipView = getTooltip(view, tooltip)
        if (tooltipView) {
            view.dispatch({ effects: setHoverEffect.of(null) })
        }
    }, clearTimeout)
}

export async function getHoverTooltip(view: EditorView, pos: number): Promise<Tooltip | null> {
    const cmLine = view.state.doc.lineAt(pos)
    const line = cmLine.number - 1
    const character = countColumn(cmLine.text, 1, pos - cmLine.from) - 1
    const atEvent = occurrenceAtPosition(view.state, new Position(line, character))
    if (!atEvent) {
        return null
    }
    const result = await hoverAtOccurrence(view, atEvent.occurrence)
    if (!result.contents) {
        return null
    }
    return markdownTooltip(view, atEvent.occurrence, result)
}

export function hoverAtOccurrence(view: EditorView, occurrence: Occurrence): Promise<HoverResult> {
    const cache = view.state.field(hoverCache)
    const fromCache = cache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const uri = toURIWithPath(view.state.facet(blobInfoFacet))
    const contents = hoverRequest(view.state.facet(codeintelFacet), occurrence, {
        position: { line: occurrence.range.start.line, character: occurrence.range.start.character + 1 },
        textDocument: { uri },
    })
    cache.set(occurrence, contents)
    return contents
}

interface HoverResult {
    contents: string
    hover?: HoverMerged | null
}
async function hoverRequest(
    codeintel: Remote<FlatExtensionHostAPI>,
    occurrence: Occurrence,
    params: TextDocumentPositionParameters
): Promise<HoverResult> {
    const hover = await codeintel.getHover(params)
    const result = await wrapRemoteObservable(hover).toPromise()
    let contents =
        result === undefined || result.isLoading || result.result === null || result.result.contents.length === 0
            ? ''
            : result.result.contents
                  .map(({ value }) => value)
                  .join('\n\n----\n\n')
                  .trimEnd()
    if (contents === '' && isInteractiveOccurrence(occurrence)) {
        contents = 'No hover information available'
    }
    return { contents, hover: result?.result }
}

export function isPrecise(hover?: HoverResult): boolean {
    for (const badge of hover?.hover?.aggregatedBadges || []) {
        if (badge.text === 'precise') {
            return true
        }
    }
    return false
}

export function markdownTooltip(view: EditorView, occurrence: Occurrence, hover: HoverResult): Tooltip {
    const { contents } = hover
    const range = rangeToSelection(view.state, occurrence.range)
    return {
        pos: range.from,
        end: range.to,
        above: true,
        create(): TooltipView {
            const blobInfo = view.state.facet(blobInfoFacet)
            const referencesURL = toPrettyBlobURL({
                ...blobInfo,
                range: occurrence.range.asOneBased(),
                viewState: 'references',
            })
            const actions: ActionItemAction[] = [
                {
                    active: true,
                    action: {
                        id: 'findReferences',
                        title: 'Find references',
                        command: 'open',
                        commandArguments: [referencesURL],
                    },
                },
            ]
            if (isPrecise(hover) && view.state.facet(textDocumentImplemenationSupport)) {
                const implementationsURL = toPrettyBlobURL({
                    ...blobInfo,
                    range: occurrence.range.asOneBased(),
                    viewState: `implementations_${blobInfo.mode}` as BlobViewState,
                })
                actions.push({
                    active: true,
                    action: {
                        id: 'findImplementations',
                        title: 'Find implementations',
                        command: 'open',
                        commandArguments: [implementationsURL],
                    },
                })
            }
            const data: HoverData = {
                actionsOrError: actions,
                hoverOrError: {
                    range: occurrence.range,
                    aggregatedBadges: hover.hover?.aggregatedBadges,
                    contents: [
                        {
                            value: contents,
                            kind: MarkupKind.Markdown,
                        },
                    ],
                },
            }
            const hovercard = new HovercardView(view, occurrence.range.asOneBased(), false, of(data))
            return hovercard
        },
    }
}
