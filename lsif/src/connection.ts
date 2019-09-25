import { Connection, createConnection as _createConnection } from 'typeorm'
import { PostgresConnectionOptions } from 'typeorm/driver/postgres/PostgresConnectionOptions'
import { SqliteConnectionOptions } from 'typeorm/driver/sqlite/SqliteConnectionOptions'

/**
 * The common options used for typeorm connections.
 */
const baseOptions: {
    synchronize: boolean
    logging: ('query' | 'schema' | 'error' | 'warn' | 'info' | 'log' | 'migration')[]
    maxQueryExecutionTime: number
} = {
    synchronize: true,
    logging: ['error', 'warn'],
    maxQueryExecutionTime: 1000,
}

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
    const options: SqliteConnectionOptions = {
        type: 'sqlite',
        name: database,
        database,
        entities,
        ...baseOptions,
    }

    return _createConnection(options)
}

/**
 * Create a Postgres connection from the given database url.
 *
 * @param name The name of the connection (in the connection pool).
 * @param url The database connection url.
 * @param entities The set of expected entities present in this schema.
 */
export function createPostgresConnection(
    name: string,
    url: string,
    // Decorators are not possible type check
    // eslint-disable-next-line @typescript-eslint/ban-types
    entities: Function[]
): Promise<Connection> {
    const options: PostgresConnectionOptions = {
        type: 'postgres',
        name,
        url,
        entities,
        ...baseOptions,
    }

    return _createConnection(options)
}
