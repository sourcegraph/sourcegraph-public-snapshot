import { FilterType } from './query/filters'
import { scanSearchQuery } from './query/scanner'

/**
 * Version contexts are a lists of repos at specific revisions.
 * When a version context is set, searching and browsing is restricted to
 * the repos and revisions in the version context.
 */
export interface VersionContextProps {
    /**
     * The currently selected version context. This is undefined when there is no version context active.
     * When undefined, we use the default behavior, which allows access to all repos.
     */
    versionContext: string | undefined
}

export function appendContextFilterToQuery(query: string, searchContextSpec: string | undefined): string {
    return !isContextFilterInQuery(query) && searchContextSpec ? `context:${searchContextSpec} ${query}` : query
}

export function isContextFilterInQuery(query: string): boolean {
    const scannedQuery = scanSearchQuery(query)
    return (
        scannedQuery.type === 'success' &&
        scannedQuery.term.some(
            token => token.type === 'filter' && token.field.value.toLowerCase() === FilterType.context
        )
    )
}
