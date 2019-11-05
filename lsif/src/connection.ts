import { Connection, createConnection as _createConnection } from 'typeorm'
import * as xrepo from './xrepo.models'
import { PostgresConnectionCredentialsOptions } from 'typeorm/driver/postgres/PostgresConnectionCredentialsOptions'
import { readEnvInt } from './util'
import { Logger } from 'winston'
import { Configuration } from './config'
import pRetry from 'p-retry'
import { TlsOptions } from 'tls'

/**
 * The minimum migration version required by this instance of the LSIF process.
 * This should be updated any time a new lsif-server migration is added to the
 * migrations/ directory, as we watch the DB to ensure we're on at least this
 * version prior to making use of the DB (which the frontend may still be
 * migrating).
 */
const MINIMUM_MIGRATION_VERSION = 1528395607

/**
 * How many times to try to check the current database migration version on startup.
 */
const MAX_SCHEMA_POLL_RETRIES = readEnvInt('MAX_SCHEMA_POLL_RETRIES', 60)

/**
 * How long to wait (in seconds) between queries to check the current database migration version on startup.
 */
const SCHEMA_POLL_INTERVAL = readEnvInt('SCHEMA_POLL_INTERVAL', 5)

/**
 * How many times to try to connect to the cross-repository database on startup.
 */
const MAX_CONNECTION_RETRIES = readEnvInt('MAX_CONNECTION_RETRIES', 60)

/**
 * How long to wait (in seconds) between cross-repository connection attempts.
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
 * Create a Postgres connection. This creates a typorm connection pool with the
 * name `xrepo`. The connection configuration is constructed by the method
 * `createPostgresConnectionOptions`. This method blocks (failing after a configured
 * time) until the connection is established, then blocks indefinitely while the
 * database migration state is behind the expected minimum, or dirty.
 *
 * @param configuration The current configuration.
 * @param logger The logger instance.
 */
export async function createPostgresConnection(configuration: Configuration, logger: Logger): Promise<Connection> {
    // Parse current PostgresDSN into connection options usable by
    // the typeorm postgres adapter.
    const url = new URL(configuration.postgresDSN)

    // TODO(efritz) - handle allow, prefer, require, 'verify-ca', and 'verify-full'
    const sslModes: { [K: string]: boolean | TlsOptions } = {
        disable: false,
    }

    const host = url.hostname
    const port = parseInt(url.port, 10) || 5432
    const username = decodeURIComponent(url.username)
    const password = decodeURIComponent(url.password)
    const database = decodeURIComponent(url.pathname).substring(1) || username
    const sslMode = url.searchParams.get('sslmode')
    const ssl = sslMode ? sslModes[sslMode] : undefined

    // Get a working connection
    const connection = await connect(
        { host, port, username, password, database, ssl },
        logger
    )

    // Poll the schema migrations table until we are up to date
    await waitForMigrations(connection, logger)

    return connection
}

/**
 * Create a connection to the cross-repository database. This will re-attempt to
 * access the database while the database does not exist. This is to give some
 * time to the frontend to run the migrations that create the LSIF tables. The
 * retry interval and attempt count can be tuned via `MAX_CONNECTION_RETRIES` and
 * `CONNECTION_RETRY_INTERVAL` environment variables.
 *
 * @param connectionOptions The connection options.
 * @param logger The logger instance.
 */
function connect(connectionOptions: PostgresConnectionCredentialsOptions, logger: Logger): Promise<Connection> {
    return pRetry(
        () => {
            logger.debug('connecting to cross-repository database')
            return connectPostgres(connectionOptions, '')
        },
        {
            factor: 1,
            retries: MAX_CONNECTION_RETRIES,
            minTimeout: CONNECTION_RETRY_INTERVAL * 1000,
            maxTimeout: CONNECTION_RETRY_INTERVAL * 1000,
        }
    )
}

/**
 * Create a connection to the cross-repository database.
 *
 * @param connectionOptions The connection options.
 * @param suffix The database suffix (used for testing).
 */
export function connectPostgres(
    connectionOptions: PostgresConnectionCredentialsOptions,
    suffix: string
): Promise<Connection> {
    return _createConnection({
        type: 'postgres',
        name: `xrepo${suffix}`,
        entities: xrepo.entities,
        logging: ['error', 'warn'],
        maxQueryExecutionTime: 1000,
        ...connectionOptions,
    })
}

/**
 * Block until we can select a migration version from the database that is at
 * least as large as our minimum migration version.
 *
 * @param connection The connection to use.
 * @param logger The logger instance.
 */
function waitForMigrations(connection: Connection, logger: Logger): Promise<void> {
    const check = async (): Promise<void> => {
        logger.debug('checking database version', { requiredVersion: MINIMUM_MIGRATION_VERSION })

        const version = parseInt(await getMigrationVersion(connection), 10)
        if (isNaN(version) || version < MINIMUM_MIGRATION_VERSION) {
            throw new Error('cross-repository database not up to date')
        }
    }

    return pRetry(check, {
        factor: 1,
        retries: MAX_SCHEMA_POLL_RETRIES,
        minTimeout: SCHEMA_POLL_INTERVAL * 1000,
        maxTimeout: SCHEMA_POLL_INTERVAL * 1000,
    })
}

/**
 * Gets the current migration version from the frontend database. Throws on query
 * error, if no migration version can be found, or if the current migration state
 * is dirty.
 *
 * @param connection The database connection.
 */
async function getMigrationVersion(connection: Connection): Promise<string> {
    const rows = (await connection.query('select * from schema_migrations')) as {
        version: string
        dirty: boolean
    }[]

    if (rows.length > 0 && !rows[0].dirty) {
        return rows[0].version
    }

    throw new Error('Unusable migration state.')
}
