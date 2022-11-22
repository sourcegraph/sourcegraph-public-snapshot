import { Facet } from '@codemirror/state'
import { Remote } from 'comlink'
import * as H from 'history'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { Range } from '@sourcegraph/shared/src/codeintel/scip'

import { BlobInfo } from '../Blob'

export const textDocumentImplemenationSupport = Facet.define<boolean, boolean>({
    combine: props => props[0],
})
export const uriFacet = Facet.define<string, string>({
    combine: props => props[0],
})
export const selectionsFacet = Facet.define<Map<string, Range>, Map<string, Range>>({
    combine: props => props[0],
})
export const blobInfoFacet = Facet.define<BlobInfo, BlobInfo>({
    combine: props => props[0],
})
export const codeintelFacet = Facet.define<Remote<FlatExtensionHostAPI>, Remote<FlatExtensionHostAPI>>({
    combine: props => props[0],
})
export const historyFacet = Facet.define<H.History, H.History>({
    combine: props => props[0],
})
