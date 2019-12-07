import * as crc32 from 'crc-32'
import { ADVISORY_LOCK_ID_SALT } from '../constants'
import { Connection } from 'typeorm'

/**
 * Hold a Postgres advisory lock while executing the given function. Note that acquiring
 * an advisory lock is an (indefinitely) blocking operation.
 *
 * For more information, see
 * https://www.postgresql.org/docs/9.6/static/explicit-locking.html#ADVISORY-LOCKS
 *
 * @param connection The Postgres connection.
 * @param name The name of the lock.
 * @param f The function to execute while holding the lock.
 */
export async function withLock<T>(connection: Connection, name: string, f: () => Promise<T>): Promise<T> {
    const lockId = createLockId(name)
    await connection.query('SELECT pg_advisory_lock($1)', [lockId])
    try {
        return await f()
    } finally {
        await connection.query('SELECT pg_advisory_unlock($1)', [lockId])
    }
}

/**
 * Hold a Postgres advisory lock while executing the given function. If the lock cannot be
 * acquired immediately, the function will return undefined without invoking the function.
 *
 * For more information, see
 * https://www.postgresql.org/docs/9.6/static/explicit-locking.html#ADVISORY-LOCKS
 *
 * @param connection The Postgres connection.
 * @param name The name of the lock.
 * @param f The function to execute while holding the lock.
 */
export async function tryWithLock<T>(
    connection: Connection,
    name: string,
    f: () => Promise<T>
): Promise<T | undefined> {
    const lockId = createLockId(name)
    if (await connection.query('SELECT pg_try_advisory_lock($1)', [lockId])) {
        try {
            return await f()
        } finally {
            await connection.query('SELECT pg_advisory_unlock($1)', [lockId])
        }
    }

    return undefined
}

/**
 * Create an integer identifier that will be unique to this app, but will always be the same for this given
 * name within the application.
 *
 * We base our advisory lock identifier generation technique on golang-migrate. For the original source, see
 * https://github.com/golang-migrate/migrate/blob/6c96ef02dfbf9430f7286b58afc15718588f2e13/database/util.go#L12.
 *
 * Advisory lock ids should be deterministically generated such that a single app will return the same lock id
 * for the same name, but distinct apps are unlikely to generate the same id (using the same name or not). To
 * accomplish this, we hash the name into an integer, then multiply it by some app-specific salt to reduce the
 * collision space with another application. Each app should choose a salt uniformly at random. This
 * application's salt is distinct from the golang-migrate salt.
 *
 * @param name The name of the lock.
 */
function createLockId(name: string): number {
    return crc32.str(name) * ADVISORY_LOCK_ID_SALT
}
