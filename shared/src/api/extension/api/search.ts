import { Unsubscribable } from 'rxjs'
import { IssueResultsProvider, QueryTransformer } from 'sourcegraph'
import { SearchAPI } from '../../client/api/search'
import { IssueResult } from '../../protocol/plainTypes'
import { ProviderMap } from './common'

export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>
    $provideIssueResults: (id: number, query: string) => Promise<IssueResult[] | null>
}

export class ExtSearch implements ExtSearchAPI, Unsubscribable {
    private registrations = new ProviderMap<QueryTransformer | IssueResultsProvider>(id => this.proxy.$unregister(id))
    constructor(private proxy: SearchAPI) {}

    public registerQueryTransformer(provider: QueryTransformer): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerQueryTransformer(id)
        return subscription
    }

    public $transformQuery(id: number, query: string): Promise<string> {
        const provider = this.registrations.get<QueryTransformer>(id)
        console.log('transform provider', provider, provider.transformQuery('query'))
        return Promise.resolve(provider.transformQuery(query))
    }

    public registerIssueResultsProvider(provider: IssueResultsProvider): Unsubscribable {
        console.log('register issue results provider in extension api')
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerIssueResultsProvider(id)
        return subscription
    }
    public $provideIssueResults(id: number, query: string): Promise<IssueResult[]> {
        console.log('id', id)
        const provider = this.registrations.get<IssueResultsProvider>(id)
        console.log('provider', provider, provider.provideIssueResults('query'))
        return Promise.resolve(provider.provideIssueResults(query))
    }

    public unsubscribe(): void {
        this.registrations.unsubscribe()
    }
}
