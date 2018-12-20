import * as clientTypes from '@sourcegraph/extension-api-types'
import { Observable, Unsubscribable } from 'rxjs'
import { SearchResult } from 'sourcegraph'
import { QueryTransformer, SearchResultProvider, Subscribable } from 'sourcegraph'
import { SearchAPI } from '../../client/api/search'
import { ProviderMap, toProviderResultObservable } from './common'
import { fromSearchResult } from './types'

export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>
    $provideSearchResults: (id: number, query: string) => Observable<clientTypes.SearchResult[] | null | undefined>
}

export class ExtSearch implements ExtSearchAPI, Unsubscribable {
    private registrations = new ProviderMap<QueryTransformer | SearchResultProvider>(id => this.proxy.$unregister(id))
    constructor(private proxy: SearchAPI) {}

    public registerQueryTransformer(provider: QueryTransformer): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerQueryTransformer(id)
        return subscription
    }

    public $transformQuery(id: number, query: string): Promise<string> {
        const provider = this.registrations.get<QueryTransformer>(id)
        return Promise.resolve(provider.transformQuery(query))
    }

    public registerSearchResultProvider(provider: SearchResultProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerSearchResultProvider(id)
        return subscription
    }
    public $provideSearchResults(id: number, query: string): Observable<clientTypes.SearchResult[] | null | undefined> {
        const provider = this.registrations.get<SearchResultProvider>(id)
        return toProviderResultObservable(
            new Promise<SearchResult[] | null | Subscribable<SearchResult[] | null | undefined> | undefined>(resolve =>
                resolve(provider.provideSearchResults(query))
            ),
            result => (result ? result.map((r: SearchResult) => fromSearchResult(r)) : result)
        )
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}
