import * as sourcegraph from 'sourcegraph'
import { DiagnosticCollection } from '../../types/diagnosticCollection'

/**
 * The diagnostics service publishes diagnostics about resources.
 */
export interface DiagnosticsService {
    /** The diagnostic collection, containing all diagnostics. */
    readonly collection: DiagnosticCollection<sourcegraph.Diagnostic>
}

/**
 * Creates a {@link DiagnosticsService} instance.
 */
export function createDiagnosticsService(): DiagnosticsService {
    const collection = new DiagnosticCollection<sourcegraph.Diagnostic>('')
    return { collection }
}
