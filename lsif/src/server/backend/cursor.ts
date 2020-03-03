/**
 * Reference pagination happens in two distinct phases:
 *
 *   (1) open a slice of dumps for the same repositories, and
 *   (2) open a slice of dumps for other repositories.
 */
export type ReferencePaginationPhase = 'same-repo' | 'remote-repo'

/**
 * Context describing the previous page of results.
 */
export interface ReferencePaginationCursor {
    /**
     * The identifier of the dump that contains the target range.
     */
    dumpId: number

    /**
     * The scheme of the moniker that has remote results.
     */
    scheme: string

    /**
     * The identifier of the moniker that has remote results.
     */
    identifier: string

    /**
     * The name of the package that has remote results.
     */
    name: string

    /**
     * The version of the package that has remote results.
     */
    version: string | null

    /**
     * The phase of the pagination.
     */
    phase: ReferencePaginationPhase

    /**
     * The number of remote dumps to skip.
     */
    offset: number
}

/**
 * Context describing the current request for paginated results.
 */
export interface ReferencePaginationContext {
    /**
     * The maximum number of remote dumps to search.
     */
    limit: number

    /**
     * Context describing the previous page of results.
     */
    cursor?: ReferencePaginationCursor
}
