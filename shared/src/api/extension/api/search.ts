import { Observable, of, Unsubscribable } from 'rxjs'
import { QueryTransformer, SearchResultProvider, Subscribable } from 'sourcegraph'
import { SearchAPI } from '../../client/api/search'
import { SearchResult } from '../../protocol/plainTypes'
import { ProviderMap, toProviderResultObservable } from './common'

export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>
    $provideSearchResult: (id: number, query: string) => Observable<SearchResult[] | null | undefined>
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
    public $provideSearchResult(id: number, query: string): Observable<SearchResult[] | null | undefined> {
        const provider = this.registrations.get<SearchResultProvider>(id)
        return toProviderResultObservable(
            new Promise<SearchResult[] | null | Subscribable<SearchResult[] | null | undefined> | undefined>(resolve =>
                resolve(provider.provideSearchResult(query))
            ),
            result => (result ? result.map(r => r) : result)
        )
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}
