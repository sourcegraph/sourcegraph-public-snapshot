import { CaptureGroupFormFields } from '../../types'

export type CaptureInsightUrlValues = Omit<CaptureGroupFormFields, 'step' | 'stepValue'>

export function encodeCaptureInsightURL(values: Partial<CaptureInsightUrlValues>): string {
    const parameters = new URLSearchParams()
    const keys = Object.keys(values) as (keyof CaptureInsightUrlValues)[]

    for (const key of keys) {
        const fields = values as CaptureInsightUrlValues

        switch (key) {
            case 'groupSearchQuery': {
                parameters.set(key, encodeURIComponent(fields[key].toString()))
                break
            }

            case 'repoQuery': {
                parameters.set(key, encodeURIComponent(fields[key].query))
                break
            }

            default:
                parameters.set(key, fields[key].toString())
        }
    }

    return parameters.toString()
}

export function decodeCaptureInsightURL(queryParameters: string): Partial<CaptureGroupFormFields> | null {
    try {
        const searchParameter = new URLSearchParams(decodeURIComponent(queryParameters))

        const repositories = searchParameter.get('repositories')
        const title = searchParameter.get('title')
        const groupSearchQuery = decodeURIComponent(searchParameter.get('groupSearchQuery') ?? '')
        const repoMode = searchParameter.get('repoMode') as any
        const repoQuery = searchParameter.get('repoQuery')

        if (repositories || title || groupSearchQuery || repoMode || repoQuery) {
            return {
                title: title ?? '',
                repositories: repositories ?? '',
                groupSearchQuery: groupSearchQuery ?? '',
                repoMode: repoMode ?? 'urls-list',
                repoQuery: { query: repoQuery ?? '' },
            }
        }

        return null
    } catch {
        return null
    }
}
