import type { TextDocumentDecoration as APITextDocumentDecoration } from 'sourcegraph'

import { Range } from './location'

/**
 * A decoration to apply to a text document.
 *
 * @see module:sourcgraph.TextDocumentDecoration
 */
export interface TextDocumentDecoration extends Omit<APITextDocumentDecoration, 'range'> {
    /** The range that the decoration applies to. */
    range: Range
}
