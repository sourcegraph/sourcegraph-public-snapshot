import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { Subscription, Unsubscribable } from 'rxjs'
import { QueryTransformer } from 'sourcegraph'
import { ClientSearchAPI } from '../../client/api/search'

export class ExtSearch implements ProxyValue {
    public readonly [proxyValueSymbol] = true

    constructor(private proxy: ProxyResult<ClientSearchAPI>) {}

    public registerQueryTransformer(provider: QueryTransformer): Unsubscribable {
        const subscription = new Subscription()
        // tslint:disable:no-floating-promises
        this.proxy.$registerQueryTransformer(proxyValue(provider)).then(s => subscription.add(s))
        return subscription
    }
}
