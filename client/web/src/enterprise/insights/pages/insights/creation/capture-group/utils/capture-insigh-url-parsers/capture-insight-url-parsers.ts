import type { CaptureGroupFormFields } from '../../types'

type UnsupportedFields = 'step' | 'stepValue' | 'repoQuery' | 'repoMode'

export type CaptureInsightUrlValues = Omit<CaptureGroupFormFields, UnsupportedFields> & {
    repoQuery: string
}

export function encodeCaptureInsightURL(values: Partial<CaptureInsightUrlValues>): string {
    const parameters = new URLSearchParams()
    const keys = Object.keys(values) as (keyof CaptureInsightUrlValues)[]

    for (const key of keys) {
        const fields = values as CaptureInsightUrlValues

        switch (key) {
            case 'repoQuery':
            case 'groupSearchQuery': {
                parameters.set(key, encodeURIComponent(fields[key].toString()))
                break
            }
            case 'repositories': {
                parameters.set(key, fields[key].join(','))
                break
            }

            default: {
                parameters.set(key, fields[key].toString())
            }
        }
    }

    return parameters.toString()
}

export function decodeCaptureInsightURL(queryParameters: string): Partial<CaptureGroupFormFields> | null {
    try {
        const searchParameter = new URLSearchParams(decodeURIComponent(queryParameters))

        const repoQuery = decodeURIComponent(searchParameter.get('repoQuery') ?? '')
        const repositories = searchParameter.get('repositories')?.split(',')
        const title = searchParameter.get('title')
        const groupSearchQuery = decodeURIComponent(searchParameter.get('groupSearchQuery') ?? '')

        if (repositories || title || groupSearchQuery || repoQuery) {
            return {
                title: title ?? '',
                repositories: repositories ?? [],
                groupSearchQuery: groupSearchQuery ?? '',
                repoMode: repoQuery ? 'search-query' : 'urls-list',
                repoQuery: { query: repoQuery ?? '' },
            }
        }

        return null
    } catch {
        return null
    }
}
