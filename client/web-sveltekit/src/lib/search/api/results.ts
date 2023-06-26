import type { Observable } from 'rxjs'

import type { HighlightResponseFormat, HighlightLineRange } from '$lib/graphql-operations'
import { fetchHighlightedFileLineRanges } from '$lib/loader/blob'
import type { ContentMatch, MatchItem } from '$lib/shared'
import { requestGraphQL } from '$lib/web'

import type { Result } from '../domain/result'

export function fetchFileRangeMatches(args: {
    result: Result
    format?: HighlightResponseFormat
    ranges: HighlightLineRange[]
}): Observable<string[][]> {
    return fetchHighlightedFileLineRanges(
        {
            repoName: args.result.repository,
            commitID: args.result.commit || '',
            filePath: args.result.path,
            disableTimeout: false,
            format: args.format,
            ranges: args.ranges,
            platformContext: { requestGraphQL: ({ request, variables }) => requestGraphQL(request, variables) },
        },
        false
    )
}

export function mapContentMatchToMatchItems(result: ContentMatch): MatchItem[] {
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
              endLine: match.ranges[match.ranges.length - 1].end.line,
              aggregableBadges: match.aggregableBadges,
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
                  aggregableBadges: match.aggregableBadges,
              })) ||
              []
        : []
}
