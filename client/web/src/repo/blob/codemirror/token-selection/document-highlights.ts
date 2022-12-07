import { Extension, StateField } from '@codemirror/state'
import { EditorView } from '@codemirror/view'

import { DocumentHighlight } from '@sourcegraph/codeintellify'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { Occurrence } from '@sourcegraph/shared/src/codeintel/scip'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import { blobPropsFacet } from '..'
import { showDocumentHighlights } from '../document-highlights'

export const documentHighlightCache = StateField.define<Map<Occurrence, Promise<DocumentHighlight[]>>>({
    create: () => new Map(),
    update: value => value,
})
const [documentHighlightsField, , setDocumentHighlights] = createUpdateableField<DocumentHighlight[]>([], field =>
    showDocumentHighlights.from(field)
)

async function getDocumentHighlights(view: EditorView, occurrence: Occurrence): Promise<DocumentHighlight[]> {
    const cache = view.state.field(documentHighlightCache)
    const fromCache = cache.get(occurrence)
    if (fromCache) {
        return fromCache
    }
    const blobProps = view.state.facet(blobPropsFacet)
    const api = await blobProps.extensionsController?.extHostAPI
    if (!api) {
        return []
    }
    const result = await api.getDocumentHighlights({
        textDocument: { uri: toURIWithPath(blobProps.blobInfo) },
        position: occurrence.range.start,
    })
    const promise = wrapRemoteObservable(result).toPromise()
    cache.set(occurrence, promise)
    return promise
}
export function showDocumentHighlightsForOccurrence(view: EditorView, occurrence: Occurrence): void {
    getDocumentHighlights(view, occurrence).then(
        result => view.dispatch({ effects: setDocumentHighlights.of(result) }),
        () => {}
    )
}

export function documentHighlightsExtension(): Extension {
    return [documentHighlightCache, documentHighlightsField]
}
