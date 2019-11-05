import * as path from 'path'
import { XrepoDatabase } from './xrepo'
import * as fs from 'mz/fs'
import { TracingContext } from './tracing'
import { createSilentLogger } from './logging'
import { hasErrorCode, dbFilename } from './util'
import * as constants from './constants'

/**
 * Remove dumps until the space occupied by the dbs directory is below
 * the given limit.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param xrepoDatabase The cross-repo database.
 * @param maximumSizeBytes The maximum number of bytes.
 * @param ctx The tracing context.
 */
export function purgeOldDumps(
    storageRoot: string,
    xrepoDatabase: XrepoDatabase,
    maximumSizeBytes: number,
    { logger = createSilentLogger() }: TracingContext = {}
): Promise<void> {
    if (maximumSizeBytes < 0) {
        return Promise.resolve()
    }

    const purge = async (): Promise<void> => {
        let currentSizeBytes = await dirsize(path.join(storageRoot, constants.DBS_DIR))

        while (currentSizeBytes > maximumSizeBytes) {
            // While our current data usage is too big, find candidate dumps to delete
            const dump = await xrepoDatabase.getOldestPrunableDump()
            if (!dump) {
                logger.warning(
                    'Unable to reduce disk usage of the DB directory because deleting any single dump would drop in-use code intel for a repository.',
                    { currentSizeBytes, softMaximumSizeBytes: maximumSizeBytes }
                )

                break
            }

            logger.debug('pruning dump', {
                repository: dump.repository,
                commit: dump.commit,
            })

            // Delete this dump and subtract its size from the current dir size
            const filename = dbFilename(storageRoot, dump.id, dump.repository, dump.commit)
            currentSizeBytes -= await filesize(filename)

            // This delete cascades to the packages and references tables as well
            await xrepoDatabase.deleteDump(dump)
        }
    }

    // Ensure only one worker is doing this at the same time so that we don't
    // choose more dumps than necessary to purge. This can happen if the directory
    // size check and the selection of a purgeable dump are interleaved between
    // multiple workers.
    return withLock(xrepoDatabase, 'retention', purge)
}

/**
 * Hold a Postgres advisory lock while executing the given function.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param name The name of the lock.
 * @param f The function to execute while holding the lock.
 */
async function withLock<T>(xrepoDatabase: XrepoDatabase, name: string, f: () => Promise<T>): Promise<T> {
    await xrepoDatabase.lock(name)
    try {
        return await f()
    } finally {
        await xrepoDatabase.unlock(name)
    }
}

/**
 * Calculate the size of a directory.
 *
 * @param directory The directory path.
 */
async function dirsize(directory: string): Promise<number> {
    return (await Promise.all(
        (await fs.readdir(directory)).map(filename => filesize(path.join(directory, filename)))
    )).reduce((a, b) => a + b, 0)
}

/**
 * Get the file size or zero if it doesn't exist.
 *
 * @param filename The filename.
 */
async function filesize(filename: string): Promise<number> {
    try {
        return (await fs.stat(filename)).size
    } catch (error) {
        if (!hasErrorCode(error, 'ENOENT')) {
            throw error
        }

        return 0
    }
}
