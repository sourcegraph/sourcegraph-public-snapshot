import { Validator } from '../../../../../../form/hooks/useField'

export const validRegexp: Validator<string> = (value = '') => {
    if (value.trim() === '') {
        return
    }

    try {
        new RegExp(value)

        return
    } catch {
        return 'Must be a valid regular expression string'
    }
}

interface InsightRepositoriesFilter {
    include: string
    exclude: string
}

export function getSerializedRepositoriesFilter(filter: InsightRepositoriesFilter): string {
    const { include, exclude } = filter
    const includeString = include ? `repo:${include}` : ''
    const excludeString = exclude ? `-repo:${exclude}` : ''

    return `${includeString} ${excludeString}`.trim()
}

type InsightContextsFilter = string

export function getSerializedSearchContextFilter(filter: InsightContextsFilter): string {
    return filter !== '' ? filter : 'global (default)'
}
