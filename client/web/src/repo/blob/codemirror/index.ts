import { Facet } from '@codemirror/state'

import { BlobProps } from '../Blob'

/**
 * This facet is a necessary evil to allow access to props passed to the blob
 * component from React components rendered by extensions.
 */
export const blobPropsFacet = Facet.define<BlobProps, BlobProps>({
    combine: props => props[0],
})
