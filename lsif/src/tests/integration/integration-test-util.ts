import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as uuid from 'uuid'
import rmfr from 'rmfr'
import { Backend, ReferencePaginationCursor } from '../../server/backend/backend'
import { child_process } from 'mz'
import { Connection } from 'typeorm'
import { connectPostgres } from '../../shared/database/postgres'
import { convertLsif } from '../../worker/importer/importer'
import { dbFilename, ensureDirectory } from '../../shared/paths'
import { lsp } from 'lsif-protocol'
import { userInfo } from 'os'
import { XrepoDatabase } from '../../shared/xrepo/xrepo'

/**
 * Create a temporary directory with a subdirectory for dbs.
 */
export async function createStorageRoot(): Promise<string> {
    const tempPath = await fs.mkdtemp('test-', { encoding: 'utf8' })
    await ensureDirectory(path.join(tempPath, constants.DBS_DIR))
    return tempPath
}

/**
 * Create a new postgres database with a random suffix, apply the frontend
 * migrations (via the ./dev/migrate.sh script) and return an open connection.
 * This uses the PG* environment variables for host, port, user, and password.
 * This also returns a cleanup function that will destroy the database, which
 * should be called at the end of the test.
 */
export async function createCleanPostgresDatabase(): Promise<{ connection: Connection; cleanup: () => Promise<void> }> {
    // Each test has a random dbname suffix
    const suffix = uuid.v4().substring(0, 8)

    // Pull test db config from environment
    const host = process.env.PGHOST || 'localhost'
    const port = parseInt(process.env.PGPORT || '5432', 10)
    const username = process.env.PGUSER || userInfo().username || 'postgres'
    const password = process.env.PGPASSWORD || ''
    const database = `sourcegraph-test-lsif-xrepo-${suffix}`

    // Determine the path of the migrate script. This will cover the case
    // where `yarn test` is run from within the root or from the lsif directory.
    // const migrateScriptPath = path.join((await fs.exists('dev')) ? '' : '..', 'dev', 'migrate.sh')
    const migrationsPath = path.join((await fs.exists('migrations')) ? '' : '..', 'migrations')

    // Ensure environment gets passed to child commands
    const env = {
        ...process.env,
        PGHOST: host,
        PGPORT: `${port}`,
        PGUSER: username,
        PGPASSWORD: password,
        PGSSLMODE: 'disable',
        PGDATABASE: database,
    }

    // Construct postgres connection string using environment above. We disable this
    // eslint rule because we want it to use bash interpolation, not typescript string
    // templates.
    //
    // eslint-disable-next-line no-template-curly-in-string
    const connectionString = 'postgres://${PGUSER}:${PGPASSWORD}@${PGHOST}:${PGPORT}/${PGDATABASE}?sslmode=disable'

    // Define command text
    const createCommand = `createdb ${database}`
    const dropCommand = `dropdb --if-exists ${database}`
    const migrateCommand = `migrate -database "${connectionString}" -path  ${migrationsPath} up`

    // Create cleanup function to run after test. This will close the connection
    // created below (if successful), then destroy the database that was created
    // for the test. It is necessary to close the database first, otherwise we
    // get failures during the after hooks:
    //
    // dropdb: database removal failed: ERROR:  database "sourcegraph-test-lsif-xrepo-5033c9e8" is being accessed by other users

    let connection: Connection
    const cleanup = async (): Promise<void> => {
        if (connection) {
            await connection.close()
        }

        await child_process.exec(dropCommand, { env }).then(() => {})
    }

    // Try to create database
    await child_process.exec(createCommand, { env })

    try {
        // Run migrations then connect to database
        await child_process.exec(migrateCommand, { env })
        connection = await connectPostgres({ host, port, username, password, database, ssl: false }, suffix)
        return { connection, cleanup }
    } catch (error) {
        // We made a database but can't use it - try to clean up
        // before throwing the original error.

        try {
            await cleanup()
        } catch (_) {
            // If a new error occurs, swallow it
        }

        // Throw the original error
        throw error
    }
}

