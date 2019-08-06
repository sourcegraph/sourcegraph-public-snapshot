import * as lsp from 'vscode-languageserver';
import { Database } from './ms/database';

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
    createDB(tempPath: string, key: string, contentLength: number): Promise<void>

    /**
     * Create a database instance from the given key. This assumes that the
     * database has been already created via a call to `createDB` (or this
     * method will otherwise fail).
     */
    loadDB(key: string): Promise<Database>

    /**
     * Return data for an LSIF hover query.
     */
    hover(db: Database, uri: string, position: lsp.Position): lsp.Hover | undefined

    /**
     * Return data for an LSIF definitions query.
     */
    definitions(db: Database, uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined

    /**
     * Return data for an LSIF references query.
     */
    references(
        db: Database,
        uri: string,
        position: lsp.Position,
        context: lsp.ReferenceContext
    ): lsp.Location[] | undefined
}
