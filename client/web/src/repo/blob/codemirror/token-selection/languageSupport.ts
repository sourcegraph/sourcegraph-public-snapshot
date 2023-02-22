import { Facet } from '@codemirror/state'

import { findLanguageMatchingDocument } from '@sourcegraph/shared/src/codeintel/api'
import { toURIWithPath } from '@sourcegraph/shared/src/util/url'

import type { BlobInfo } from '../../CodeMirrorBlob'

/**
 * Facet providing information whether the current blob language has code intelligence support.
 */
export const languageSupport = Facet.define<BlobInfo, boolean>({
    static: true,
    combine: blobInfos => (blobInfos[0] ? !!findLanguageMatchingDocument({ uri: toURIWithPath(blobInfos[0]) }) : false),
})
