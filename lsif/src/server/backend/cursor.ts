import * as sqliteModels from '../../shared/models/sqlite'

/** Context describing the current request for paginated results. */
export interface ReferencePaginationContext {
    /** The maximum number of remote dumps to search. */
    limit: number

    /** Context describing the next page of results. */
    cursor?: ReferencePaginationCursor
}

/** Context describing the next page of results. */
export type ReferencePaginationCursor = SameDumpReferenceCursor | RemoteDumpReferenceCursor

/** The cursor phase is a tag that indicates the shape of the object. */
export type ReferencePaginationPhase = 'same-dump' | 'same-dump-monikers' | 'same-repo' | 'remote-repo'

/** TODO */
interface ReferencePaginationCursorCommon {
    /** The identifier of the dump that contains the target range. */
    dumpId: number

    /** The phase of the pagination. */
    phase: ReferencePaginationPhase
}

/** TODO */
export interface SameDumpReferenceCursor extends ReferencePaginationCursorCommon {
    phase: 'same-dump' | 'same-dump-monikers'

    /** TODO */
    path: string

    /** TODO */
    monikers: sqliteModels.MonikerData[]
}

/** TODO */
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

    /** The number of dumps to skip. */
    offset: number
}

/** TODO */
export function makeInitialSameDumpCursor(args: {
    dumpId: number
    path: string
    monikers: sqliteModels.MonikerData[]
}): ReferencePaginationCursor {
    return { phase: 'same-dump', ...args }
}

/** TODO */
export function makeInitialSameDumpMonikersCursor(previousCursor: SameDumpReferenceCursor): ReferencePaginationCursor {
    return { ...previousCursor, phase: 'same-dump-monikers' }
}

/** TODO */
export function makeInitialSameRepoCursor(
    previousCursor: SameDumpReferenceCursor,
    { scheme, identifier }: sqliteModels.MonikerData,
    { name, version }: sqliteModels.PackageInformationData
): ReferencePaginationCursor {
    return {
        ...previousCursor,
        phase: 'same-repo',
        scheme,
        identifier,
        name,
        version,
        offset: 0,
    }
}

/** TODO */
export function makeInitialRemoteRepoCursor(previousCursor: RemoteDumpReferenceCursor): ReferencePaginationCursor {
    return { ...previousCursor, phase: 'remote-repo', offset: 0 }
}
