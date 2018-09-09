import { Range } from './range'

/**
 * A selection in a document. A selection is a range with direction.
 */
export interface Selection extends Readonly<Range> {
    /**
     * Whether the selection is reversed. In a reversed selection, the cursor is at the start of the range (instead
     * of the end of the range).
     */
    readonly isReversed: boolean
}
