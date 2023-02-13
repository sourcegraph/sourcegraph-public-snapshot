import { Config } from '@backstage/config'
import { SearchService, createService, createDummySearch } from '../client'
import { parserForType, ParserFunction } from './parsers'
import { EntityProvider, EntityProviderConnection } from '@backstage/plugin-catalog-backend'

export type EntityType = 'file' | 'grpc' | 'graphql'

abstract class BaseEntityProvider implements EntityProvider {
    private connection?: EntityProviderConnection
    private readonly sourcegraph: SearchService
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

        // TODO(@burmudar): remove - only temporary
        console.log(entityType, 'QUERY', this.query)
        this.disabled = this.query == ''
        if (this.disabled) {
            this.sourcegraph = createDummySearch()
        } else {
            this.sourcegraph = createService({ endpoint: this.endpoint, token, sudoUsername }).Search
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
        console.log(`STARTING ${this.entityType} SEARCH ⌛`)
        const results = await this.sourcegraph.searchQuery(this.query)
        console.log(`${results.length} items matched query ${this.query}`)
        console.log(`END ${this.entityType} SEARCH ⌛`)

        const entities = this.entityParseFn(results, this.getProviderName())

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
