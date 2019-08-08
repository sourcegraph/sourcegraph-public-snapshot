import * as lsp from 'vscode-languageserver'
import { CreateRunnerStats, InsertStats, QueryStats } from './stats'

export const ERRNOLSIFDATA = 'NoLSIFData'

/**
 * The exception thrown from `createRunner` when
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
export interface Backend<T extends QueryRunner> {
    /**
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `createRunner` method.
     */
    insertDump(
        tempPath: string,
        repository: string,
        commit: string,
        contentLength: number
    ): Promise<{ insertStats: InsertStats }>

    /**
     * Lists the query methods available from this backend.
     */
    availableQueries(): string[]

    /**
     * Create a query runner relevant to the given repository and commit hash. This
     * assumes that data for this subset of data has already been inserted via
     * `insertDump` (otherwise this method is expected to throw).
     */
    createRunner(repository: string, commit: string): Promise<{ queryRunner: T; createRunnerStats: CreateRunnerStats }>

    /**
     * Free any resources used by this object.
     */
    close(): Promise<void>
}

/**
 * A query runner is created by a `Backend` and can answer LSIF queries.
 */
export interface QueryRunner {
    /**
     * Determines whether or not data exists for the given file.
     */
    exists(file: string): Promise<boolean>

    /**
     * Return data for an LSIF query.
     */
    query(method: string, uri: string, position: lsp.Position): Promise<{ result: any; queryStats: QueryStats }>

    /**
     * Free any resources used by this object.
     */
    close(): Promise<void>
}
