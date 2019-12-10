import { Connection, createConnection as _createConnection } from 'typeorm'

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
        logging: ['warn', 'error'],
        maxQueryExecutionTime: 1000,
    })
}
