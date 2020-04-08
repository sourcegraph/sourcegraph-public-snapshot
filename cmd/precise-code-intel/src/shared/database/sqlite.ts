import { Connection, createConnection as _createConnection } from 'typeorm'
import { Logger } from 'winston'
import { DatabaseLogger } from './logger'

/**
 * Create a SQLite connection from the given filename.
 *
 * @param database The database filename.
 * @param entities The set of expected entities present in this schema.
 * @param logger The logger instance.
 */
export function createSqliteConnection(
    database: string,
    // Decorators are not possible type check
    // eslint-disable-next-line @typescript-eslint/ban-types
    entities: Function[],
    logger: Logger
): Promise<Connection> {
    return _createConnection({
        type: 'sqlite',
        name: database,
        database,
        entities,
        synchronize: true,
        logger: new DatabaseLogger(logger),
        maxQueryExecutionTime: 1000,
    })
}
