import type { Hover as APIHover } from 'sourcegraph'

import type { Range } from './location'

/**
 * A hover message.
 *
 * @see module:sourcegraph.Hover
 */
export interface Hover extends Pick<APIHover, 'contents'> {
    /** The range that the hover applies to. */
    readonly range?: Range
}
