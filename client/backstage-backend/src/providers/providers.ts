import { Config } from '@backstage/config'
import { EntityProvider, EntityProviderConnection } from '@backstage/plugin-catalog-backend'

import { SearchService, createService } from '../client'

import { parserForType, ParserFunction } from './parsers'

export type EntityType = 'file' | 'grpc' | 'graphql'

abstract class BaseEntityProvider implements EntityProvider {
    private connection?: EntityProviderConnection
    private sourcegraph?: SearchService
    private query: string
    private endpoint: string
    private disabled: boolean = false
    private readonly entityType: EntityType
    private entityParseFn: ParserFunction

    protected constructor(config: Config, entityType: EntityType) {
        const token = config.getString('sourcegraph.token')
        const sudoUsername = config.getOptionalString('sourcegraph.sudoUsername')
        const queryConfig = config.getConfig(`sourcegraph.${entityType}`)

        this.endpoint = config.getString('sourcegraph.endpoint')
        this.entityType = entityType
        this.entityParseFn = parserForType(entityType)
        this.query = queryConfig.getOptionalString('query') ?? ''

        this.disabled = this.query == ''
        if (!this.disabled) {
            createService({ endpoint: this.endpoint, token, sudoUsername }).then(
                service => (this.sourcegraph = service.Search)
            )
        } else {
            console.error("query '' is invalid - provider will be in disabled state")
        }
    }

    getProviderName(): string {
        return `url: ${this.endpoint}/sourcegraph-${this.entityType}-entity-provider`
    }

    async connect(connection: EntityProviderConnection): Promise<void> {
        this.connection = connection
    }

    async run(): Promise<void> {
        if (!this.connection) {
            throw new Error('connection not initialized')
        }

        await this.fullMutation()
    }

    async fullMutation() {
        // TODO(@burmudar): remove - only temporary
        if (this.disabled) {
            console.log(`${this.getProviderName} is disabled`)
            return
        }
        const results = await this.sourcegraph?.searchQuery(this.query)
        console.log(`${results?.length} items matched query ${this.query}`)

        const entities = this.entityParseFn(results ?? [], this.getProviderName())

        await this.connection?.applyMutation({
            type: 'full',
            entities: entities,
        })
    }
}

export class GrpcEntityProvider extends BaseEntityProvider {
    constructor(config: Config) {
        super(config, 'grpc')
    }
}

export class GraphQLEntityProvider extends BaseEntityProvider {
    constructor(config: Config) {
        super(config, 'graphql')
    }
}

export class YamlFileEntityProvider extends BaseEntityProvider {
    constructor(config: Config) {
        super(config, 'file')
    }
}
