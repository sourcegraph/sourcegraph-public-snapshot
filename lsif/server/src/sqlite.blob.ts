import { BlobStore } from './ms/server/blobStore'
import { Database } from './ms/server/database'
import { SQLiteBackend } from './sqlite'
import { convertToBlob } from './ms/sqlite/convert'

/**
 * Backend for blob-encoded SQLite dumps.
 */
export class SQLiteBlobBackend extends SQLiteBackend {
    /**
     * Generate a SQLite dump from a temporary file to the given target file.
     */
    protected async convert(inFile: string, outFile: string): Promise<void> {
        await convertToBlob(inFile, outFile)
    }

    /**
     * Create a new, empty Database. This object should be able to load the file
     * created by `buildCommand`.
     */
    protected createStore(): Database {
        return new BlobStore()
    }
}
