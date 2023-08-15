import { Facet } from '@codemirror/state'

import type { BlobPropsFacet } from '../CodeMirrorBlob'

/**
 * This facet is a necessary evil to allow access to props passed to the blob
 * component from React components rendered by extensions.
 */
export const blobPropsFacet = Facet.define<BlobPropsFacet, BlobPropsFacet>({
    combine: props => props[0],
})
