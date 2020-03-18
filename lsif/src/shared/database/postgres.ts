import * as metrics from './metrics'
import * as pgModels from '../models/pg'
import pRetry from 'p-retry'
import { Configuration } from '../config/config'
import { Connection, createConnection as _createConnection, EntityManager } from 'typeorm'
import { instrument } from '../metrics'
import { Logger } from 'winston'
import { PostgresConnectionCredentialsOptions } from 'typeorm/driver/postgres/PostgresConnectionCredentialsOptions'
import { TlsOptions } from 'tls'
import { DatabaseLogger } from './logger'
import * as settings from './settings'

/**
 * The minimum migration version required by this instance of the LSIF process.
 * This should be updated any time a new lsif-server migration is added to the
 * migrations/ directory, as we watch the DB to ensure we're on at least this
 * version prior to making use of the DB (which the frontend may still be
 * migrating).
 */
const MINIMUM_MIGRATION_VERSION = 1528395652

/**
 * Create a Postgres connection. This creates a typorm connection pool with
 * the name `lsif`. The connection configuration is constructed by the method
 * `createPostgresConnectionOptions`. This method blocks until the connection
 * is established, then blocks indefinitely while the database migration state
 * is behind the expected minimum, or dirty. If a connection is not made within
 * a configurable timeout, an exception is thrown.
 *
 * @param configuration The current configuration.
 * @param logger The logger instance.
 */
export async function createPostgresConnection(configuration: Configuration, logger: Logger): Promise<Connection> {
    // Parse current PostgresDSN into connection options usable by
    // the typeorm postgres adapter.
    const url = new URL(configuration.postgresDSN)

    // TODO(efritz) - handle allow, prefer, require, 'verify-ca', and 'verify-full'
    const sslModes: { [name: string]: boolean | TlsOptions } = {
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
    const connection = await connect({ host, port, username, password, database, ssl }, logger)

    // Poll the schema migrations table until we are up to date
    await waitForMigrations(connection, logger)

    return connection
}

/**
 * Create a connection to Postgres. This will re-attempt to access the database while
 * the database does not exist. This is to give some time to the frontend to run the
 * migrations that create the LSIF tables. The retry interval and attempt count can
 * be tuned via `MAX_CONNECTION_RETRIES` and `CONNECTION_RETRY_INTERVAL` environment
 * variables.
 *
 * @param connectionOptions The connection options.
 * @param logger The logger instance.
 */
function connect(connectionOptions: PostgresConnectionCredentialsOptions, logger: Logger): Promise<Connection> {
    return pRetry(
        () => {
            logger.debug('Connecting to Postgres')
            return connectPostgres(connectionOptions, '', logger)
        },
        {
            factor: 1,
            retries: settings.MAX_CONNECTION_RETRIES,
            minTimeout: settings.CONNECTION_RETRY_INTERVAL * 1000,
            maxTimeout: settings.CONNECTION_RETRY_INTERVAL * 1000,
        }
    )
}

/**
 * Create a connection to Postgres.
 *
 * @param connectionOptions The connection options.
 * @param suffix The database suffix (used for testing).
 * @param logger The logger instance
 */
export function connectPostgres(
    connectionOptions: PostgresConnectionCredentialsOptions,
    suffix: string,
    logger: Logger
): Promise<Connection> {
    return _createConnection({
        type: 'postgres',
        name: `lsif${suffix}`,
        entities: pgModels.entities,
        logger: new DatabaseLogger(logger),
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
        logger.debug('Checking database version', { requiredVersion: MINIMUM_MIGRATION_VERSION })

        const version = parseInt(await getMigrationVersion(connection), 10)
        if (isNaN(version) || version < MINIMUM_MIGRATION_VERSION) {
            throw new Error('Postgres database not up to date')
        }
    }

    return pRetry(check, {
        factor: 1,
        retries: settings.MAX_SCHEMA_POLL_RETRIES,
        minTimeout: settings.SCHEMA_POLL_INTERVAL * 1000,
        maxTimeout: settings.SCHEMA_POLL_INTERVAL * 1000,
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
    const rows: {
        version: string
        dirty: boolean
    }[] = await connection.query('select * from schema_migrations')

    if (rows.length > 0 && !rows[0].dirty) {
        return rows[0].version
    }

    throw new Error('Unusable migration state.')
}

/**
 * Instrument `callback` with Postgres query histogram and error counter.
 *
 * @param callback The function invoke with the connection.
 */
export function instrumentQuery<T>(callback: () => Promise<T>): Promise<T> {
    return instrument(metrics.postgresQueryDurationHistogram, metrics.postgresQueryErrorsCounter, callback)
}

/**
 * Invoke `callback` with a transactional Postgres entity manager created
 * from the wrapped connection.
 *
 * @param connection The Postgres connection.
 * @param callback The function invoke with the entity manager.
 */
export function withInstrumentedTransaction<T>(
    connection: Connection,
    callback: (connection: EntityManager) => Promise<T>
): Promise<T> {
    return instrumentQuery(() => connection.transaction(callback))
}

/**
 * Invokes the callback wrapped in instrumentQuery with the given entityManager, if  supplied,
 * and runs the callback in a transaction with a fresh entityManager otherwise.
 *
 * @param connection The Postgres connection.
 * @param entityManager The EntityManager to use as part of a transaction.
 * @param callback The function invoke with the entity manager.
 */
export function instrumentQueryOrTransaction<T>(
    connection: Connection,
    entityManager: EntityManager | undefined,
    callback: (connection: EntityManager) => Promise<T>
): Promise<T> {
    return entityManager
        ? instrumentQuery(() => callback(entityManager))
        : withInstrumentedTransaction(connection, callback)
}
