import { Connection } from 'typeorm'
import { QueryDeepPartialEntity } from 'typeorm/query-builder/QueryPartialEntity'

/**
 * `Inserter` is a batch inserter for a SQLite table. Inserting hundreds or
 * thousands of rows in a loop is too inefficient, but due to the limit of
 * SQLITE_MAX_VARIABLE_NUMBER, the entire set of values cannot be inserted
 * in one bulk operation either.
 *
 * One inserter instance is created for each table that will receive a bulk
 * payload. The inserter will periodically perform the insert operation
 * when the number of values is at this maximum.
 *
 * See https://www.sqlite.org/limits.html#max_variable_number.
 */
export class Inserter<T> {
    /**
     * `batch` is the set of entity values that will be inserted in the
     * next invocation of `executeBatch`.
     */
    private batch: QueryDeepPartialEntity<T>[] = []

    /**
     * Creates a new `Inserter` with the given connection (which must remain
     * active during the lifetime of this instance), the constructor of the
     * model object for the table, and the maximum batch size. This number
     * should be calculated by floor(MAX_VAR_NUMBER / fields_in_record).
     *
     * @param connection The SQLite ocnnection.
     * @param model The model object constructor.
     * @param maxBatchSize The maximum number of records that can be inserted at once.
     */
    constructor(private connection: Connection, private model: Function, private maxBatchSize: number) {}

    /**
     * Submit a model for insertion. This may happen immediately, on a
     * subsequent call to insert, or when the `finalize` method is called.
     *
     * @param model The instance to save.
     */
    public async insert(model: QueryDeepPartialEntity<T>): Promise<void> {
        this.batch.push(model)

        if (this.batch.length >= this.maxBatchSize) {
            await this.executeBatch()
        }
    }

    /**
     * Ensure any outstanding records are inserted into the database.
     */
    public finalize(): Promise<void> {
        return this.executeBatch()
    }

    /**
     * If the current batch is non-empty, then perform an insert operation
     * and reset the batch array.
     */
    private async executeBatch(): Promise<void> {
        if (this.batch.length === 0) {
            return
        }

        await this.connection
            .createQueryBuilder()
            .insert()
            .into(this.model)
            .values(this.batch)
            .execute()
            .then(() => {})

        this.batch = []
    }
}
