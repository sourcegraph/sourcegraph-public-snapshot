import { quoteIfNeeded } from '../../search'

export function queryWithReplacementText(query: string, replace: string): string {
    return `${query.includes('replace:') ? query.slice(0, query.indexOf(' replace:')) : query} replace:${quoteIfNeeded(
        replace
    )}`
}
