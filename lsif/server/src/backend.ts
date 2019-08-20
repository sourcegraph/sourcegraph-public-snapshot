import * as path from 'path'
import { fs } from 'mz'
import { Database } from './database'
import { hasErrorCode, readEnv } from './util'
import { convertToBlob, SymbolReference } from './importer'
import { XrepoDatabase } from './xrepo'
import { Readable } from 'stream'
import { BlobCache, ConnectionCache } from './cache'

export const ERRNOLSIFDATA = 'NoLSIFData'

/**
 * An error thrown when no LSIF database can be found on disk.
 */
export class NoLSIFDataError extends Error {
    public readonly name = ERRNOLSIFDATA
    public readonly code = ERRNOLSIFDATA

    constructor(repository: string, commit: string) {
        super(`No LSIF data available for ${repository}@${commit}.`)
    }
}

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = readEnv('LSIF_STORAGE_ROOT', 'lsif-storage')

/**
 * Backend for LSIF dumps stored in SQLite.
 */
export class SQLiteBackend {
    constructor(
        private xrepoDatabase: XrepoDatabase,
        private connectionCache: ConnectionCache,
        private blobCache: BlobCache
    ) {}

    /**
     * Read the content of the temporary file containing a JSON-encoded LSIF
     * dump. Insert these contents into some storage with an encoding that
     * can be subsequently read by the `createRunner` method.
     */
    public async insertDump(input: Readable, repository: string, commit: string): Promise<void> {
        const { exported, imported } = await convertToBlob(
            this.connectionCache,
            input,
            makeFilename(repository, commit)
        )

        for (const exportedPackage of exported) {
            await this.xrepoDatabase.addPackage(
                exportedPackage.scheme,
                exportedPackage.name,
                exportedPackage.version,
                repository,
                commit
            )
        }

        const identifiers: Map<
            string,
            {
                importedSymbol: SymbolReference
                ids: Set<string>
            }
        > = new Map()

        for (const importedSymbol of imported) {
            const hash = `${importedSymbol.scheme}::${importedSymbol.name}::${importedSymbol.version}`
            const result = identifiers.get(hash)
            if (result) {
                const { ids } = result
                ids.add(importedSymbol.identifier)
            } else {
                identifiers.set(hash, { importedSymbol, ids: new Set([importedSymbol.identifier]) })
            }
        }

        for (const { importedSymbol, ids } of identifiers.values()) {
            await this.xrepoDatabase.addReference(
                importedSymbol.scheme,
                importedSymbol.name,
                importedSymbol.version,
                repository,
                commit,
                Array.from(ids)
            )
        }
    }

    /**
     * Create a query runner relevant to the given repository and commit hash. This
     * assumes that data for this subset of data has already been inserted via
     * `insertDump` (otherwise this method is expected to throw).
     */
    public async createRunner(repository: string, commit: string): Promise<Database> {
        const file = makeFilename(repository, commit)

        try {
            await fs.stat(file)
        } catch (e) {
            if (hasErrorCode(e, 'ENOENT')) {
                throw new NoLSIFDataError(repository, commit)
            }

            throw e
        }

        return new Database(this.xrepoDatabase, this.connectionCache, this.blobCache, file)
    }
}

/**
 *.Computes the filename of the LSIF dump from the given repository and commit hash.
 */
export function makeFilename(repository: string, commit: string): string {
    return path.join(STORAGE_ROOT, `${encodeURIComponent(repository)}@${commit}.lsif.db`)
}

export async function makeBackend(connectionCache: ConnectionCache, blobCache: BlobCache): Promise<SQLiteBackend> {
    try {
        await fs.mkdir(STORAGE_ROOT)
    } catch (e) {
        if (!hasErrorCode(e, 'EEXIST')) {
            throw e
        }
    }

    const filename = path.join(STORAGE_ROOT, 'correlation.db')

    try {
        await fs.stat(filename)
    } catch (e) {
        if (!hasErrorCode(e, 'ENOENT')) {
            throw e
        }
    }

    return new SQLiteBackend(new XrepoDatabase(connectionCache, filename), connectionCache, blobCache)
}
