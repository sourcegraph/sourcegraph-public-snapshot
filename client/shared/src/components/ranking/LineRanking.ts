import { flatMap } from 'lodash'

import type { MatchGroup, MatchItem, PerFileResultRanking, RankingResult } from './PerFileResultRanking'

/**
 * LineRanking orders hunks purely by line number, disregarding the relevance ranking provided by Zoekt.
 */
export class LineRanking implements PerFileResultRanking {
    constructor(private maxMatches: number) {}

    public collapsedResults(matches: MatchItem[], context: number): RankingResult {
        return calculateMatchGroupsSorted(matches, this.maxMatches, context)
    }

    public expandedResults(matches: MatchItem[], context: number): RankingResult {
        return calculateMatchGroupsSorted(matches, 0, context)
    }
}

const calculateGroupPositions = (
    matches: {
        startLine: number
        startCharacter: number
        endLine: number
        endCharacter: number
    }[],
    context: number,
    highestLineNumberWithinSubsetMatches: number
): MatchGroup => {
    {
        let startLine = matches[0].startLine - context
        startLine = startLine < 0 ? 0 : startLine

        const highlightRangeEndLines = matches.map(range => range.endLine)
        // The highest line number of all highlights in this excerpt.
        const lastHighlightLineNumber = Math.max(...highlightRangeEndLines)

        // If the last highlight line is greater than the highest line number within the subset of matches
        // we are displaying, then we know that there's at least one highlight in the context lines.
        const contextLineHasHighlight = lastHighlightLineNumber > highestLineNumberWithinSubsetMatches

        // The gap between the last highlight provided to this excerpt, and the line number of the last highlighted
        // match that is not a context line. If this value is larger than context lines, it means that we are
        // displaying _all_ matches, and therefore, do not need to reduce the number of context lines shown.
        const remainingContextLinesToShow = lastHighlightLineNumber - highestLineNumberWithinSubsetMatches

        const numberOfContextLinesToShow = contextLineHasHighlight
            ? context - (remainingContextLinesToShow <= context ? remainingContextLinesToShow : 0)
            : context

        // Of the matches in this excerpt, pick the one with the highest line number + lines of context.
        // Don't add the context value to calculate the last line if the last highlight match is the highlight range + context
        const endLine = contextLineHasHighlight
            ? Math.max(...highlightRangeEndLines) + numberOfContextLinesToShow
            : Math.max(...highlightRangeEndLines) + context
        return {
            matches,

            // 1-based position describing the starting place of the matches.
            position: { line: matches[0].startLine + 1, character: matches[0].startCharacter + 1 },

            // 0-based range describing the start and end lines (end line is exclusive.)
            startLine,
            endLine: endLine + 1,
        }
    }
}

/**
 * Calculates how to group together matches for display. Takes into account:
 *
 * - Whether or not displaying a subset or all matches is desired
 * - A surrounding number of context lines to display
 * - Whether or not context lines have (or do not have) matches within them
 * - Grouping based on whether or not there is overlapping or adjacent context.
 *
 * @param matches The matches to split into groups.
 * @param maxMatches The maximum number of matches to show, or null for all.
 * @param context The number of surrounding context lines to show for each match.
 * @returns The subset of matches that were sorted and chosen for display, as well as that same
 * list of matches grouped together.
 */
export const calculateMatchGroupsSorted = (
    groups: MatchGroup[],
    maxMatches: number | null,
    context: number
): { matches: MatchItem[]; grouped: MatchGroup[] } => {
    const sortedGroups = groups.sort((a, b) => {
        // Groups should never be equal or o
        a.startLine < b.startLine ? -1 : 1
    })
    return { matches: showMatches, grouped }
}
