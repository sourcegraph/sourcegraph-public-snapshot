import type {
    MatchGroup,
    MatchGroupMatch,
    MatchItem,
    PerFileResultRanking,
    RankingResult,
} from './PerFileResultRanking'

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
    hunk.endLine = Math.max(item.endLine, hunk.endLine)
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

function mergeHunks(hunks: Hunk[], context: number): void {
    const hunksSortedByLineNumber = hunks.slice().sort((a, b) => {
        if (a.startLine < b.startLine) {
            return -1
        }
        if (a.startLine === b.startLine) {
            return 0
        }
        return 1
    })

    for (let index = 0; index < hunksSortedByLineNumber.length - 1; index++) {
        const current = hunksSortedByLineNumber[index]
        const next = hunksSortedByLineNumber[index + 1]

        if (next.startLine - current.startLine <= context) {
            const originalHunkIndex = hunks.indexOf(current)
            const nextHunkIndex = hunks.indexOf(next)

            if (originalHunkIndex > -1 && nextHunkIndex > -1) {
                const originalHunk = hunks[originalHunkIndex]
                for (const match of next.matches) {
                    addHunkItem(originalHunk, match)
                }
                hunks.splice(nextHunkIndex, 1)
                index++
            }
        }
    }
}

function isMatchWithinGroup(group: Hunk, item: MatchItem, context: number): boolean {
    // Return true if item and group have overlapping or adjacent context
    return (
        (item.startLine >= group.endLine && item.startLine - group.endLine - 2 * context <= 1) ||
        (item.endLine <= group.startLine && group.startLine - item.endLine - 2 * context <= 1)
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

    // Merge hunks with overlapping or adjacent line ranges
    mergeHunks(hunks, context)

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
