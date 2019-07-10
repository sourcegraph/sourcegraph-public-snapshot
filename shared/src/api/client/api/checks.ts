import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { Unsubscribable } from 'rxjs'
import * as sourcegraph from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { CheckService } from '../services/checkService'
import { wrapRemoteObservable } from './common'

export interface ProxiedCheckProvider extends ProxyValue {
    information: ProxyResult<ProxySubscribable<sourcegraph.CheckInformation>>
    provideDiagnosticGroups: ProxyResult<(() => ProxySubscribable<sourcegraph.DiagnosticGroup[]>) & ProxyValue>
}

export interface ClientChecksAPI extends ProxyValue {
    $registerCheckProvider(
        type: Parameters<typeof sourcegraph.checks.registerCheckProvider>[0],
        providerFactory: ProxyResult<((context: sourcegraph.CheckContext<any>) => ProxiedCheckProvider) & ProxyValue>
    ): Unsubscribable & ProxyValue
}

export function createClientChecks(checkService: CheckService): ClientChecksAPI {
    return {
        $registerCheckProvider: (name, providerFactory) => {
            return proxyValue(
                checkService.registerCheckProvider(name, context => {
                    const provider = providerFactory(context)
                    return {
                        information: wrapRemoteObservable(provider.then(provider => provider.information)),
                        provideDiagnosticGroups: () =>
                            wrapRemoteObservable(provider.then(provider => provider.provideDiagnosticGroups())),
                    }
                })
            )
        },
        [proxyValueSymbol]: true,
    }
}
