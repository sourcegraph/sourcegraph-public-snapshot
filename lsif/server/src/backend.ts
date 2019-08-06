import * as lsp from 'vscode-languageserver'
import { Database } from './ms/database'
import { EncodingStats, HandleStats, QueryStats } from './stats'

export const ERRNOLSIFDATA = 'NoLSIFData'

/**
 * The exception thrown from `loadDB` when
 */
export class NoLSIFDataError extends Error {
    public readonly name = ERRNOLSIFDATA
    public readonly code = ERRNOLSIFDATA

    constructor(key: string) {
        super(`No LSIF data available for ${key}.`)
    }
}

/**
 * Backend is the interface to the way an LSIF dump is encoded on disk,
 * in-memory, or in a separate process. This abstraction allows us to test
 * out several different experiments on how we encode and query data without
 * having to mess with all of the http-server plumbing.
 */
export interface Backend {
    /**
     * Re-encode the given file containing a JSON-encoded LSIF dump to the
     * proper format loadable by `loadDB`.
     */
    createDB(tempPath: string, key: string, contentLength: number): Promise<{ encodingStats: EncodingStats }>

    /**
     * Create a database instance from the given key. This assumes that the
     * database has been already created via a call to `createDB` (or this
     * method will otherwise fail).
     */
    loadDB(key: string): Promise<{ database: Database; handleStats: HandleStats }>

    /**
     * Return data for an LSIF hover query.
     */
    hover(
        db: Database,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: lsp.Hover | undefined; queryStats: QueryStats }>

    /**
     * Return data for an LSIF definitions query.
     */
    definitions(
        db: Database,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: lsp.Location | lsp.Location[] | undefined; queryStats: QueryStats }>

    /**
     * Return data for an LSIF references query.
     */
    references(
        db: Database,
        uri: string,
        position: lsp.Position,
        context: lsp.ReferenceContext
    ): Promise<{ result: lsp.Location[] | undefined; queryStats: QueryStats }>
}
