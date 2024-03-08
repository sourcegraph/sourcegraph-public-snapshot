import { getContext, setContext, type ComponentProps } from 'svelte'

import type { QueryStateStore } from '$lib/search/state'
import type { SearchMatch } from '$lib/shared'

import PreviewPanel from './PreviewPanel.svelte'

interface SearchResultsContext {
    isExpanded(match: SearchMatch): boolean
    setExpanded(match: SearchMatch, expanded: boolean): void
    setPreview(props: ComponentProps<PreviewPanel> | undefined): void
    queryState: QueryStateStore
}

const CONTEXT_KEY = {}

export function getSearchResultsContext(): SearchResultsContext {
    return getContext(CONTEXT_KEY)
}

export function setSearchResultsContext(context: SearchResultsContext): SearchResultsContext {
    return setContext(CONTEXT_KEY, context)
}
