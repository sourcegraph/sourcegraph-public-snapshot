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
    $acceptDiagnosticCollection(name: string, data: DiagnosticData | null): void
}

/** @internal */
export class ClientDiagnostics implements ClientDiagnosticsAPI, Unsubscribable {
    public readonly [proxyValueSymbol] = true

    constructor(private diagnosticsService: Pick<DiagnosticsService, 'set'>) {}

    public $acceptDiagnosticCollection(name: string, data: DiagnosticData | null): void {
        this.diagnosticsService.set(
            name,
            data ? data.map(([uri, diagnostics]) => [new URL(uri), diagnostics.map(toDiagnostic)]) : null
        )
    }

    public unsubscribe(): void {
        // noop
    }
}
