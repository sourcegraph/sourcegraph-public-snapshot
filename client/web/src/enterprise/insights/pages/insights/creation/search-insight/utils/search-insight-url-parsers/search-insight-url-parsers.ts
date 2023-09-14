import { createDefaultEditSeries } from '../../../../../../components'
import type { SearchBasedInsightSeries } from '../../../../../../core'
import type { CreateInsightFormFields } from '../../types'

export function decodeSearchInsightUrl(queryParameters: string): Partial<CreateInsightFormFields> | null {
    try {
        const searchParameter = new URLSearchParams(decodeURIComponent(queryParameters))

        const repoQuery = searchParameter.get('repoQuery')
        const repositories = searchParameter.get('repositories')
        const title = searchParameter.get('title')
        const rawSeries = JSON.parse(searchParameter.get('series') ?? '[]') as SearchBasedInsightSeries[]
        const editableSeries = rawSeries.map(series => createDefaultEditSeries({ ...series, edit: false, valid: true }))

        if (repoQuery || repositories || title || editableSeries.length > 0) {
            return {
                title: title ?? '',
                repoQuery: { query: repoQuery ?? '' },
                repositories: repositories?.split(',') ?? [],
                series: editableSeries,
                repoMode: repoQuery ? 'search-query' : 'urls-list',
            }
        }

        return null
    } catch {
        return null
    }
}

type UnsupportedValues = 'series' | 'step' | 'visibility' | 'stepValue' | 'repoQuery' | 'repoMode'

export interface SearchInsightURLValues extends Omit<CreateInsightFormFields, UnsupportedValues> {
    repoQuery: string
    series: (Omit<SearchBasedInsightSeries, 'id'> & { id?: string | number })[]
}

export function encodeSearchInsightUrl(values: Partial<SearchInsightURLValues>): string {
    const parameters = new URLSearchParams()
    const keys = Object.keys(values) as (keyof SearchInsightURLValues)[]

    for (const key of keys) {
        const fields = values as SearchInsightURLValues

        switch (key) {
            case 'title':
            case 'repoQuery': {
                parameters.set(key, encodeURIComponent(fields[key].toString()))
                break
            }
            case 'repositories': {
                const encodedRepoURLs = fields[key].join(',')
                parameters.set(key, encodedRepoURLs)

                break
            }
            case 'series': {
                parameters.set(key, encodeURIComponent(JSON.stringify(fields[key])))

                break
            }
        }
    }

    return parameters.toString()
}
