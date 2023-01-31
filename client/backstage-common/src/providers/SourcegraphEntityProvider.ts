import { EntityProvider, EntityProviderConnection } from '@backstage/plugin-catalog-backend'
import { Config } from '@backstage/config'
import { SourcegraphService, createService } from '../client'
import { parseCatalog } from '../catalog/parsers'

const DEFAULT_QUERY = `"file:^catalog-info.yaml$"`

function withDefault(v: any, defaultValue: any) {
    return v ? v : defaultValue
}
export class SourcegraphEntityProvider implements EntityProvider {
    private connection?: EntityProviderConnection
    private readonly sourcegraph: SourcegraphService
    private query: string

    static create(config: Config) {
        return new SourcegraphEntityProvider(config)
    }

    private constructor(config: Config) {
        const endpoint = config.getString('sourcegraph.endpoint')
        const token = config.getString('sourcegraph.token')
        const sudoUsername = config.getOptionalString('sourcegraph.sudoUsername')
        this.query = withDefault(config.getOptionalString('sourcegraph.catalog_query'), DEFAULT_QUERY)
        console.log('sourcegraph query 🔍', this.query)

        this.sourcegraph = createService({ endpoint, token, sudoUsername })
    }

    getProviderName(): string {
        return 'sourcegraph-entity-provider'
    }

    async connect(connection: EntityProviderConnection): Promise<void> {
        this.connection = connection
    }

    async fullMutation() {
        console.log('STARTING SEARCH ⌛')
        const results = await this.sourcegraph.Search.SearchQuery(this.query)
        console.log('sourcegraph results', results)
        console.log('END SEARCH ⌛')

        const entities = parseCatalog(results, this.getProviderName())

        await this.connection?.applyMutation({
            type: 'full',
            entities: entities,
        })
    }
}
