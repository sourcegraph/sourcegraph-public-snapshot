import * as clientType from '@sourcegraph/extension-api-types'
import { from, Observable, Subscription } from 'rxjs'
import { map } from 'rxjs/operators'
import { createProxyAndHandleRequests } from '../../common/proxy'
import { ExtSearch } from '../../extension/api/search'
import { Connection } from '../../protocol/jsonrpc2/connection'
import { TransformQuerySignature } from '../services/queryTransformer'
import { FeatureProviderRegistry } from '../services/registry'
import { ProvideSearchResultSignature } from '../services/searchResults'
import { SubscriptionMap } from './common'

/** @internal */
export interface SearchAPI {
    $registerQueryTransformer(id: number): void
    $registerSearchResultProvider(id: number): void
    $unregister(id: number): void
}

/** @internal */
export class Search implements SearchAPI {
    private subscriptions = new Subscription()
    private registrations = new SubscriptionMap()
    private proxy: ExtSearch

    constructor(
        connection: Connection,
        private queryTransformerRegistry: FeatureProviderRegistry<{}, TransformQuerySignature>,
        private searchResultProviderRegistry: FeatureProviderRegistry<{}, ProvideSearchResultSignature>
    ) {
        this.subscriptions.add(this.registrations)

        this.proxy = createProxyAndHandleRequests('search', connection, this)
    }

    public $registerQueryTransformer(id: number): void {
        this.registrations.add(
            id,
            this.queryTransformerRegistry.registerProvider(
                {},
                (query: string): Observable<string> => from(this.proxy.$transformQuery(id, query))
            )
        )
    }

    public $registerSearchResultProvider(id: number): void {
        this.registrations.add(
            id,
            this.searchResultProviderRegistry.registerProvider(
                {},
                (query: string): Observable<clientType.SearchResult[]> =>
                    from(this.proxy.$provideSearchResult(id, query)).pipe(map(result => result || []))
            )
        )
    }

    public $unregister(id: number): void {
        this.registrations.remove(id)
    }

    public unsubscribe(): void {
        this.subscriptions.unsubscribe()
    }
}
