import { Database } from './ms/server/database'
import { GraphStore } from './ms/server/graphStore'
import { SQLiteBackend } from './sqlite'
import { convertToGraph } from './ms/sqlite/convert'

/**
 * Backend for graph-encoded SQLite dumps.
 */
export class SQLiteGraphBackend extends SQLiteBackend {
    /**
     * Generate a SQLite dump from a temporary file to the given target file.
     */
    protected async convert(inFile: string, outFile: string): Promise<void> {
        await convertToGraph(inFile, outFile)
    }

    /**
     * Create a new, empty Database. This object should be able to load the file
     * created by `buildCommand`.
     */
    protected createStore(): Database {
        return new GraphStore()
    }
}
