import * as path from 'path'
import { XrepoDatabase } from './xrepo'
import * as fs from 'mz/fs'
import { TracingContext } from './tracing'
import { createSilentLogger } from './logging'
import { Logger } from 'winston'
import { hasErrorCode, dbFilename } from './util'
import * as constants from './constants'

/**
 * Remove dumps until the space occupied by the dbs directory is below
 * the given limit.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param xrepoDatabase The cross-repo database.
 * @param maximumBytes The maximum number of bytes.
 * @param ctx The tracing context.
 */
export async function purgeOldDumps(
    storageRoot: string,
    xrepoDatabase: XrepoDatabase,
    maximumBytes: number,
    { logger = createSilentLogger() }: TracingContext = {}
): Promise<void> {
    if (maximumBytes < 0) {
        return
    }

    const lockName = 'retention'
    await xrepoDatabase.lock(lockName)

    try {
        // Ensure only one worker is doing this at the same time so that we don't
        // choose more dumps than necessary to purge. This can happen if the directory
        // size check and the selection of a purgeable dump are interleaved between
        // multiple workers.

        await purgeOldDumpsLocked(storageRoot, xrepoDatabase, maximumBytes, logger)
    } finally {
        await xrepoDatabase.unlock(lockName)
    }
}

/**
 * Remove dumps until the space occupied by the dbs directory is below
 * the given limit.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param xrepoDatabase The cross-repo database.
 * @param maximumBytes The maximum number of bytes.
 * @param logger The logger instance.
 */
async function purgeOldDumpsLocked(
    storageRoot: string,
    xrepoDatabase: XrepoDatabase,
    maximumBytes: number,
    logger: Logger
): Promise<void> {
    let currentSizeBytes = await dirsize(path.join(storageRoot, constants.DBS_DIR))

    while (currentSizeBytes > maximumBytes) {
        // While our current data usage is too big, find candidate dumps to delete
        const dumps = await xrepoDatabase.getOldestPrunableDumps()
        if (dumps.length === 0) {
            logger.error('Failed to select a prunable dump', { currentSizeBytes, maximumBytes })
            break
        }

        logger.debug('pruning dumps', {
            repository: dumps[0].repository,
            commit: dumps[0].commit,
        })

        for (const dump of dumps) {
            // Delete this dump and subtract its size from the current dir size
            const filename = dbFilename(storageRoot, dump.id, dump.repository, dump.commit)
            await xrepoDatabase.deleteDump(dump)
            currentSizeBytes -= await filesize(filename)
        }
    }
}

/**
 * Calculate the size of a directory.
 *
 * @param directory The directory path.
 */
async function dirsize(directory: string): Promise<number> {
    return (await Promise.all((await fs.readdir(directory)).map(filesize))).reduce((a, b) => a + b, 0)
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
