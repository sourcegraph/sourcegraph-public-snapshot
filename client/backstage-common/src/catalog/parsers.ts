import { ANNOTATION_LOCATION, ANNOTATION_ORIGIN_LOCATION, ApiEntity } from '@backstage/catalog-model'
import { CatalogProcessorEntityResult, DeferredEntity, parseEntityYaml } from '@backstage/plugin-catalog-backend'
import { SearchResult } from '../client'

export type ParserFunction = (results: SearchResult[], providerName: string) => DeferredEntity[];

const parseEntityAsType = (searchResults: SearchResult[], providerName: string, entityType: string): DeferredEntity[] => {
  const results: DeferredEntity[] = []

  console.log(`parsing ${searchResults.length} proto results`)
  searchResults.forEach((r: SearchResult) => {
    const location = {
      type: 'url',
      target: `${r.repository}/${r.filename}`,
    }
    const protoContent = Buffer.from(r.fileContent, 'utf8').toString()
    const entity: ApiEntity = {
      kind: 'API',
      apiVersion: 'backstage.io/v1alpha1',
      metadata: {
        name: r.filename,
        annotations: {
          [ANNOTATION_LOCATION]: `url:${location.target}`,
          [ANNOTATION_ORIGIN_LOCATION]: providerName,
        }
      },
      spec: {
        type: entityType,
        lifecycle: 'production',
        owner: 'engineering',
        definition: protoContent,
      }

    }
    results.push({
      entity: entity,
      locationKey: `https://${location.target}`
    })
  });

  console.log(`${results.length} api entities`)
  return results;
}

export const parseCatalog: ParserFunction = (src: SearchResult[], providerName: string): DeferredEntity[] => {
  const results: DeferredEntity[] = []

  src.forEach((r: SearchResult) => {
    const location = {
      type: 'url',
      target: `https://${r.repository}/${r.filename}`,
    }
    const yaml = Buffer.from(r.fileContent, 'utf8')

    for (const item of parseEntityYaml(yaml, location)) {
      const parsed = item as CatalogProcessorEntityResult
      results.push({
        entity: {
          ...parsed.entity,
          metadata: {
            ...parsed.entity.metadata,
            annotations: {
              ...parsed.entity.metadata.annotations,
              [ANNOTATION_LOCATION]: `url:${parsed.location.target}`,
              [ANNOTATION_ORIGIN_LOCATION]: providerName,
            },
          },
        },
        locationKey: parsed.location.target,
      })
    }
  })
  return results
}

export const parserForType = (type: string): ParserFunction => {
  return (src: SearchResult[], providerName: string): DeferredEntity[] => parseEntityAsType(src, providerName, type)
}

export const parserGrpc = parserForType('grpc')
export const parserGraphQL = parserForType('graphql')
