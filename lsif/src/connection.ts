import { Connection, createConnection as _createConnection } from 'typeorm'
import { entities } from './models.xrepo'
import { PostgresConnectionCredentialsOptions } from 'typeorm/driver/postgres/PostgresConnectionCredentialsOptions'
import { readEnvInt } from './util'
import { Configuration } from './config'

/**
 * The minimum migration version required by this instance of the LSIF process.
 * This should be updated any time a new lsif-server migration is added to the
 * migrations/ directory, as we watch the DB to ensure we're on at least this
 * version prior to making use of the DB (which the frontend may still be
 * migrating).
 */
const MINIMUM_MIGRATION_VERSION = 1528395595

/**
 * How long to wait between queries to check the current database migration version on startup.
 */
const SCHEMA_POLL_INTERVAL = readEnvInt('SCHEMA_POLL_INTERVAL', 5)

/**
 * How many times to try to connect to the cross-repository database on startup.
 */
const MAX_CONNECTION_RETRIES = readEnvInt('MAX_CONNECTION_RETRIES', 60)

/**
 * How long to wait between cross-repository connection attempts (in seconds).
 */
const CONNECTION_RETRY_INTERVAL = readEnvInt('CONNECTION_RETRY_INTERVAL', 5)

/**
 * Create a SQLite connection from the given filename.
 *
 * @param database The database filename.
 * @param entities The set of expected entities present in this schema.
 */
export function createSqliteConnection(
    database: string,
    // Decorators are not possible type check
    // eslint-disable-next-line @typescript-eslint/ban-types
    entities: Function[]
): Promise<Connection> {
    return _createConnection({
        type: 'sqlite',
        name: database,
        database,
        entities,
        synchronize: true,
        logging: ['error', 'warn'],
        maxQueryExecutionTime: 1000,
    })
}

/**
 * Create a Postgres connection. This creates a typorm connection pool
 * with the name `xrepo`. The connection configuration is constructed by
 * `createPostgresConnectionOptions`. This method blocks (failing after
 * a configured time) until the connection is established, then blocks
 * indefinitely while the database migration state is behind the
 * expected minimum, or dirty.
 *
 * @param configuration The current configuration.
 */
export async function createPostgresConnection(configuration: Configuration): Promise<Connection> {
    // Parse current PostgresDSN into connection options usable by
    // the typeorm postgres adapter.
    const url = new URL(configuration.postgresDSN)
    const connectionOptions = {
        host: url.hostname,
        port: parseInt(url.port, 10) || 5432,
        username: url.username,
        password: url.password,
        database: url.pathname.substring(1),
        ssl: url.searchParams.get('sslmode') === 'disable' ? false : undefined,
    }

    // Override the database name we're connecting to
    const connection = await connect({
        ...connectionOptions,
        database: connectionOptions.database + '_lsif',
    })

    // Poll the schema migrations table until we are up to date
    await waitForMigrations(connection, connectionOptions.database || '')

    return connection
}

/**
 * Create a connection to the cross-repository database. This will re-attempt to
 * access the database while the database does not exist. This is to give some
 * time to the frontend to run the migrations that create the LSIF database. The
 * retry interval and attempt count can be tuned via `MAX_CONNECTION_RETRIES` and
 * `CONNECTION_RETRY_INTERVAL` environment variables.
 *
 * @param connectionOptions The connection options.
 */
async function connect(connectionOptions: PostgresConnectionCredentialsOptions): Promise<Connection> {
    for (let attempts = 0; ; attempts++) {
        try {
            return await _createConnection({
                type: 'postgres',
                name: 'xrepo',
                entities,
                logging: ['error', 'warn'],
                maxQueryExecutionTime: 1000,
                ...connectionOptions,
            })
        } catch (error) {
            // This error occurs when no such database exists. Search for
            // invalid_catalog_name in:
            // https://www.postgresql.org/docs/9.4/errcodes-appendix.html.

            if (error && error.code === '3D000' && attempts + 1 < MAX_CONNECTION_RETRIES) {
                // snooze for a bit then retry
                await new Promise(resolve => setTimeout(resolve, CONNECTION_RETRY_INTERVAL * 1000))
                continue
            }

            throw error
        }
    }
}

/**
 * Block until we can select a migration version from the frontend database that
 * is at leasts as large as our minimum migration version.
 *
 * @param connection The connection to use.
 * @param database The target database in which to perform the query.
 */
async function waitForMigrations(connection: Connection, database: string): Promise<void> {
    while (true) {
        try {
            // Get migration version from frontend database
            const currentVersion = await getMigrationVersion(connection, database)

            // Check to see if the current version is at least the minimum version
            if (parseInt(currentVersion, 10) >= MINIMUM_MIGRATION_VERSION) {
                return
            }

            console.log(`waiting for migrations to be applied (${currentVersion} < ${MINIMUM_MIGRATION_VERSION})`)
        } catch (error) {
            console.log('failed to determine current database migration state', error)
        }

        // snooze for a bit then retry
        await new Promise(resolve => setTimeout(resolve, SCHEMA_POLL_INTERVAL * 1000))
    }
}

/**
 * Gets the current migration version from the frontend database. Throws on query
 * error, if no migration version can be found, or if the current migration state
 * is dirty.
 *
 * This process was configured to point to the primary Sourcegraph database, but we
 * have connected to the LSIF-specific database. We ue dblink to issue a query in the
 * 'remote' database without creating a second connection.
 *
 * @param connection The database connection.
 * @param database The target database in which to perform the query.
 */
async function getMigrationVersion(connection: Connection, database: string): Promise<string> {
    const query = `
        select * from
        dblink('dbname=' || $1 || ' user=' || current_user, 'select * from schema_migrations')
        as temp(version text, dirty bool);
    `

    const rows = (await connection.query(query, [database])) as {
        version: string
        dirty: boolean
    }[]

    if (rows.length > 0 && !rows[0].dirty) {
        return rows[0].version
    }

    throw new Error('Unusable migration state.')
}
