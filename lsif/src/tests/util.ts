import rmfr from 'rmfr'
import { Backend, ReferencePaginationCursor } from '../backend'
import { Configuration } from '../config'
import { createCleanPostgresDatabase, createStorageRoot, convertTestData } from '../test-utils'
import { XrepoDatabase } from '../xrepo'
import { lsp } from 'lsif-protocol'

/**
 * A wrapper around tests for the Backend class. This abstracts a lot
 * of the common setup and teardown around creating a temporary Postgres
 * database, a storage root, across-repository database instance, and a
 * backend instance.
 */
export class BackendTestContext {
    /**
     * The backend instance.
     */
    public backend?: Backend

    /**
     * The cross-repository database instance.
     */
    public xrepoDatabase?: XrepoDatabase

    /**
     * A temporary directory.
     */
    private storageRoot?: string

    /**
     * A reference to a function that destroys the temporary database.
     */
    private cleanup?: () => Promise<void>

    /**
     * Create a backend and a cross-repository database. This will create
     * temporary resources (database and temporary directory) that should
     * be cleaned up via the `teardown` method.
     *
     * The backend and cross-repository database values can be referenced
     * by the public fields of this class.
     */
    public async init(): Promise<void> {
        this.storageRoot = await createStorageRoot()

        const { connection, cleanup } = await createCleanPostgresDatabase()
        this.cleanup = cleanup

        this.xrepoDatabase = new XrepoDatabase(this.storageRoot, connection)
        this.backend = new Backend(this.storageRoot, this.xrepoDatabase, () => ({} as Configuration))
    }

    /**
     * Mock an upload of the given file. This will create a SQLite database in the
     * given storage root and will insert dump, package, and reference data into
     * the given cross-repository database.
     *
     * @param repository The repository name.
     * @param commit The commit.
     * @param root The root of the dump.
     * @param filename The filename of the (gzipped) LSIF dump.
     * @param updateCommits Whether not to update commits.
     */
    public convertTestData(
        repository: string,
        commit: string,
        root: string,
        filename: string,
        updateCommits: boolean = true
    ): Promise<void> {
        if (!this.xrepoDatabase || !this.storageRoot) {
            return Promise.resolve()
        }

        return convertTestData(this.xrepoDatabase, this.storageRoot, repository, commit, root, filename, updateCommits)
    }

    /**
     * Clean up disk and database resources created for this test.
     */
    public async teardown(): Promise<void> {
        if (this.storageRoot) {
            await rmfr(this.storageRoot)
        }

        if (this.cleanup) {
            await this.cleanup()
        }
    }
}

/**
 * Remove all node_modules locations from the output of a references result.
 *
 * @param args Parameter bag.
 */
export function filterNodeModules<T>({
    locations,
    cursor,
}: {
    /** The reference locations. */
    locations: lsp.Location[]
    /** The pagination cursor. */
    cursor?: ReferencePaginationCursor
}): { locations: lsp.Location[]; cursor?: ReferencePaginationCursor } {
    return { locations: locations.filter(l => !l.uri.includes('node_modules')), cursor }
}
