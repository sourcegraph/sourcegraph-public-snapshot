import { Connection, createConnection as _createConnection } from 'typeorm'
import { entities } from './models.xrepo'
import { PostgresConnectionCredentialsOptions } from 'typeorm/driver/postgres/PostgresConnectionCredentialsOptions'
import { userInfo } from 'os'

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
    return _createConnection({ type: 'sqlite', name: database, database, entities, ...baseOptions })
}

/**
 * Create a Postgres connection. This creates a typorm connection
 * pool with the name `xrepo`. The connection configuration is
 * constructed by `createPostgresConnectionOptions.
 */
export function createPostgresConnection(): Promise<Connection> {
    return _createConnection({
        type: 'postgres',
        name: 'xrepo',
        entities,
        ...baseOptions,
        ...createPostgresConnectionOptions(),
    })
}

/**
 * Create Postgres connection options. This attempts to mirror the
 * variables used by the frontend. If `PGDATASOURCE` is set, that is
 * used as the full Postgres connection string. Otherwise the connection
 * parameters are determined by `PG_` environment variables,
 */
function createPostgresConnectionOptions(): PostgresConnectionCredentialsOptions {
    if (process.env.PGDATASOURCE) {
        return { url: process.env.PGDATASOURCE }
    }

    return {
        host: process.env.PGHOST || '127.0.0.1',
        port: process.env.PGPORT ? parseInt(process.env.PGPORT, 10) : 5432,
        username: process.env.PGUSER || userInfo().username || 'postgres',
        password: process.env.PGPASSWORD,
        database: process.env.PGDATABASE || 'lsif',
        ssl: process.env.PGSSLMODE === 'disable' ? false : undefined,
    }
}
