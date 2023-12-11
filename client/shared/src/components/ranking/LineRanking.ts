import type { MatchGroup, PerFileResultRanking } from './PerFileResultRanking'

/**
 * LineRanking orders hunks purely by line number, disregarding the relevance ranking provided by Zoekt.
 */
export class LineRanking implements PerFileResultRanking {
    constructor(private maxMatches: number) {}

    public collapsedResults(groups: MatchGroup[]): MatchGroup[] {
        return sortMatchGroups(groups, this.maxMatches)
    }

    public expandedResults(groups: MatchGroup[]): MatchGroup[] {
        return sortMatchGroups(groups, Number.MAX_VALUE)
    }
}

export const sortMatchGroups = (groups: MatchGroup[], maxMatches: number): MatchGroup[] => {
    groups.sort((a, b) => {
        if (a.startLine < b.startLine) {
            return -1
        }
        if (a.startLine > b.startLine) {
            return 1
        }
        return 0
    })

    const visibleGroups = []
    let visibleMatches = 0
    for (const group of groups) {
        if (visibleMatches > maxMatches) {
            break
        }
        visibleGroups.push(group)
        visibleMatches += 1
    }

    return visibleGroups
}
