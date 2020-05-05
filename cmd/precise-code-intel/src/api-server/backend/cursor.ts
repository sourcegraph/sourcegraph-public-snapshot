import * as sqliteModels from '../../shared/models/sqlite'
import * as lsp from 'vscode-languageserver-protocol'

/** Context describing the current request for paginated results. */
export interface ReferencePaginationContext {
    /** The maximum number of locations to return on this page. */
    limit: number

    /** Context describing the next page of results. */
    cursor?: ReferencePaginationCursor
}

/** Context describing the next page of results. */
export type ReferencePaginationCursor =
    | SameDumpReferenceCursor
    | DefinitionMonikersReferenceCursor
    | RemoteDumpReferenceCursor

/** A label that indicates which pagination phase is being expanded. */
export type ReferencePaginationPhase = 'same-dump' | 'definition-monikers' | 'same-repo' | 'remote-repo'

/** Fields common to all reference pagination cursors. */
interface ReferencePaginationCursorCommon {
    /** The identifier of the dump that contains the target range. */
    dumpId: number

    /** The phase of the pagination. */
    phase: ReferencePaginationPhase
}

/** Bookkeeping data for the reference results that come from the initial dump. */
export interface SameDumpReferenceCursor extends ReferencePaginationCursorCommon {
    phase: 'same-dump'

    /** The (database-relative) document path containing the symbol ranges. */
    path: string

    /** The current hover position. */
    position: lsp.Position

    /** A normalized list of monikers attached to the symbol ranges. */
    monikers: sqliteModels.MonikerData[]

    /** The number of reference results to skip. */
    skipResults: number
}

/** Bookkeeping data for the reference results that come from dumps defining a moniker. */
export interface DefinitionMonikersReferenceCursor extends ReferencePaginationCursorCommon {
    phase: 'definition-monikers'

    /** The (database-relative) document path containing the symbol ranges. */
    path: string

    /** A normalized list of monikers attached to the symbol ranges. */
    monikers: sqliteModels.MonikerData[]

    /** The number of location results to skip for the current moniker. */
    skipResults: number
}

/** Bookkeeping data for the reference results that come from additional (remote) dumps. */
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

    /** The current batch of dumps to open. */
    dumpIds: number[]

    /** The total count of candidate dumps that can be opened. */
    totalDumpsWhenBatching: number

    /** The number of dumps we have already processed or bloom filtered. */
    skipDumpsWhenBatching: number

    /** The number of dumps we have already completed in the current batch. */
    skipDumpsInBatch: number

    /** The number of location results to skip for the current dump. */
    skipResultsInDump: number
}
