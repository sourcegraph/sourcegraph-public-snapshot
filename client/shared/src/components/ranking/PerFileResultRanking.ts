export interface MatchItem {
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
}

/**
 * Describes a single group of matches.
 */
export interface MatchGroup {
    // The un-highlighted plain text for the lines in this group.
    plaintextLines: string[]

    // The highlighted HTML corresponding to plaintextLines.
    // The strings each contain a HTML <tr> containing the highlighted code.
    highlightedHTMLRows?: string[]

    // The matches in this group to display.
    matches: MatchGroupMatch[]

    // The 0-based start line of the group (inclusive.)
    startLine: number

    // The 0-based end line of the group (inclusive.)
    endLine: number
}

export interface MatchGroupMatch {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}

// rankPassthrough is a no-op re-ranker
export function rankPassthrough(groups: MatchGroup[]): MatchGroup[] {
    return groups
}

// rankByLine re-ranks a set of groups to order them by starting line number
export function rankByLine(groups: MatchGroup[]): MatchGroup[] {
    const groupsCopy = [...groups]
    groupsCopy.sort((a, b) => {
        if (a.startLine < b.startLine) {
            return -1
        }
        if (a.startLine > b.startLine) {
            return 1
        }
        return 0
    })
    return groupsCopy
}

export function truncateGroups(groups: MatchGroup[], maxMatches: number, contextLines: number): MatchGroup[] {
    const visibleGroups = []
    let remainingMatches = maxMatches
    for (const group of groups) {
        if (remainingMatches === 0) {
            break
        } else if (group.matches.length > remainingMatches) {
            visibleGroups.push(truncateGroup(group, remainingMatches, contextLines))
            break
        }
        visibleGroups.push(group)
        remainingMatches -= group.matches.length
    }

    return visibleGroups
}

function truncateGroup(group: MatchGroup, maxMatches: number, contextLines: number): MatchGroup {
    const keepMatches = group.matches.slice(0, maxMatches)
    const newStartLine = Math.max(
        Math.min(...keepMatches.map(match => match.startLine)) - contextLines,
        group.startLine
    )
    const newEndLine = Math.min(Math.max(...keepMatches.map(match => match.endLine)) + contextLines, group.endLine)
    const matchesInKeepContext = group.matches
        .slice(maxMatches)
        .filter(match => match.startLine >= newStartLine && match.endLine <= newEndLine)
    return {
        ...group,
        plaintextLines: group.plaintextLines.slice(newStartLine - group.startLine, newEndLine - group.startLine + 1),
        highlightedHTMLRows: group.highlightedHTMLRows?.slice(
            newStartLine - group.startLine,
            newEndLine - group.startLine + 1
        ),
        matches: [...keepMatches, ...matchesInKeepContext],
        startLine: newStartLine,
        endLine: newEndLine,
    }
}
