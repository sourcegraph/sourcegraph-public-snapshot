import promClient from 'prom-client'
import { EntityManager } from 'typeorm'
import { instrument } from '../metrics'
import { QueryDeepPartialEntity } from 'typeorm/query-builder/QueryPartialEntity'

/**
 * A bag of prometheus metric objects that apply to a particular instance of
 * `TableInserter`.
 */
interface TableInserterMetrics {
    /**
     * A histogram that is observed on each round-trip to the database.
     */
    durationHistogram: promClient.Histogram

    /**
     * A counter that increments on each error that occurs during an insertion.
     */
    errorsCounter: promClient.Counter
}

/**
 * A batch inserter for a SQLite table. Inserting hundreds or thousands of rows in
 * a loop is too inefficient, but due to the limit of SQLITE_MAX_VARIABLE_NUMBER,
 * the entire set of values cannot be inserted in one bulk operation either.
 *
 * One inserter instance is created for each table that will receive a bulk
 * payload. The inserter will periodically perform the insert operation
 * when the number of values is at this maximum.
 *
 * See https://www.sqlite.org/limits.html#max_variable_number.
 */
export class TableInserter<T, M extends new () => T> {
    /**
     * The set of entity values that will be inserted in the next invocation of `executeBatch`.
     */
    private batch: QueryDeepPartialEntity<T>[] = []

    /**
     * Creates a new `TableInserter` with the given entity manager, the constructor
     * of the model object for the table, and the maximum batch size. This number
     * should be calculated by floor(MAX_VAR_NUMBER / fields_in_record).
     *
     * @param entityManager A transactional SQLite entity manager.
     * @param model The model object constructor.
     * @param maxBatchSize The maximum number of records that can be inserted at once.
     * @param metrics The bag of metrics to use for this instance of the inserter.
     * @param ignoreConflicts Whether or not to ignore conflicting data on unique constraint violations.
     */
    constructor(
        private entityManager: EntityManager,
        private model: M,
        private maxBatchSize: number,
        private metrics: TableInserterMetrics,
        private ignoreConflicts: boolean = false
    ) {}

    /**
     * Submit a model for insertion. This may happen immediately, on a
     * subsequent call to insert, or when the `flush` method is called.
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
    public flush(): Promise<void> {
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

        let query = this.entityManager
            .createQueryBuilder()
            .insert()
            .into(this.model)
            .values(this.batch)

        if (this.ignoreConflicts) {
            query = query.onConflict('do nothing')
        }

        await instrument(this.metrics.durationHistogram, this.metrics.errorsCounter, () => query.execute())

        this.batch = []
    }
}
