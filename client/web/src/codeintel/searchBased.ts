import type { Range } from '@sourcegraph/extension-api-types'

export interface Result {
    /** The name of the repository containing the result. */
    repo: string

    /** The commit containing the result. */
    rev: string

    /** The path to the result file relative to the repository root. */
    file: string

    /** The content of the file. */
    content: string

    /** The unique URL to this result. */

    url: string

    /** The range of the match. */
    range: Range

    /** The type of symbol, if the result came from a symbol search. */
    symbolKind?: string

    /**
     * Whether or not the symbol is local to the containing file, if
     * the result came from a symbol search.
     */
    fileLocal?: boolean
}
