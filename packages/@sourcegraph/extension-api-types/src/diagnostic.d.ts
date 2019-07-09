import * as sourcegraph from 'sourcegraph'
import { Range } from './location'

/**
 * A diagnostic.
 *
 * @see module:sourcegraph.Diagnostic
 */
export interface Diagnostic
    extends Pick<sourcegraph.Diagnostic, Exclude<keyof sourcegraph.Diagnostic, 'resource' | 'range'>> {
    /** The resource that the diagnostic applies to. */
    resource: string

    /** The range that the diagnostic applies to. */
    range: Range
}
