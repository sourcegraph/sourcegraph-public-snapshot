import { Logger as TypeORMLogger } from 'typeorm'
import { PlatformTools } from 'typeorm/platform/PlatformTools'
import { Logger as WinstonLogger } from 'winston'

/**
 * A custom TypeORM logger that only logs slow database queries.
 *
 * We had previously set the TypeORM logging config to `['warn', 'error']`.
 * This caused some issues as it would print the entire parameter set for large
 * batch inserts to stdout. These parameters often included gzipped json payloads,
 * which would lock up the terminal where the server was running.
 *
 * This logger will only print slow database queries. Any other error condition
 * will be printed with the query, parameters, and underlying constraint violation
 * as part of a thrown error, making it unnecessary to log here.
 */
export class DatabaseLogger implements TypeORMLogger {
    constructor(private logger: WinstonLogger) {}

    public logQuerySlow(time: number, query: string, parameters?: unknown[]): void {
        this.logger.warn('Slow database query', {
            query: PlatformTools.highlightSql(query),
            parameters,
            executionTime: time,
        })
    }

    public logQuery(): void {
        /* no-op */
    }

    public logSchemaBuild(): void {
        /* no-op */
    }

    public logMigration(): void {
        /* no-op */
    }

    public logQueryError(): void {
        /* no-op */
    }

    public log(): void {
        /* no-op */
    }
}
