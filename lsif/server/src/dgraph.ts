import * as lsp from 'vscode-languageserver'
import { Backend, QueryRunner } from './backend'
import { InsertStats, QueryStats, CreateRunnerStats } from './stats'

/**
 * Backend for SQLite dumps stored in Dgraph.
 */
export class DgraphBackend implements Backend<DgraphQueryRunner> {
    /**
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `createRunner` method.
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
     * Create a query runner relevant to the given repository and commit hash. This
     * assumes that data for this subset of data has already been inserted via
     * `insertDump` (otherwise this method is expected to throw).
     */
    public createRunner(
        repository: string,
        commit: string
    ): Promise<{ queryRunner: DgraphQueryRunner; createRunnerStats: CreateRunnerStats }> {
        // TODO(chris) - implement
        // MUST reject if `key` doesn't exist
        return Promise.reject()
    }
}

export class DgraphQueryRunner implements QueryRunner {
    /**
     * Determines whether or not data exists for the given file.
     */
    public exists(file: string): Promise<boolean> {
        // TODO(chris) - implement
        return Promise.reject()
    }

    /**
     * Return data for an LSIF query.
     */
    public query(
        method: string,
        uri: string,
        position: lsp.Position
    ): Promise<{ result: any; queryStats: QueryStats }> {
        // TODO(chris) - implement
        return Promise.reject()
    }

    /**
     * Free any resources used by this object.
     */
    public close(): Promise<void> {
        // TODO(chris) - implement
        return Promise.reject()
    }
}
