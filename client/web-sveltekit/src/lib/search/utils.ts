import type { ContentMatch, MatchGroup, ChunkMatch, LineMatch, Filter } from '$lib/shared'

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

export function resultToMatchGroups(result: ContentMatch): MatchGroup[] {
    return result.chunkMatches?.map(chunkToMatchGroup) || result.lineMatches?.map(lineToMatchGroup) || []
}

interface FilterGroups {
    repo: Filter[]
    file: Filter[]
    lang: Filter[]
}

export function groupFilters(filters: Filter[] | null | undefined): FilterGroups {
    const groupedFilters: FilterGroups = {
        file: [],
        repo: [],
        lang: [],
    }
    if (filters) {
        for (const filter of filters) {
            switch (filter.kind) {
                case 'repo':
                case 'file':
                case 'lang': {
                    groupedFilters[filter.kind].push(filter)
                    break
                }
            }
        }
    }
    return groupedFilters
}

export function chunkToMatchGroup(chunk: ChunkMatch): MatchGroup {
    const matches = chunk.ranges.map(range => ({
        startLine: range.start.line,
        startCharacter: range.start.column,
        endLine: range.end.line,
        endCharacter: range.end.column,
    }))
    const plaintextLines = chunk.content.split(/\r?\n/)
    return {
        plaintextLines,
        highlightedHTMLRows: undefined, // populated lazily
        matches,
        startLine: chunk.contentStart.line,
        endLine: chunk.contentStart.line + plaintextLines.length,
    }
}

export function lineToMatchGroup(match: LineMatch): MatchGroup {
    const matches = match.offsetAndLengths.map(offsetAndLength => ({
        startLine: match.lineNumber,
        startCharacter: offsetAndLength[0],
        endLine: match.lineNumber,
        endCharacter: offsetAndLength[0] + offsetAndLength[1],
    }))
    return {
        plaintextLines: [match.line],
        highlightedHTMLRows: undefined, // populated lazily
        matches,
        startLine: match.lineNumber,
        endLine: match.lineNumber,
    }
}
