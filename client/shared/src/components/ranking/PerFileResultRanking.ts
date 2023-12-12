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

    // The 1-based position of where the first match in the group.
    position: {
        line: number
        character: number
    }

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
