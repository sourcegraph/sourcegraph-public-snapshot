import { Facet } from '@codemirror/state'
import { Remote } from 'comlink'
import * as H from 'history'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'

import { BlobInfo } from '../../Blob'

export const textDocumentImplemenationSupport = Facet.define<boolean, boolean>({
    combine: props => props[0],
})
export const uriFacet = Facet.define<string, string>({
    combine: props => props[0],
})
export const blobInfoFacet = Facet.define<BlobInfo, BlobInfo>({
    combine: props => props[0],
})
export const codeintelFacet = Facet.define<Remote<FlatExtensionHostAPI>, Remote<FlatExtensionHostAPI>>({
    combine: props => props[0],
})

interface HistoryState {
    // The destination URL if we trigger `history.goBack()`.  We use this state
    // to avoid inserting redundant 'A->B->A->B' entries when the user triggers
    // "go to definition" twice in a row from the same location.
    previousURL?: string
}
export const historyFacet = Facet.define<H.History, H.History<HistoryState>>({
    combine: props => props[0] as H.History<HistoryState>,
})
