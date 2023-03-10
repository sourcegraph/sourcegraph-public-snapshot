import { ANNOTATION_LOCATION, ANNOTATION_ORIGIN_LOCATION, ApiEntity } from '@backstage/catalog-model'
import { CatalogProcessorEntityResult, DeferredEntity, parseEntityYaml } from '@backstage/plugin-catalog-backend'

import { SearchResult } from '../client'

import type { EntityType } from './providers'

export type ParserFunction = (results: SearchResult[], providerName: string) => DeferredEntity[]

const apiEntityDefinitionParser = (
    searchResults: SearchResult[],
    providerName: string,
    entityType: EntityType
): DeferredEntity[] => {
    const results: DeferredEntity[] = []

    // TODO(@burmudar): remove - only temporary
    console.log(`parsing ${searchResults.length} proto results`)
    searchResults.forEach((r: SearchResult) => {
        const location = {
            type: 'url',
            target: `${r.repository}/${r.filename}`,
        }
        const definition = Buffer.from(r.fileContent, 'utf8').toString()
        const entity: ApiEntity = {
            kind: 'API',
            apiVersion: 'backstage.io/v1alpha1',
            metadata: {
                name: r.filename,
                annotations: {
                    [ANNOTATION_LOCATION]: `url:${location.target}`,
                    [ANNOTATION_ORIGIN_LOCATION]: providerName,
                },
            },
            spec: {
                type: entityType,
                lifecycle: 'production',
                owner: 'engineering',
                definition: definition,
            },
        }
        results.push({
            entity: entity,
            locationKey: `https://${location.target}`,
        })
    })

    console.log(`${results.length} ${entityType} entities`)
    return results
}

export const parseCatalogContent: ParserFunction = (
    searchResults: SearchResult[],
    providerName: string
): DeferredEntity[] => {
    const results: DeferredEntity[] = []

    searchResults.forEach((r: SearchResult) => {
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

const apiEntityParser =
    (type: EntityType): ParserFunction =>
    (src: SearchResult[], providerName: string): DeferredEntity[] =>
        apiEntityDefinitionParser(src, providerName, type)

export const parserForType = (entityType: EntityType): ParserFunction => {
    if (entityType === 'file') {
        return parseCatalogContent
    } else if (entityType === 'grpc' || entityType === 'graphql') {
        return apiEntityParser(entityType)
    } else {
        throw new Error(`unknown Entity Type: ${entityType}`)
    }
}
