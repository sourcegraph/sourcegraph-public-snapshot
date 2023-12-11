import type { AggregableBadge, Badge as ExtensionBadgeType } from '../../codeintel/legacy-extensions/api'

/**
 * Interface for different ranking algorithms that determine how to display search results in the client.
 *
 * Determines only ranking of results for a local file.
 */
export interface PerFileResultRanking {
    /**
     * Returns the hunks that should be displayed by default before the user expands them
     */
    collapsedResults(groups: MatchGroup[]): MatchGroup[]
    /**
     * Returns the hunks that should be displayed after the user has explicitly requested to see all results.
     */
    expandedResults(groups: MatchGroup[]): MatchGroup[]
}

export interface MatchItem extends ExtensionBadgeType {
    highlightRanges: {
        startLine: number
        startCharacter: number
        endLine: number
        endCharacter: number
    }[]
    /**
     * The matched content, which contains all matched ranges in highlightRanges.
     */
    content: string
    /**
     * The 0-based line number where the matched content begins.
     */
    startLine: number
    /**
     * The 0-based line number where the matched content ends.
     */
    endLine: number
    aggregableBadges?: AggregableBadge[]
}

/**
 * Describes a single group of matches.
 */
export interface MatchGroup {
    blobLines: string[]

    // The matches in this group to display.
    matches: MatchGroupMatch[]

    // The 1-based position of where the first match in the group.
    position: {
        line: number
        character: number
    }

    // The 0-based start line of the group (inclusive.)
    startLine: number

    // The 0-based end line of the group (exclusive.)
    endLine: number
}

export interface MatchGroupMatch {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}
