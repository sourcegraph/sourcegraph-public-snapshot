import * as crc32 from 'crc-32'
import { ADVISORY_LOCK_ID_SALT } from '../constants'
import { PostgresManager } from '../database/postgres'

/**
 * A wrapper around Postgres advisory locks.
 */
export class PostgresLocker extends PostgresManager {
    /**
     * Hold a Postgres advisory lock while executing the given function.
     *
     * @param locker The Postgres locker instance.
     * @param name The name of the lock.
     * @param f The function to execute while holding the lock.
     */
    public async withLock<T>(name: string, f: () => Promise<T>): Promise<T> {
        await this.lock(name)
        try {
            return await f()
        } finally {
            await this.unlock(name)
        }
    }

    /**
     * Acquire an advisory lock with the given name. This will block until the lock can be
     * acquired.
     *
     * See https://www.postgresql.org/docs/9.6/static/explicit-locking.html#ADVISORY-LOCKS.
     *
     * @param name The lock name.
     */
    private lock(name: string): Promise<void> {
        return this.withConnection(connection =>
            connection.query('SELECT pg_advisory_lock($1)', [this.generateLockId(name)])
        )
    }

    /**
     * Release an advisory lock acquired by `lock`.
     *
     * @param name The lock name.
     */
    private unlock(name: string): Promise<void> {
        return this.withConnection(connection =>
            connection.query('SELECT pg_advisory_unlock($1)', [this.generateLockId(name)])
        )
    }

    /**
     * Generate an advisory lock identifier from the given name and application salt. This is
     * based on golang-migrate's advisory lock identifier generation technique, which is in turn
     * inspired by rails migrations.
     *
     * See https://github.com/golang-migrate/migrate/blob/6c96ef02dfbf9430f7286b58afc15718588f2e13/database/util.go#L12.
     *
     * @param name The lock name.
     */
    private generateLockId(name: string): number {
        return crc32.str(name) * ADVISORY_LOCK_ID_SALT
    }
}
