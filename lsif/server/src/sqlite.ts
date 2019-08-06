import * as lsp from 'vscode-languageserver'
import * as path from 'path'
import { Backend } from './backend'
import { BlobStore } from './ms/blobStore'
import { child_process, fs } from 'mz'
import { Database } from './ms/database'
import { GraphStore } from './ms/graphStore'
import { readEnvInt } from './env'

/**
 * Where on the file system to store LSIF files.
 */
const STORAGE_ROOT = process.env.LSIF_STORAGE_ROOT || 'lsif-storage'

/**
 * Soft limit on the amount of storage used by LSIF files. Storage can exceed
 * this limit if a single LSIF file is larger than this, otherwise storage will
 * be kept under this limit. Defaults to 100GB.
 */
const SOFT_MAX_STORAGE = readEnvInt({ key: 'LSIF_SOFT_MAX_STORAGE', defaultValue: 100 * 1024 * 1024 * 1024 })

/**
 * The path of the binary used to convert json dumps to sqlite dumps.
 * See https://github.com/microsoft/lsif-node/tree/master/sqlite.
 */
const SQLITE_CONVERTER_BINARY = './node_modules/lsif-sqlite/bin/lsif-sqlite'

/**
 * The abstract SQLite backend base that supports graph and blob subclasses.
 */
export abstract class SQLiteBackend implements Backend {
    /**
     * Re-encode the given file containing a JSON-encoded LSIF dump to the
     * proper format loadable by `loadDB`.
     */
    public async createDB(tempPath: string, key: string, contentLength: number): Promise<void> {
        await this.createStorageRoot()
        await child_process.exec(this.buildCommand(tempPath, key))
        this.cleanStorageRoot(Math.max(0, SOFT_MAX_STORAGE - contentLength))
    }

    /**
     * Create a database instance from the given key. This assumes that the
     * database has been already created via a call to `createDB` (or this
     * method will otherwise fail).
     */
    public async loadDB(key: string): Promise<Database> {
        const file = path.join(STORAGE_ROOT, key + '.db')

        const db = this.createStore()
        await db.load(file, projectRootURI => ({
            toDatabase: pathRelativeToProjectRoot => projectRootURI + '/' + pathRelativeToProjectRoot,
            fromDatabase: uri => (uri.startsWith(projectRootURI) ? uri.slice(`${projectRootURI}/`.length) : uri),
        }))

        return db
    }

    /**
     * Return data for an LSIF hover query.
     */
    public hover(db: Database, uri: string, position: lsp.Position): lsp.Hover | undefined {
        return db.hover(uri, position)
    }

    /**
     * Return data for an LSIF definitions query.
     */
    public definitions(db: Database, uri: string, position: lsp.Position): lsp.Location | lsp.Location[] | undefined {
        // TODO(efritz) - add multi-repo intel here
        return db.definitions(uri, position)
    }

    /**
     * Return data for an LSIF references query.
     */
    public references(
        db: Database,
        uri: string,
        position: lsp.Position,
        context: lsp.ReferenceContext
    ): lsp.Location[] | undefined {
        // TODO(efritz) - add multi-repo intel here
        return db.references(uri, position, context)
    }

    /**
     * Ensure the storage root directory exists.
     */
    private async createStorageRoot(): Promise<void> {
        try {
            await fs.mkdir(STORAGE_ROOT)
        } catch (e) {
            if (e.code === 'EEXIST') {
                return
            }

            throw e
        }
    }

    /**
     * Deletes old files (sorted by last modified time) to keep the disk usage below
     * the given `max`.
     */
    private async cleanStorageRoot(max: number): Promise<void> {
        const entries = await fs.readdir(STORAGE_ROOT)
        const files = await Promise.all(
            entries.map(async f => {
                const realPath = path.join(STORAGE_ROOT, f)
                return { path: realPath, stat: await fs.stat(realPath) }
            })
        )

        let totalSize = files.reduce((subtotal, f) => subtotal + f.stat.size, 0)

        // TODO - come up with a better fair-eviction policy so that one repo
        // with a  lot of dumps do not starve out disk space for other repos.

        for (const f of files.sort((a, b) => a.stat.atimeMs - b.stat.atimeMs)) {
            if (totalSize <= max) {
                return
            }

            console.log(`Deleting ${f.path} to help keep disk usage under ${SOFT_MAX_STORAGE}.`)
            await fs.unlink(f.path)
            totalSize = totalSize - f.stat.size
        }
    }

    /**
     * Build the command used to generate the SQLite dump from a temporary file.
     */
    protected abstract buildCommand(tempPath: string, key: string): string

    /**
     * Create a new, empty Database. This object should be able to load the file
     * created by `buildCommand`.
     */
    protected abstract createStore(): Database
}

/**
 * Backend for graph-encoded SQLite dumps.
 */
export class SQLiteGraphBackend extends SQLiteBackend {
    /**
     * Build the command used to generate the SQLite dump from a temporary file.
     */
    protected buildCommand(tempPath: string, key: string): string {
        const outFile = path.join(STORAGE_ROOT, key + '.db')
        return [SQLITE_CONVERTER_BINARY, '--in', tempPath, '--out', outFile].join(' ')
    }

    /**
     * Create a new, empty Database. This object should be able to load the file
     * created by `buildCommand`.
     */
    protected createStore(): Database {
        return new GraphStore()
    }
}

/**
 * Backend for blob-encoded SQLite dumps.
 */
export class SQLiteBlobBackend extends SQLiteBackend {
    /**
     * Build the command used to generate the SQLite dump from a temporary file.
     */
    protected buildCommand(tempPath: string, key: string): string {
        // TODO(efritz) - give this a meaningful value
        const projectVersion = '0'
        const outFile = path.join(STORAGE_ROOT, key + '.db')

        return [
            SQLITE_CONVERTER_BINARY,
            '--in',
            tempPath,
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
