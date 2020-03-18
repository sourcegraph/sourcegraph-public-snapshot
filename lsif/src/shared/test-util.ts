import * as constants from './constants'
import * as fs from 'mz/fs'
import * as nodepath from 'path'
import * as uuid from 'uuid'
import * as pgModels from './models/pg'
import { child_process } from 'mz'
import { Connection } from 'typeorm'
import { connectPostgres } from './database/postgres'
import { ensureDirectory } from './paths'
import { userInfo } from 'os'
import { DumpManager } from './store/dumps'
import { createSilentLogger } from './logging'

/** Create a temporary directory with a subdirectory for dbs. */
export async function createStorageRoot(): Promise<string> {
    const tempPath = await fs.mkdtemp('test-', { encoding: 'utf8' })
    await ensureDirectory(nodepath.join(tempPath, constants.DBS_DIR))
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
    const database = `sourcegraph-test-lsif-${suffix}`

    // Determine the path of the migrate script. This will cover the case
    // where `yarn test` is run from within the root or from the lsif directory.
    // const migrateScriptPath = nodepath.join((await fs.exists('dev')) ? '' : '..', 'dev', 'migrate.sh')
    const migrationsPath = nodepath.join((await fs.exists('migrations')) ? '' : '..', 'migrations')

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
    // dropdb: database removal failed: ERROR:  database "sourcegraph-test-lsif-5033c9e8" is being accessed by other users

    let connection: Connection
    const cleanup = async (): Promise<void> => {
        if (connection) {
            await connection.close()
        }

        await child_process.exec(dropCommand, { env }).then(() => undefined)
    }

    // Try to create database
    await child_process.exec(createCommand, { env })

    try {
        // Run migrations then connect to database
        await child_process.exec(migrateCommand, { env })
        connection = await connectPostgres(
            { host, port, username, password, database, ssl: false },
            suffix,
            createSilentLogger()
        )
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
 * Insert an upload entity and return the corresponding dump entity.
 *
 * @param connection The Postgres connection.
 * @param dumpManager The dumps manager instance.
 * @param repositoryId The repository identifier.
 * @param commit The commit.
 * @param root The root of all files in the dump.
 * @param indexer The type of indexer used to produce this dump.
 */
export async function insertDump(
    connection: Connection,
    dumpManager: DumpManager,
    repositoryId: number,
    commit: string,
    root: string,
    indexer: string
): Promise<pgModels.LsifDump> {
    await dumpManager.deleteOverlappingDumps(repositoryId, commit, root, indexer, {})

    const upload = new pgModels.LsifUpload()
    upload.repositoryId = repositoryId
    upload.commit = commit
    upload.root = root
    upload.indexer = indexer
    upload.filename = '<test>'
    upload.uploadedAt = new Date()
    upload.state = 'completed'
    upload.tracingContext = '{}'
    await connection.createEntityManager().save(upload)

    const dump = new pgModels.LsifDump()
    dump.id = upload.id
    dump.repositoryId = repositoryId
    dump.commit = commit
    dump.root = root
    dump.indexer = indexer
    return dump
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
