import * as sourcegraph from 'sourcegraph'
import { Range } from './location'

/**
 * A text edit represents edits to apply to a document.
 *
 * @see module:sourcegraph.TextEdit
 */
export interface TextEdit extends Pick<sourcegraph.TextEdit, 'newText'> {
    /**
     * The range this edit applies to.
     */
    readonly range: Range
}
