import { ProxyResult, proxyValue } from '@sourcegraph/comlink'
import { Unsubscribable } from 'rxjs'
import { QueryTransformer } from 'sourcegraph'
import { ClientSearchAPI } from '../../client/api/search'
import { syncSubscription } from '../../util'

export class ExtSearch {
    constructor(private proxy: ProxyResult<ClientSearchAPI>) {}

    public registerQueryTransformer(provider: QueryTransformer): Unsubscribable {
        return syncSubscription(this.proxy.$registerQueryTransformer(proxyValue(provider)))
    }
}
