import type { MatchGroup, PerFileResultRanking } from './PerFileResultRanking'

/**
 * ZoektRanking preserves the original relevance that's computed by Zoekt.
 */
export class ZoektRanking implements PerFileResultRanking {
    constructor(private maxResults: number) {}

    public collapsedResults(groups: MatchGroup[]): MatchGroup[] {
        return results(groups, this.maxResults)
    }

    public expandedResults(groups: MatchGroup[]): MatchGroup[] {
        return results(groups, Number.MAX_VALUE)
    }
}

function results(groups: MatchGroup[], maxResults: number): MatchGroup[] {
    const visibleGroups = []
    let visibleMatches = 0
    for (const group of groups) {
        if (visibleMatches > maxResults) {
            break
        }
        visibleGroups.push(group)
        visibleMatches += 1
    }

    return visibleGroups
}
