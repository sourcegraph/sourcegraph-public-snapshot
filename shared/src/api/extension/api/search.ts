import { Unsubscribable } from 'rxjs'
import { QueryTransformer, SearchResultsProvider } from 'sourcegraph'
import { SearchAPI } from '../../client/api/search'
import { GenericSearchResult } from '../../protocol/plainTypes'
import { ProviderMap } from './common'

export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>
    $provideIssueResults: (id: number, query: string) => Promise<GenericSearchResult[] | null>
}

export class ExtSearch implements ExtSearchAPI, Unsubscribable {
    private registrations = new ProviderMap<QueryTransformer | SearchResultsProvider>(id => this.proxy.$unregister(id))
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

    public registerSearchResultsProvider(provider: SearchResultsProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerSearchResultsProvider(id)
        return subscription
    }
    public $provideIssueResults(id: number, query: string): Promise<GenericSearchResult[]> {
        const provider = this.registrations.get<SearchResultsProvider>(id)
        return Promise.resolve(provider.provideSearchResults(query))
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}
