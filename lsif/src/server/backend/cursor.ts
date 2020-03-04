import * as sqliteModels from '../../shared/models/sqlite'

/** Context describing the current request for paginated results. */
export interface ReferencePaginationContext {
    /** The maximum number of remote dumps to search. */
    limit: number

    /** Context describing the next page of results. */
    cursor?: ReferencePaginationCursor
}

/** Context describing the next page of results. */
export type ReferencePaginationCursor =
    | SameDumpReferenceCursor
    | DefinitionMonikersReferenceCursor
    | RemoteDumpReferenceCursor

/** The cursor phase is a tag that indicates the shape of the object. */
export type ReferencePaginationPhase = 'same-dump' | 'definition-monikers' | 'same-repo' | 'remote-repo'

/** Fields common to all reference pagination cursors. */
interface ReferencePaginationCursorCommon {
    /** The identifier of the dump that contains the target range. */
    dumpId: number

    /** The phase of the pagination. */
    phase: ReferencePaginationPhase
}

/** Bookkeeping data for the part of the reference result set that deals with the initial dump. */
export interface SameDumpReferenceCursor extends ReferencePaginationCursorCommon {
    phase: 'same-dump'

    /** The (database-relative) document path containing the symbol ranges. */
    path: string

    /** A normalized list of monikers attached to the symbol ranges. */
    monikers: sqliteModels.MonikerData[]

    /** The number of monikers to skip processing. */
    skipMonikers: number

    /** The number of location results to skip for the current moniker. */
    skipResults: number
}

/** Bookkeeping data for the part of the reference result set that deals with the dumps that defines a moniker. */
export interface DefinitionMonikersReferenceCursor extends ReferencePaginationCursorCommon {
    phase: 'definition-monikers'

    /** The (database-relative) document path containing the symbol ranges. */
    path: string

    /** A normalized list of monikers attached to the symbol ranges. */
    monikers: sqliteModels.MonikerData[]

    /** The number of location results to skip for the current moniker. */
    skipResults: number
}

/** Bookkeeping data for the part of the reference result set that deals with additional dumps. */
export interface RemoteDumpReferenceCursor extends ReferencePaginationCursorCommon {
    phase: 'same-repo' | 'remote-repo'

    /** The identifier of the moniker that has remote results. */
    identifier: string

    /** The scheme of the moniker that has remote results. */
    scheme: string

    /** The name of the package that has remote results. */
    name: string

    /** The version of the package that has remote results. */
    version: string | null

    /** The number of dump batches we have already completed. */
    skipReferences: number

    /** The number of dumps we have already completed in the current batch. */
    skipDumps: number

    /** The number of location results to skip for the current dump. */
    skipResults: number
}
