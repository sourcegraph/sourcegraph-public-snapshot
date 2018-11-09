import { Unsubscribable } from 'rxjs'
import { IssueResultsProvider, QueryTransformer } from 'sourcegraph'
import { SearchAPI } from 'src/client/api/search'
import { IssueResult } from 'src/protocol/plainTypes'
import { ProviderMap } from './common'

export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>
    $provideIssueResults: (id: number, query: string) => Promise<IssueResult[] | null>
}

export class ExtSearch implements ExtSearchAPI {
    private registrations = new ProviderMap<QueryTransformer | IssueResultsProvider>(id => this.proxy.$unregister(id))
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

    public registerIssueResultsProvider(provider: IssueResultsProvider): Unsubscribable {
        const { id, subscription } = this.registrations.add(provider)
        this.proxy.$registerIssueResultsProvider(id)
        return subscription
    }

    public $provideIssueResults(id: number, query: string): Promise<IssueResult[]> {
        const provider = this.registrations.get<IssueResultsProvider>(id)
        return Promise.resolve(provider.provideIssueResults(query))
    }
}
