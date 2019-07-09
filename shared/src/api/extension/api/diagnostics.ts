import { ProxyResult, ProxyValue, proxyValueSymbol, proxyValue, ProxyInput } from '@sourcegraph/comlink'
import { from, Unsubscribable } from 'rxjs'
import { map } from 'rxjs/operators'
import * as sourcegraph from 'sourcegraph'
import { ClientDiagnosticsAPI } from '../../client/api/diagnostics'
import { DiagnosticCollection } from '../../types/diagnosticCollection'
import { toDiagnosticData, fromDiagnosticData, DiagnosticData } from '../../types/diagnostic'
import { syncSubscription } from '../../util'
import { toProxyableSubscribable } from './common'

/** @internal */
export class ExtDiagnostics
    implements Pick<typeof sourcegraph.workspace, 'registerDiagnosticProvider'>, Unsubscribable {
    public readonly [proxyValueSymbol] = true

    constructor(private proxy: ProxyResult<ClientDiagnosticsAPI>) {}

    public registerDiagnosticProvider(
        name: string,
        provider: sourcegraph.DiagnosticProvider
    ): sourcegraph.Unsubscribable {
        const providerFunction: ProxyInput<
            Parameters<ClientDiagnosticsAPI['$registerDiagnosticProvider']>[1]
        > = proxyValue(async scope =>
            toProxyableSubscribable(
                provider.provideDefinition(await this.documents.getSync(textDocument.uri), toPosition(position)),
                toLocations
            )
        )
        return syncSubscription(this.proxy.$registerDiagnosticProvider(selector, providerFunction))
    }
}
