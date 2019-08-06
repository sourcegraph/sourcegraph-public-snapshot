import * as lsp from 'vscode-languageserver'
import { Backend } from './backend'
import { Database } from './ms/database'
import { EncodingStats, HandleStats, QueryStats } from './stats'

/**
 * Backend for SQLite dumps stored in Dgraph.
 */
export class DgraphBackend implements Backend {
    /**
     * Re-encode the given file containing a JSON-encoded LSIF dump to the
     * proper format loadable by `loadDB`.
     */
    public createDB(tempPath: string, key: string, contentLength: number): Promise<{ encodingStats: EncodingStats }> {
        // TODO(chris) - implement
        return Promise.reject()
    }

    /**
     * Create a database instance from the given key. This assumes that the
     * database has been already created via a call to `createDB` (or this
     * method will otherwise fail).
     */
    public loadDB(key: string): Promise<{ database: Database; handleStats: HandleStats }> {
        // TODO(chris) - implement
        // MUST reject if `key` doesn't exist
        return Promise.reject()
    }

    /**
     * Return data for an LSIF hover query.
     */
    public hover(
        db: Database,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: lsp.Hover | undefined; queryStats: QueryStats }> {
        // TODO(chris) - implement
        return Promise.reject()
    }

    /**
     * Return data for an LSIF definitions query.
     */
    public definitions(
        db: Database,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: lsp.Location | lsp.Location[] | undefined; queryStats: QueryStats }> {
        // TODO(chris) - implement
        return Promise.reject()
    }

    /**
     * Return data for an LSIF references query.
     */
    public references(
        db: Database,
        uri: string,
        position: lsp.Position,
        context: lsp.ReferenceContext
    ): Promise<{ result: lsp.Location[] | undefined; queryStats: QueryStats }> {
        // TODO(chris) - implement
        return Promise.reject()
    }
}
