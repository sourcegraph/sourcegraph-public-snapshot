import { Unsubscribable, Observable } from 'rxjs'
import { QueryTransformer, SearchResultProvider } from 'sourcegraph'
import { SearchAPI } from '../../client/api/search'
import { SearchResult } from '../../protocol/plainTypes'
import { ProviderMap, toProviderResultObservable } from './common'

export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>
    $provideSearchResults: (id: number, query: string) => Promise<SearchResult[] | null>
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
    public $provideSearchResults(id: number, query: string): Promise<SearchResult[]> {
        const provider = this.registrations.get<SearchResultProvider>(id)
        return provider.provideSearchResults(query)
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}
