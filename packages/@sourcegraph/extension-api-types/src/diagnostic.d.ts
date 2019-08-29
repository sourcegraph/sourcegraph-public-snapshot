import * as sourcegraph from 'sourcegraph'
import { Range, Location } from './location'

/**
 * A diagnostic.
 *
 * @see module:sourcegraph.Diagnostic
 */
export interface Diagnostic
    extends Pick<
        sourcegraph.Diagnostic,
        Exclude<keyof sourcegraph.Diagnostic, 'resource' | 'range' | 'relatedInformation'>
    > {
    /** The resource that the diagnostic applies to. */
    resource: string

    /** The range that the diagnostic applies to. */
    range: Range

    relatedInformation?: DiagnosticRelatedInformation[]
}

export interface DiagnosticRelatedInformation
    extends Pick<
        sourcegraph.DiagnosticRelatedInformation,
        Exclude<keyof sourcegraph.DiagnosticRelatedInformation, 'location'>
    > {
    location: Location
}
