import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import { Unsubscribable } from 'rxjs'
import { toDiagnostic } from '../../extension/api/types'
import { DiagnosticsService } from '../services/diagnosticsService'

/** The format for sending {@link Diagnostic} data between the client and extension host. */
export type DiagnosticData = [string, Diagnostic[]][]

/** @internal */
export interface ClientDiagnosticsAPI extends ProxyValue {
    // TODO!(sqs): inefficient
    $acceptDiagnosticData(updates: DiagnosticData): void
}

/** @internal */
export class ClientDiagnostics implements ClientDiagnosticsAPI, Unsubscribable {
    public readonly [proxyValueSymbol] = true

    constructor(private diagnosticsService: Pick<DiagnosticsService, 'collection'>) {}

    public $acceptDiagnosticData(data: DiagnosticData): void {
        this.diagnosticsService.collection.set(data.map(([uri, diagnostics]) => [uri, diagnostics.map(toDiagnostic)]))
    }

    public unsubscribe(): void {
        // noop
    }
}
