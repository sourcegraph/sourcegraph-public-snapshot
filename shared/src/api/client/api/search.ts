import { ProxyResult, ProxyValue, proxy, proxyMarker } from '@sourcegraph/comlink'
import { from } from 'rxjs'
import { QueryTransformer, Unsubscribable } from 'sourcegraph'
import { TransformQuerySignature } from '../services/queryTransformer'
import { FeatureProviderRegistry } from '../services/registry'

/** @internal */
export interface ClientSearchAPI extends ProxyValue {
    $registerQueryTransformer(transformer: ProxyResult<QueryTransformer & ProxyValue>): Unsubscribable & ProxyValue
}

/** @internal */
export class ClientSearch implements ClientSearchAPI, ProxyValue {
    public readonly [proxyMarker] = true

    constructor(private queryTransformerRegistry: FeatureProviderRegistry<{}, TransformQuerySignature>) {}

    public $registerQueryTransformer(
        transformer: ProxyResult<QueryTransformer & ProxyValue>
    ): Unsubscribable & ProxyValue {
        return proxy(
            this.queryTransformerRegistry.registerProvider({}, query => from(transformer.transformQuery(query)))
        )
    }
}