/**
 * Truncate all tables that do not match `schema_migrations`.
 *
 * @param connection The connection to use.
 */
export async function truncatePostgresTables(connection: Connection): Promise<void> {
    const results: { table_name: string }[] = await connection.query(
        "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' AND table_name != 'schema_migrations'"
    )

    const tableNames = results.map(row => row.table_name).join(', ')
    await connection.query(`truncate ${tableNames} restart identity`)
}

/**
 * Mock an upload of the given file. This will create a SQLite database in the
 * given storage root and will insert dump, package, and reference data into
 * the given cross-repository database.
 *
 * @param xrepoDatabase The cross-repository database.
 * @param storageRoot The temporary storage root.
 * @param repository The repository name.
 * @param commit The commit.
 * @param root The root of the dump.
 * @param filename The filename of the (gzipped) LSIF dump.
 * @param updateCommits Whether not to update commits.
 */
export async function convertTestData(
    xrepoDatabase: XrepoDatabase,
    storageRoot: string,
    repository: string,
    commit: string,
    root: string,
    filename: string,
    updateCommits: boolean = true
): Promise<void> {
    // Create a filesystem read stream for the given test file. This will cover
    // the cases where `yarn test` is run from the root or from the lsif directory.
    const fullFilename = path.join((await fs.exists('lsif')) ? 'lsif' : '', 'src/tests/integration/data', filename)

    const tmp = path.join(storageRoot, constants.TEMP_DIR, uuid.v4())
    const { packages, references } = await convertLsif(fullFilename, tmp)
    const dump = await xrepoDatabase.addPackagesAndReferences(
        repository,
        commit,
        root,
        new Date(),
        packages,
        references
    )
    await fs.rename(tmp, dbFilename(storageRoot, dump.id, repository, commit))

    if (updateCommits) {
        await xrepoDatabase.updateCommits(repository, [[commit, undefined]])
        await xrepoDatabase.updateDumpsVisibleFromTip(repository, commit)
    }
}

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
        this.backend = new Backend(this.storageRoot, this.xrepoDatabase, () => ({ gitServers: [] }))
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
 * Create an LSP location.
 *
 * @param uri The document path.
 * @param startLine The starting line.
 * @param startCharacter The starting character.
 * @param endLine The ending line.
 * @param endCharacter The ending character.
 */
export function createLocation(
    uri: string,
    startLine: number,
    startCharacter: number,
    endLine: number,
    endCharacter: number
): lsp.Location {
    return lsp.Location.create(uri, {
        start: { line: startLine, character: startCharacter },
        end: { line: endLine, character: endCharacter },
    })
}

/**
 * Create an LSP location with a remote URI.
 *
 * @param repository The repository name.
 * @param commit The commit.
 * @param documentPath The document path.
 * @param startLine The starting line.
 * @param startCharacter The starting character.
 * @param endLine The ending line.
 * @param endCharacter The ending character.
 */
export function createRemoteLocation(
    repository: string,
    commit: string,
    documentPath: string,
    startLine: number,
    startCharacter: number,
    endLine: number,
    endCharacter: number
): lsp.Location {
    const url = new URL(`git://${repository}`)
    url.search = commit
    url.hash = documentPath

    return createLocation(url.href, startLine, startCharacter, endLine, endCharacter)
}

/** A counter used for unique commit generation. */
let commitBase = 0

/**
 * Create a 40-character commit.
 *
 * @param base A unique numeric base to repeat.
 */
export function createCommit(base?: number): string {
    if (base === undefined) {
        base = commitBase
        commitBase++
    }

    // Add 'a' to differentiate between similar numeric bases such as `1a1a...` and `11a11a...`.
    return (base + 'a').repeat(40).substring(0, 40)
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
