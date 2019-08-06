import * as lsp from 'vscode-languageserver'
import { Backend } from './backend'
import { Database } from './ms/database'
import { GetHandleStats, InsertStats, QueryStats } from './stats'

/**
 * Backend for SQLite dumps stored in Dgraph.
 */
export class DgraphBackend implements Backend {
    /**
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `getDatabaseHandle` method.
     */
    public insertDump(
        tempPath: string,
        repository: string,
        commit: string,
        contentLength: number
    ): Promise<{ insertStats: InsertStats }> {
        // TODO(chris) - implement
        return Promise.reject()
    }

    /**
     * Create a handle to the database relevant to the given repository and
     * commit hash.  This assumes that data for this database has already been
     * inserted via `insertDump` (otherwise this method is expected to throw).
     */
    public getDatabaseHandle(
        repository: string,
        commit: string
    ): Promise<{ database: Database; getHandleStats: GetHandleStats }> {
        // TODO(chris) - implement
        // MUST reject if `key` doesn't exist
        return Promise.reject()
    }

    /**
     * Return data for an LSIF query.
     */
    public query(
        db: Database,
        method: string,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: any; queryStats: QueryStats }> {
        // TODO(chris) - implement
        return Promise.reject()
    }
}
