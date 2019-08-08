import { BlobStore } from './ms/blobStore'
import { Database } from './ms/database'
import { SQLiteBackend } from './sqlite'

/**
 * The path of the binary used to convert JSON dumps to SQLite dumps.
 * See https://github.com/microsoft/lsif-node/tree/master/sqlite.
 */
const SQLITE_CONVERTER_BINARY = './node_modules/lsif-sqlite/bin/lsif-sqlite'

/**
 * Backend for blob-encoded SQLite dumps.
 */
export class SQLiteBlobBackend extends SQLiteBackend {
    /**
     * Build the command used to generate the SQLite dump from a temporary file.
     */
    protected buildCommand(inFile: string, outFile: string): string {
        // TODO(efritz) - give this a meaningful value
        const projectVersion = '0'

        return [
            SQLITE_CONVERTER_BINARY,
            '--in',
            inFile,
            '--out',
            outFile,
            '--format',
            'blob',
            '--delete',
            '--projectVersion',
            projectVersion,
        ].join(' ')
    }

    /**
     * Create a new, empty Database. This object should be able to load the file
     * created by `buildCommand`.
     */
    protected createStore(): Database {
        return new BlobStore()
    }
}
