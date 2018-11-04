import { Unsubscribable } from 'rxjs'
import { QueryTransformer } from 'sourcegraph'
import { SearchAPI } from 'src/client/api/search'
import { ProviderMap } from './common'

export interface ExtSearchAPI {
    $transformQuery: (id: number, query: string) => Promise<string>
}

export class ExtSearch implements ExtSearchAPI {
    private registrations = new ProviderMap<QueryTransformer>(id => this.proxy.$unregister(id))
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
}
