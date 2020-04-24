import { Remote, ProxyMarked, proxy, proxyMarker } from '@sourcegraph/comlink'
import { from } from 'rxjs'
import { QueryTransformer, Unsubscribable } from 'sourcegraph'
import { TransformQuerySignature } from '../services/queryTransformer'
import { FeatureProviderRegistry } from '../services/registry'

/** @internal */
export interface ClientSearchAPI extends ProxyMarked {
    $registerQueryTransformer(transformer: Remote<QueryTransformer & ProxyMarked>): Unsubscribable & ProxyMarked
}

/** @internal */
export class ClientSearch implements ClientSearchAPI, ProxyMarked {
    public readonly [proxyMarker] = true

    constructor(private queryTransformerRegistry: FeatureProviderRegistry<{}, TransformQuerySignature>) {}

    public $registerQueryTransformer(
        transformer: Remote<QueryTransformer & ProxyMarked>
    ): Unsubscribable & ProxyMarked {
        return proxy(
            this.queryTransformerRegistry.registerProvider({}, query => from(transformer.transformQuery(query)))
        )
    }
}
