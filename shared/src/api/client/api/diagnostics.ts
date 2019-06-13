import { ProxyValue, proxyValueSymbol, ProxyResult, proxyValue } from '@sourcegraph/comlink'
import { Diagnostic } from '@sourcegraph/extension-api-types'
import { Unsubscribable, Subscription } from 'rxjs'
import { DiagnosticService } from '../services/diagnosticService'
import { ProxySubscribable } from '../../extension/api/common'
import { wrapRemoteObservable } from './common'
import { map } from 'rxjs/operators'
import { toDiagnostic } from '../../types/diagnostic'
import { ContextValues } from 'sourcegraph'

/** @internal */
export interface ClientDiagnosticsAPI extends ProxyValue {
    $registerDiagnosticProvider(
        name: string,
        providerFunction: ProxyResult<
            ((scope: {}, context: ContextValues) => ProxySubscribable<Diagnostic[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue
}

/** @internal */
export const createClientDiagnostics = (
    diagnosticService: Pick<DiagnosticService, 'registerDiagnosticProvider'>
): ClientDiagnosticsAPI & Unsubscribable => {
    const subscriptions = new Subscription()
    return {
        [proxyValueSymbol]: true,
        $registerDiagnosticProvider: (
            type: string,
            providerFunction: ProxyResult<
                ((scope: {}, context: ContextValues) => ProxySubscribable<Diagnostic[]>) & ProxyValue
            >
        ): Unsubscribable & ProxyValue => {
            const subscription = diagnosticService.registerDiagnosticProvider(type, {
                provideDiagnostics: (scope, context) =>
                    wrapRemoteObservable(providerFunction(scope, context)).pipe(
                        map(diagnostics => diagnostics.map(d => toDiagnostic(d)))
                    ),
            })
            return proxyValue(subscription)
        },
        unsubscribe: () => subscriptions.unsubscribe(),
    }
}
