import {
  EntityProvider,
  EntityProviderConnection,
} from '@backstage/plugin-catalog-backend';
import { Config } from '@backstage/config';
import { SearchQuery, SourcegraphClient } from '../client/SourcegraphClient';
import { parseCatalog } from '../catalog/parsers';

export class SourcegraphEntityProvider implements EntityProvider {
  private connection?: EntityProviderConnection;
  private readonly sourcegraph: SourcegraphClient;

  static create(config: Config) {
    return new SourcegraphEntityProvider(config);
  }

  private constructor(config: Config) {
    this.sourcegraph = SourcegraphClient.create(config)
  }

  getProviderName(): string {
    return 'sourcegraph-entity-provider';
  }

  async connect(connection: EntityProviderConnection): Promise<void> {
    this.connection = connection;
  }

  async fullMutation() {
    const query: SearchQuery = new SearchQuery(`"file:^catalog-info.yaml$"`)

    const results = await this.sourcegraph.fetch(query)

    const entities = parseCatalog(results, this.getProviderName());

    await this.connection?.applyMutation({
      type: "full",
      entities: entities
    });
  }


}
