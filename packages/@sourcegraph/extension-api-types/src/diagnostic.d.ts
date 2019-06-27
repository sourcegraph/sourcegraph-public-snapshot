import * as sourcegraph from 'sourcegraph'
import { Range } from './location'

/**
 * A diagnostic.
 *
 * @see module:sourcegraph.Diagnostic
 */
export interface Diagnostic extends Pick<sourcegraph.Diagnostic, Exclude<keyof sourcegraph.Diagnostic, 'range'>> {
    /** The range that the diagnostic applies to. */
    range: Range
}
