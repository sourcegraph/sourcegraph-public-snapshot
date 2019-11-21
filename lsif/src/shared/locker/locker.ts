import * as crc32 from 'crc-32'
import { ADVISORY_LOCK_ID_SALT } from '../constants'
import { Connection } from 'typeorm'
import { instrumentQuery } from '../database/postgres'

/**
 * Hold a Postgres advisory lock while executing the given function. We base our advisory lock
 * identifier generation technique on golang-migrate. Note that acquiring an advisory lock is
 * an (indefinitely) blocking operation.
 *
 * For more information, see
 * https://www.postgresql.org/docs/9.6/static/explicit-locking.html#ADVISORY-LOCKS, and
 * https://github.com/golang-migrate/migrate/blob/6c96ef02dfbf9430f7286b58afc15718588f2e13/database/util.go#L12.
 *
 * @param connection The Postgres connection.
 * @param name The name of the lock.
 * @param f The function to execute while holding the lock.
 */
export async function withLock<T>(connection: Connection, name: string, f: () => Promise<T>): Promise<T> {
    // Create an integer identifier that will be unique to this app, but
    // will always be the same for this given name within the application.
    const lockId = crc32.str(name) * ADVISORY_LOCK_ID_SALT

    // Acquire lock
    await instrumentQuery(() => connection.query('SELECT pg_advisory_lock($1)', [lockId]))

    try {
        // Critical section
        return await f()
    } finally {
        // Release lock
        await instrumentQuery(() => connection.query('SELECT pg_advisory_unlock($1)', [lockId]))
    }
}
