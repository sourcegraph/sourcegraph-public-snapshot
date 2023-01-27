import { ANNOTATION_LOCATION, ANNOTATION_ORIGIN_LOCATION } from '@backstage/catalog-model'
import { CatalogProcessorEntityResult, DeferredEntity, parseEntityYaml } from '@backstage/plugin-catalog-backend'
import { SearchResult } from '../client'

export const parseCatalog = (src: SearchResult[], providerName: string): DeferredEntity[] => {
    const results: DeferredEntity[] = []

    src.forEach((r: SearchResult) => {
        const location = {
            type: 'url',
            target: `${r.repository}/catalog-info.yaml`,
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
