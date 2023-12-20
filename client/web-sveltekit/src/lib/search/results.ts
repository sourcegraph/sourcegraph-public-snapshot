import { sortBy } from 'lodash'

import {
    appendFilter,
    buildSearchURLQuery,
    FilterKind,
    findFilter,
    getMatchUrl,
    omitFilter,
    truncateGroups,
    type ContentMatch,
    type OwnerMatch,
    type RepositoryMatch,
    type MatchGroup,
    type Range,
} from '$lib/shared'

import type { QueryState } from './state'
import { resultToMatchGroups } from './utils'

const REPO_DESCRIPTION_CHAR_LIMIT = 500

export function limitDescription(value: string): string {
    return value.length <= REPO_DESCRIPTION_CHAR_LIMIT ? value : value.slice(0, REPO_DESCRIPTION_CHAR_LIMIT) + '...'
}

export interface Meta {
    key: string
    value?: string | null
}

export function getMetadata(result: RepositoryMatch): Meta[] {
    const { metadata } = result
    if (!metadata) {
        return []
    }
    return sortBy(
        Object.entries(metadata).map(([key, value]) => ({ key, value })),
        ['key', 'value']
    )
}

export function buildSearchURLQueryForMeta(queryState: QueryState, meta: Meta): string {
    const query = appendFilter(
        queryState.query,
        'repo',
        meta.value ? `has.meta(${meta.key}:${meta.value})` : `has.meta(${meta.key})`
    )

    return buildSearchURLQuery(
        query,
        queryState.patternType,
        queryState.caseSensitive,
        queryState.searchContext,
        queryState.searchMode
    )
}

export function getOwnerDisplayName(result: OwnerMatch): string {
    switch (result.type) {
        case 'team': {
            return result.displayName || result.name || result.handle || result.email || 'Unknown team'
        }
        case 'person': {
            return (
                result.user?.displayName || result.user?.username || result.handle || result.email || 'Unknown person'
            )
        }
    }
}

export function getOwnerMatchURL(result: OwnerMatch): string | null {
    const url = getMatchUrl(result)
    return /^(\/teams\/|\/users\/|mailto:)/.test(url) ? url : null
}

export function buildSearchURLQueryForOwner(queryState: QueryState, result: OwnerMatch): string {
    const handle = result.handle || result.email
    if (!handle) {
        return ''
    }

    let query = queryState.query
    const selectFilter = findFilter(queryState.query, 'select', FilterKind.Global)
    if (selectFilter && selectFilter.value?.value === 'file.owners') {
        query = omitFilter(query, selectFilter)
    }
    query = appendFilter(query, 'file', `has.owner(${handle})`)

    return buildSearchURLQuery(
        query,
        queryState.patternType,
        queryState.caseSensitive,
        queryState.searchContext,
        queryState.searchMode
    )
}

function sumHighlightRanges(count: number, item: MatchGroup): number {
    return count + item.matches.length
}

export function rankContentMatch(
    result: ContentMatch,
    ranking: (groups: MatchGroup[]) => MatchGroup[],
    maxMatches: number,
    contextLines: number
): {
    expandedMatchGroups: MatchGroup[]
    collapsedMatchGroups: MatchGroup[]
    hiddenMatchesCount: number
} {
    const expandedMatchGroups = ranking(resultToMatchGroups(result))
    const collapsedMatchGroups = truncateGroups(expandedMatchGroups, maxMatches, contextLines)

    const highlightRangesCount = expandedMatchGroups.reduce(sumHighlightRanges, 0)
    const collapsedHighlightRangesCount = collapsedMatchGroups.reduce(sumHighlightRanges, 0)
    const hiddenMatchesCount = highlightRangesCount - collapsedHighlightRangesCount

    return {
        expandedMatchGroups,
        collapsedMatchGroups,
        hiddenMatchesCount,
    }
}

export function simplifyLineRange(range: Range): [number, number] {
    return [range.start.column, range.end.column]
}
