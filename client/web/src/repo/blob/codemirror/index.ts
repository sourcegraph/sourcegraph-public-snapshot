import { Extension, Facet, StateField } from '@codemirror/state'
import { LineOrPositionOrRange } from '@sourcegraph/common'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { UIPositionSpec } from '@sourcegraph/shared/src/util/url'
import { BlobProps } from '../Blob'
import { hovercardRanges } from './hovercard'
import { offsetToUIPosition, uiPositionToOffset } from './utils'

/**
 * This facet is a necessary evil to allow access to props passed to the blob
 * component from React components rendered by extensions.
 */
export const blobPropsFacet = Facet.define<BlobProps, BlobProps>({
    combine: props => {
        return props[0]
    },
})
