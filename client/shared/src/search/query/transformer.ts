import { replaceRange } from '../../util/strings'
import { Filter } from './token'
import { isContextFilterInQuery } from './validate'

export function appendContextFilter(query: string, searchContextSpec: string | undefined): string {
    return !isContextFilterInQuery(query) && searchContextSpec ? `context:${searchContextSpec} ${query}` : query
}

export const omitContextFilter = (query: string, contextFilter: Filter): string => {
    let finalQuery = replaceRange(query, contextFilter.range)
    if (contextFilter.range.start === 0) {
        // Remove space at the start
        finalQuery = finalQuery.slice(1)
    }
    return finalQuery
}
