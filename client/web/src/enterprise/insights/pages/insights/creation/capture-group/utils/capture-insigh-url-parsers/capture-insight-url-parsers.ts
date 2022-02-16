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
        const allRepos = searchParameter.get('allRepos')

        if (repositories || title || groupSearchQuery || allRepos) {
            return {
                title: title ?? '',
                repositories: repositories ?? '',
                allRepos: !!allRepos,
                groupSearchQuery: groupSearchQuery ?? '',
            }
        }

        return null
    } catch {
        return null
    }
}
