import { ProxyValue, proxyValue, proxyValueSymbol } from 'comlink'
import { from, Subscription } from 'rxjs'
import { QueryTransformer, Unsubscribable } from 'sourcegraph'
import { TransformQuerySignature } from '../services/queryTransformer'
import { FeatureProviderRegistry } from '../services/registry'

/** @internal */
export interface ClientSearchAPI {
    $registerQueryTransformer(transformer: QueryTransformer): Unsubscribable & ProxyValue
}

/** @internal */
export class ClientSearch implements ClientSearchAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    private subscriptions = new Subscription()
    private queryTransformers = new Set<QueryTransformer>()

    constructor(private queryTransformerRegistry: FeatureProviderRegistry<{}, TransformQuerySignature>) {}

    public $registerQueryTransformer(transformer: QueryTransformer): Unsubscribable & ProxyValue {
        this.queryTransformerRegistry.registerProvider({}, query => from(transformer.transformQuery(query)))
        return proxyValue(new Subscription(() => this.queryTransformers.delete(transformer)))
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
