import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { wrapRemoteObservable } from './common'
import { Unsubscribable } from 'rxjs'
import { StatusService } from '../services/statusService'

export interface ClientStatusAPI extends ProxyValue {
    $registerStatusProvider(
        type: Parameters<typeof sourcegraph.status.registerStatusProvider>[0],
        providerFunction: ProxyResult<
            ((
                ...args: Parameters<sourcegraph.StatusProvider['provideStatus']>
            ) => ProxySubscribable<sourcegraph.Status | null | undefined>) &
                ProxyValue
        >
    ): Unsubscribable & ProxyValue
}

export function createClientStatus(statusService: StatusService): ClientStatusAPI {
    return {
        $registerStatusProvider: (type, providerFunction) => {
            return proxyValue(
                statusService.registerStatusProvider(type, {
                    provideStatus: (...args) => wrapRemoteObservable(providerFunction(...args)),
                })
            )
        },
        [proxyValueSymbol]: true,
    }
}
