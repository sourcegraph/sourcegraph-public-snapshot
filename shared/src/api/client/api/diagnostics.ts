import { ProxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Unsubscribable } from 'rxjs'
import { DiagnosticsService } from '../services/diagnosticsService'
import { fromDiagnosticData, DiagnosticData } from '../../types/diagnostic'

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
        this.diagnosticsService.set(name, data ? fromDiagnosticData(data) : null)
    }

    public unsubscribe(): void {
        // noop
    }
}
