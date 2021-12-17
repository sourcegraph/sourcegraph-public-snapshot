import { MatchGroup, MatchGroupMatch, MatchItem, PerFileResultRanking, RankingResult } from './PerFileResultRanking'

/**
 * ZoektRanking preserves the original relevance that's computed by Zoekt.
 */
export class ZoektRanking implements PerFileResultRanking {
    public expandedResults(matches: MatchItem[], context: number): RankingResult {
        return results(matches, Number.MAX_VALUE, context)
    }
    public collapsedResults(matches: MatchItem[], context: number): RankingResult {
        return results(matches, 1, context)
    }
}

interface Hunk {
    matches: MatchItem[]
    startLine: number
    endLine: number
}

function newHunk(): Hunk {
    return { matches: [], startLine: Number.MAX_VALUE, endLine: Number.MIN_VALUE }
}

function addHunkItem(hunk: Hunk, item: MatchItem): void {
    hunk.matches.push(item)
    hunk.startLine = Math.min(item.line, hunk.startLine)
    hunk.endLine = Math.max(item.line, hunk.endLine)
}

function sortHunkMatches(hunk: Hunk): void {
    hunk.matches.sort((a, b) => {
        if (a.line < b.line) {
            return -1
        }
        if (a.line === b.line) {
            if (a.highlightRanges[0].start < b.highlightRanges[0].start) {
                return -1
            }
            if (a.highlightRanges[0].start === b.highlightRanges[0].start) {
                return 0
            }
        }
        return 1
    })
}

function isMatchWithinGroup(group: Hunk, item: MatchItem, context: number): boolean {
    return item.line + context >= group.startLine - context && item.line - context <= group.endLine + context
}

function results(matches: MatchItem[], maxResults: number, context: number): RankingResult {
    let hunks: Hunk[] = []
    for (const match of matches) {
        let isMergedWithExistingGroup = false
        for (const existingGroup of hunks) {
            if (isMatchWithinGroup(existingGroup, match, context)) {
                addHunkItem(existingGroup, match)
                isMergedWithExistingGroup = true
                break
            }
        }
        if (!isMergedWithExistingGroup) {
            const hunk = newHunk()
            addHunkItem(hunk, match)
            hunks.push(hunk)
        }
    }
    const groups: MatchGroup[] = []
    const visibleMatches: MatchItem[] = []
    hunks = hunks.slice(0, maxResults)
    for (const hunk of hunks) {
        sortHunkMatches(hunk)
        const groupMatches: MatchGroupMatch[] = []
        for (const match of hunk.matches) {
            visibleMatches.push(match)
            for (const range of match.highlightRanges) {
                groupMatches.push({
                    line: match.line,
                    character: range.start,
                    highlightLength: range.highlightLength,
                    isInContext: false,
                })
            }
        }
        const topGroupMatch = groupMatches[0]

        const matchGroup: MatchGroup = {
            matches: groupMatches,
            startLine: Math.max(0, hunk.startLine - context),
            endLine: hunk.endLine + context + 1,
            // 1-based position describing the starting place of the matches.
            position: { line: topGroupMatch.line + 1, character: topGroupMatch.character + 1 },
        }
        groups.push(matchGroup)
    }

    return { matches: visibleMatches, grouped: groups }
}
