import * as comlink from '@sourcegraph/comlink'
import { Unsubscribable } from 'rxjs'
import { QueryTransformer } from 'sourcegraph'
import { ClientSearchAPI } from '../../client/api/search'
import { syncSubscription } from '../../util'

export class ExtSearch {
    constructor(private proxy: comlink.Remote<ClientSearchAPI>) {}

    public registerQueryTransformer(provider: QueryTransformer): Unsubscribable {
        return syncSubscription(this.proxy.$registerQueryTransformer(comlink.proxy(provider)))
    }
}
