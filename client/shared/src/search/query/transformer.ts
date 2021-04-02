import { replaceRange } from '../../util/strings'
import { FilterType } from './filters'
import { Filter } from './token'
import { filterExists } from './validate'

export function appendContextFilter(query: string, searchContextSpec: string | undefined): string {
    return !filterExists(query, FilterType.context) && searchContextSpec
        ? `context:${searchContextSpec} ${query}`
        : query
}

export function omitFilter(query: string, filter: Filter): string {
    let finalQuery = replaceRange(query, filter.range)
    if (filter.range.start === 0) {
        // Remove space at the start
        finalQuery = finalQuery.slice(1)
    }
    return finalQuery
}
