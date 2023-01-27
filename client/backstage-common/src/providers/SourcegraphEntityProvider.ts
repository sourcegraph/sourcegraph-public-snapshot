import { EntityProvider, EntityProviderConnection } from '@backstage/plugin-catalog-backend'
import { Config } from '@backstage/config'
import { SourcegraphService, createService } from '../client'
import { parseCatalog } from '../catalog/parsers'

export class SourcegraphEntityProvider implements EntityProvider {
    private connection?: EntityProviderConnection
    private readonly sourcegraph: SourcegraphService

    static create(config: Config) {
        return new SourcegraphEntityProvider(config)
    }

    private constructor(config: Config) {
        const endpoint = config.getString('sourcegraph.endpoint')
        const token = config.getString('sourcegraph.token')
        const sudoUsername = config.getOptionalString('sourcegraph.sudoUsername')

        this.sourcegraph = createService({ endpoint, token, sudoUsername })
    }

    getProviderName(): string {
        return 'sourcegraph-entity-provider'
    }

    async connect(connection: EntityProviderConnection): Promise<void> {
        this.connection = connection
    }

    async fullMutation() {
        const results = await this.sourcegraph.Search.SearchQuery(`"file:^catalog-info.yaml$"`)

        const entities = parseCatalog(results, this.getProviderName())

        await this.connection?.applyMutation({
            type: 'full',
            entities: entities,
        })
    }
}
