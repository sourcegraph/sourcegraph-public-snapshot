import { Extension, StateField } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { DocumentHighlight } from '@sourcegraph/codeintellify'
import { getOrCreateCodeIntelAPI } from '@sourcegraph/shared/src/codeintel/api'
import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'

const documentHighlightCache = StateField.define<Map<Occurrence, Promise<DocumentHighlight[]>>>({
    create: () => new Map(),
    update: value => value,
})
export const [documentHighlightsField, , setDocumentHighlights] = createUpdateableField<DocumentHighlight[]>([])

async function getDocumentHighlights(view: EditorView, occurrence: Occurrence): Promise<DocumentHighlight[]> {
    const cache = view.state.field(documentHighlightCache)
    const fromCache = cache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const blobProps = view.state.facet(blobPropsFacet)

    const api = await getOrCreateCodeIntelAPI(blobProps.platformContext)
    const promise = api
        .getDocumentHighlights({
            textDocument: { uri: toURIWithPath(blobProps.blobInfo) },
            position: occurrence.range.start,
        })
        .toPromise()
    cache.set(occurrence, promise)
    return promise
}
export function showDocumentHighlightsForOccurrence(view: EditorView, occurrence: Occurrence): void {
    getDocumentHighlights(view, occurrence).then(
        result => view.dispatch({ effects: setDocumentHighlights.of(result) }),
        () => {}
    )
}

export function findByOccurrence(
    highlights: DocumentHighlight[],
    occurrence: Occurrence
): DocumentHighlight | undefined {
    return highlights.find(
        highlight =>
            occurrence.range.start.line === highlight.range.start.line &&
            occurrence.range.start.character === highlight.range.start.character &&
            occurrence.range.end.line === highlight.range.end.line &&
            occurrence.range.end.character === highlight.range.end.character
    )
}

export function documentHighlightsExtension(): Extension {
    return [documentHighlightCache, documentHighlightsField]
}
