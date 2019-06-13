import { ProxyResult, ProxyValue, proxyValue, proxyValueSymbol } from '@sourcegraph/comlink'
import { from } from 'rxjs'
import { first, switchMap } from 'rxjs/operators'
import { QueryTransformer, TextSearchResult, Unsubscribable } from 'sourcegraph'
import { ProxySubscribable, toProxyableSubscribable } from '../../extension/api/common'
import { TransformQuerySignature } from '../services/queryTransformer'
import { FeatureProviderRegistry } from '../services/registry'
import { ProvideTextSearchResultsParams, SearchProviderRegistry } from '../services/searchProviders'
import { wrapRemoteObservable } from './common'

/** @internal */
export interface ClientSearchAPI extends ProxyValue {
    $findTextInFiles(params: ProvideTextSearchResultsParams): ProxySubscribable<TextSearchResult[]> & ProxyValue
    $registerQueryTransformer(transformer: ProxyResult<QueryTransformer & ProxyValue>): Unsubscribable & ProxyValue
    $registerTextSearchProvider(
        providerFunction: ProxyResult<
            ((params: ProvideTextSearchResultsParams) => ProxySubscribable<TextSearchResult[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue
}

/** @internal */
export class ClientSearch implements ClientSearchAPI, ProxyValue {
    public readonly [proxyValueSymbol] = true

    constructor(
        private queryTransformerRegistry: FeatureProviderRegistry<{}, TransformQuerySignature>,
        private searchProviderRegistry: SearchProviderRegistry
    ) {}

    public $findTextInFiles(
        params: ProvideTextSearchResultsParams
    ): ProxySubscribable<TextSearchResult[]> & ProxyValue {
        return toProxyableSubscribable(
            this.searchProviderRegistry.getResults(params).pipe(
                // TODO!(sqs): this only takes the first inner subscribable, so it won't catch search
                // providers registered after the initial call to findTextInFiles.
                first(),
                switchMap(providerResults => providerResults)
            ),
            items => items
        )
    }

    public $registerQueryTransformer(
        transformer: ProxyResult<QueryTransformer & ProxyValue>
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.queryTransformerRegistry.registerProvider({}, query => from(transformer.transformQuery(query)))
        )
    }

    public $registerTextSearchProvider(
        providerFunction: ProxyResult<
            ((params: ProvideTextSearchResultsParams) => ProxySubscribable<TextSearchResult[]>) & ProxyValue
        >
    ): Unsubscribable & ProxyValue {
        return proxyValue(
            this.searchProviderRegistry.registerProvider({}, params => wrapRemoteObservable(providerFunction(params)))
        )
    }
}
