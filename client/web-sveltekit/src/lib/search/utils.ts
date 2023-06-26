import type { ContentMatch, MatchItem } from '$lib/shared'

export interface SidebarFilter {
    value: string
    label: string
    count?: number
    limitHit?: boolean
    kind: 'file' | 'repo' | 'lang' | 'utility'
    runImmediately?: boolean
}

/**
 * A context object provided on pages with the main search input to interact
 * with the main input.
 */
export interface SearchPageContext {
    setQuery(query: string | ((query: string) => string)): void
}
