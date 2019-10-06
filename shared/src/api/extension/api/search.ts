import { ProxyResult, proxyValue } from '@sourcegraph/comlink'
import { Subscribable, Unsubscribable } from 'rxjs'
import { QueryTransformer, SearchOptions, SearchQuery, TextSearchResult } from 'sourcegraph'
import { wrapRemoteObservable } from '../../client/api/common'
import { ClientSearchAPI } from '../../client/api/search'
import { syncSubscription } from '../../util'

/** @internal */
export class ExtSearch {
    constructor(private proxy: ProxyResult<ClientSearchAPI>) {}

    public findTextInFiles(query: SearchQuery, options?: SearchOptions): Subscribable<TextSearchResult[]> {
        return wrapRemoteObservable(this.proxy.$findTextInFiles({ query, options }))
    }

    public registerQueryTransformer(provider: QueryTransformer): Unsubscribable {
        return syncSubscription(this.proxy.$registerQueryTransformer(proxyValue(provider)))
    }
}
