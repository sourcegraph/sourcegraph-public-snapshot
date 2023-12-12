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

export function resultToMatchItems(result: ContentMatch): MatchItem[] {
    return result.type === 'content'
        ? result.chunkMatches?.map(match => ({
              highlightRanges: match.ranges.map(range => ({
                  startLine: range.start.line,
                  startCharacter: range.start.column,
                  endLine: range.end.line,
                  endCharacter: range.end.column,
              })),
              content: match.content,
              startLine: match.contentStart.line,
              endLine: match.ranges.at(-1)?.end.line ?? match.contentStart.line,
          })) ||
              result.lineMatches?.map(match => ({
                  highlightRanges: match.offsetAndLengths.map(offsetAndLength => ({
                      startLine: match.lineNumber,
                      startCharacter: offsetAndLength[0],
                      endLine: match.lineNumber,
                      endCharacter: offsetAndLength[0] + offsetAndLength[1],
                  })),
                  content: match.line,
                  startLine: match.lineNumber,
                  endLine: match.lineNumber,
              })) ||
              []
        : []
}
