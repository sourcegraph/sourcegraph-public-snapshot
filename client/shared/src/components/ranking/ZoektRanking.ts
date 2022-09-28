import { MatchGroup, MatchGroupMatch, MatchItem, PerFileResultRanking, RankingResult } from './PerFileResultRanking'

/**
 * ZoektRanking preserves the original relevance that's computed by Zoekt.
 */
export class ZoektRanking implements PerFileResultRanking {
    constructor(private maxResults: number) {}

    public collapsedResults(matches: MatchItem[], context: number): RankingResult {
        return results(matches, this.maxResults, context)
    }

    public expandedResults(matches: MatchItem[], context: number): RankingResult {
        return results(matches, Number.MAX_VALUE, context)
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
    hunk.startLine = Math.min(item.startLine, hunk.startLine)
    hunk.endLine = Math.max(item.highlightRanges[item.highlightRanges.length - 1].endLine, hunk.endLine)
}

function sortHunkMatches(hunk: Hunk): void {
    hunk.matches.sort((a, b) => {
        if (a.startLine < b.startLine) {
            return -1
        }
        if (a.startLine === b.startLine) {
            if (a.highlightRanges[0].startCharacter < b.highlightRanges[0].startCharacter) {
                return -1
            }
            if (a.highlightRanges[0].startCharacter === b.highlightRanges[0].startCharacter) {
                return 0
            }
        }
        return 1
    })
}

function isMatchWithinGroup(group: Hunk, item: MatchItem, context: number): boolean {
    return (
        item.startLine + context + 1 >= group.startLine - context &&
        item.highlightRanges[item.highlightRanges.length - 1].endLine - context - 1 <= group.endLine + context
    )
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
                    startLine: range.startLine,
                    startCharacter: range.startCharacter,
                    endLine: range.endLine,
                    endCharacter: range.endCharacter,
                })
            }
        }
        const topGroupMatch = groupMatches[0]

        const matchGroup: MatchGroup = {
            matches: groupMatches,
            startLine: Math.max(0, hunk.startLine - context),
            endLine: hunk.endLine + context + 1,
            // 1-based position describing the starting place of the matches.
            position: { line: topGroupMatch.startLine + 1, character: topGroupMatch.startCharacter + 1 },
        }
        groups.push(matchGroup)
    }

    return { matches: visibleMatches, grouped: groups }
}
