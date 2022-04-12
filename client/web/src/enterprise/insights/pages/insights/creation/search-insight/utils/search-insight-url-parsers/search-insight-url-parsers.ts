import { SearchBasedInsightSeries } from '../../../../../../core/types'
import { createDefaultEditSeries } from '../../components/search-insight-creation-content/hooks/use-editable-series'
import { CreateInsightFormFields } from '../../types'

export function decodeSearchInsightUrl(queryParameters: string): Partial<CreateInsightFormFields> | null {
    try {
        const searchParameter = new URLSearchParams(decodeURIComponent(queryParameters))

        const repositories = searchParameter.get('repositories')
        const title = searchParameter.get('title')
        const rawSeries = JSON.parse(searchParameter.get('series') ?? '[]') as SearchBasedInsightSeries[]
        const editableSeries = rawSeries.map(series => createDefaultEditSeries({ ...series, edit: false, valid: true }))
        const allRepos = searchParameter.get('allRepos')

        if (repositories || title || editableSeries.length > 0 || allRepos) {
            return {
                title: title ?? '',
                repositories: repositories ?? '',
                allRepos: !!allRepos,
                series: editableSeries,
            }
        }

        return null
    } catch {
        return null
    }
}

type UnsupportedValues = 'series' | 'step' | 'visibility' | 'stepValue'

export interface SearchInsightURLValues extends Omit<CreateInsightFormFields, UnsupportedValues> {
    series: (Omit<SearchBasedInsightSeries, 'id'> & { id?: string })[]
}

export function encodeSearchInsightUrl(values: Partial<SearchInsightURLValues>): string {
    const parameters = new URLSearchParams()
    const keys = Object.keys(values) as (keyof SearchInsightURLValues)[]

    for (const key of keys) {
        const fields = values as SearchInsightURLValues

        switch (key) {
            case 'title':
            case 'repositories':
            case 'allRepos': {
                parameters.set(key, fields[key].toString())

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
