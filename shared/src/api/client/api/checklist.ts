import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import * as sourcegraph from 'sourcegraph'
import { ProxySubscribable } from '../../extension/api/common'
import { wrapRemoteObservable } from './common'
import { Unsubscribable } from 'rxjs'
import { ChecklistService } from '../services/checklistService'

export interface ClientChecklistAPI extends ProxyValue {
    $registerChecklistProvider(
        type: Parameters<typeof sourcegraph.checklist.registerChecklistProvider>[0],
        providerFunction: ProxyResult<
            ((
                ...args: Parameters<sourcegraph.ChecklistProvider['provideChecklistItems']>
            ) => ProxySubscribable<sourcegraph.ChecklistItem[] | null | undefined>) &
                ProxyValue
        >
    ): Unsubscribable & ProxyValue
}

export function createClientChecklist(checklistService: ChecklistService): ClientChecklistAPI {
    return {
        $registerChecklistProvider: (type, providerFunction) => {
            return proxyValue(
                checklistService.registerChecklistProvider(type, {
                    provideChecklistItems: (...args) => wrapRemoteObservable(providerFunction(...args)),
                })
            )
        },
        [proxyValueSymbol]: true,
    }
}
