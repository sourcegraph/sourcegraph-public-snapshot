import { StateField } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { TextDocumentPositionParameters } from '@sourcegraph/client-api'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { Occurrence, Position, Range } from '@sourcegraph/shared/src/codeintel/scip'
import { parseRepoURI, toPrettyBlobURL, toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { Location } from '@sourcegraph/extension-api-types'
import { occurrenceAtEvent } from './positions'
import { blobInfoFacet, codeintelFacet, historyFacet, selectionsFacet } from './textdocument-facets'
import { selectRange } from './textdocument-selections'
import { showTemporaryTooltip } from './textdocument-hover'

export function goToDefinitionAtEvent(view: EditorView, event: MouseEvent): Promise<() => void> {
    const atEvent = occurrenceAtEvent(view, event)
    if (!atEvent) {
        return Promise.resolve(() => {})
    }
    const { occurrence, position } = atEvent
    return goToDefinitionAtOccurrence(view, position, occurrence)
}

export function goToDefinitionAtOccurrence(
    view: EditorView,
    position: Position,
    occurrence: Occurrence
): Promise<() => void> {
    const cache = view.state.field(definitionCache)
    const fromCache = cache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const uri = toURIWithPath(view.state.facet(blobInfoFacet))
    const promise = goToDefinition(view, { position, textDocument: { uri } })
    cache.set(occurrence, promise)
    return promise
}

export const definitionCache = StateField.define<Map<Occurrence, Promise<() => void>>>({
    create: () => new Map(),
    update: value => value,
})

export async function goToDefinition(view: EditorView, params: TextDocumentPositionParameters): Promise<() => void> {
    const codeintel = view.state.facet(codeintelFacet)
    const definition = await codeintel.getDefinition(params)

    const result = await wrapRemoteObservable(definition).toPromise()
    if (result.isLoading) {
        return () => {}
    }
    if (result.result.length === 0) {
        return () => showTemporaryTooltip(view, 'No definition found', params.position, 2000)
    }
    for (const location of result.result) {
        if (location.uri === params.textDocument.uri && location.range && location.range) {
            const requestPosition = new Position(params.position.line, params.position.character)
            const {
                start: { line: startLine, character: startCharacter },
                end: { line: endLine, character: endCharacter },
            } = location.range
            const resultRange = Range.fromNumbers(startLine, startCharacter, endLine, endCharacter)
            if (resultRange.contains(requestPosition)) {
                return () => showTemporaryTooltip(view, 'You are at the definition', params.position, 2000)
            }
        }
    }
    if (result.result.length === 1) {
        const destination = result.result[0]
        const hrefTo = locationToURL(destination)
        const { range, uri } = result.result[0]
        if (hrefTo && range) {
            return () => {
                const history = view.state.facet(historyFacet)
                const selectionRange = Range.fromNumbers(
                    range.start.line,
                    range.start.character,
                    range.end.line,
                    range.end.character
                )
                const source: Location = {
                    range: { start: params.position, end: params.position },
                    uri: params.textDocument.uri,
                }
                const hrefFrom = locationToURL(source)
                if (hrefFrom) {
                    history.push(hrefFrom)
                }
                if (uri === params.textDocument.uri) {
                    selectRange(view, selectionRange)
                } else {
                    const selections = view.state.facet(selectionsFacet)
                    selections.set(uri, selectionRange)
                }
                history.push(hrefTo)
            }
        }
    }
    return () => showTemporaryTooltip(view, 'FIXME: Multiple definitions', params.position, 2000)
}

function locationToURL(location: Location): string | undefined {
    const { range, uri } = location
    const { filePath, repoName, revision } = parseRepoURI(uri)
    if (filePath && range) {
        return toPrettyBlobURL({
            repoName,
            revision,
            filePath,
            position: { line: range.start.line + 1, character: range.start.character + 1 },
        })
    }
    return undefined
}
