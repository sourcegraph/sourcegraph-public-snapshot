import * as lsp from 'vscode-languageserver';
import { Database } from './ms/database';
import { InsertStats, GetHandleStats, QueryStats } from './stats';

export const ERRNOLSIFDATA = 'NoLSIFData'

/**
 * The exception thrown from `getDatabaseHandle` when
 */
export class NoLSIFDataError extends Error {
    public readonly name = ERRNOLSIFDATA
    public readonly code = ERRNOLSIFDATA

    constructor(repository: string, commit: string) {
        super(`No LSIF data available for ${repository}@${commit}.`)
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
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `getDatabaseHandle` method.
     */
    insertDump(
        tempPath: string,
        repository: string,
        commit: string,
        contentLength: number
    ): Promise<{ insertStats: InsertStats }>

    /**
     * Create a handle to the database relevant to the given repository and
     * commit hash.  This assumes that data for this database has already been
     * inserted via `insertDump` (otherwise this method is expected to throw).
     */
    getDatabaseHandle(repository: string, commit: string): Promise<{ database: Database; getHandleStats: GetHandleStats }>

    /**
     * Return data for an LSIF query.
     */
    query(
        db: Database,
        method: string,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: any; queryStats: QueryStats }>
}
