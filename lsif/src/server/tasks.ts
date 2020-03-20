import * as settings from './settings'
import { Connection, EntityManager } from 'typeorm'
import { Logger } from 'winston'
import { UploadManager } from '../shared/store/uploads'
import { DumpManager } from '../shared/store/dumps'
import { ExclusivePeriodicTaskRunner } from '../shared/tasks'
import * as constants from '../shared/constants'
import * as fs from 'mz/fs'
import * as metrics from './metrics'
import * as path from 'path'
import { chunk } from 'lodash'
import { createSilentLogger } from '../shared/logging'
import { TracingContext } from '../shared/tracing'
import { withLock } from '../shared/store/locks'
import { dbFilename, idFromFilename } from '../shared/paths'
import { SRC_FRONTEND_INTERNAL } from '../shared/config/settings'
import { updateCommitsAndDumpsVisibleFromTip } from '../shared/visibility'

/**
 * Begin running cleanup tasks on a schedule in the background.
 *
 * @param connection The Postgres connection.
 * @param dumpManager The dumps manager instance.
 * @param uploadManager The uploads manager instance.
 * @param logger The logger instance.
 */
export function startTasks(
    connection: Connection,
    dumpManager: DumpManager,
    uploadManager: UploadManager,
    logger: Logger
): void {
    const runner = new ExclusivePeriodicTaskRunner(connection, logger)

    runner.register({
        name: 'Resetting stalled uploads',
        intervalMs: settings.RESET_STALLED_UPLOADS_INTERVAL,
        task: ({ ctx }) => resetStalledUploads(uploadManager, ctx),
    })

    runner.register({
        name: 'Cleaning old uploads',
        intervalMs: settings.CLEAN_OLD_UPLOADS_INTERVAL,
        task: ({ ctx }) => cleanOldUploads(uploadManager, ctx),
    })

    runner.register({
        name: 'Purging old dumps',
        intervalMs: settings.PURGE_OLD_DUMPS_INTERVAL,
        task: ({ ctx, connection: taskConnection }) =>
            purgeOldDumps(
                taskConnection,
                dumpManager,
                uploadManager,
                settings.STORAGE_ROOT,
                settings.DBS_DIR_MAXIMUM_SIZE_BYTES,
                ctx
            ),
    })

    runner.register({
        name: 'Cleaning failed uploads',
        intervalMs: settings.CLEAN_FAILED_UPLOADS_INTERVAL,
        task: ({ ctx }) => cleanFailedUploads(ctx),
    })

    runner.register({
        name: 'Updating metrics',
        intervalMs: settings.UPDATE_QUEUE_SIZE_GAUGE_INTERVAL,
        task: () => updateQueueSizeGauge(uploadManager),
        silent: true,
    })

    runner.run()
}

/**
 * Update the value of the unconverted uploads gauge.
 *
 * @param uploadManager The uploads manager instance.
 */
async function updateQueueSizeGauge(uploadManager: UploadManager): Promise<void> {
    metrics.unconvertedUploadSizeGauge.set(await uploadManager.getCount('queued'))
}

/**
 * Move all unlocked uploads that have been in `processing` state for longer than
 * `STALLED_UPLOAD_MAX_AGE` back to the `queued` state.
 *
 * @param uploadManager The uploads manager instance.
 * @param ctx The tracing context.
 */
async function resetStalledUploads(
    uploadManager: UploadManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> {
    for (const id of await uploadManager.resetStalled(settings.STALLED_UPLOAD_MAX_AGE)) {
        logger.debug('Reset stalled upload conversion', { id })
    }
}

/**
 * Remove all upload data older than `UPLOAD_MAX_AGE`.
 *
 * @param uploadManager The uploads manager instance.
 * @param ctx The tracing context.
 */
async function cleanOldUploads(
    uploadManager: UploadManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> {
    const count = await uploadManager.clean(settings.UPLOAD_MAX_AGE)
    if (count > 0) {
        logger.debug('Cleaned old uploads', { count })
    }
}

/**
 * Remove dumps until the space occupied by the dbs directory is below
 * the given limit.
 *
 * @param connection The Postgres connection.
 * @param dumpManager The dumps manager instance.
 * @param uploadManager The uploads manager instance.
 * @param storageRoot The path where SQLite databases are stored.
 * @param maximumSizeBytes The maximum number of bytes.
 * @param ctx The tracing context.
 */
function purgeOldDumps(
    connection: Connection,
    dumpManager: DumpManager,
    uploadManager: UploadManager,
    storageRoot: string,
    maximumSizeBytes: number,
    { logger = createSilentLogger(), span }: TracingContext = {}
): Promise<void> {
    const purge = async (): Promise<void> => {
        // First, remove all the files in the DB dir that don't have a corresponding
        // lsif_upload record in the database. This will happen in the cases where an
        // upload overlaps existing uploads which are deleted in batch from the db,
        // but not from disk. This can also happen if the db file is written during
        // processing but fails later while updating commits for that repo.
        await removeDeadDumps(dumpManager, storageRoot, { logger })

        if (maximumSizeBytes < 0) {
            return Promise.resolve()
        }

        let currentSizeBytes = await dirsize(path.join(storageRoot, constants.DBS_DIR))

        while (currentSizeBytes > maximumSizeBytes) {
            // While our current data usage is too big, find candidate dumps to delete
            const dump = await dumpManager.getOldestPrunableDump()
            if (!dump) {
                logger.warn(
                    'Unable to reduce disk usage of the DB directory because deleting any single dump would drop in-use code intel for a repository.',
                    { currentSizeBytes, softMaximumSizeBytes: maximumSizeBytes }
                )

                break
            }

            logger.info('Pruning dump', {
                repository: dump.repositoryId,
                commit: dump.commit,
                root: dump.root,
            })

            // Delete this dump and subtract its size from the current dir size
            const filename = dbFilename(storageRoot, dump.id)
            currentSizeBytes -= await filesize(filename)

            // This delete cascades to the packages and references tables as well
            await uploadManager.deleteUpload(
                dump.id,
                (entityManager: EntityManager, repositoryId: number): Promise<void> =>
                    updateCommitsAndDumpsVisibleFromTip({
                        entityManager,
                        dumpManager,
                        frontendUrl: SRC_FRONTEND_INTERNAL,
                        repositoryId,
                        ctx: { logger, span },
                    })
            )
        }
    }

    // Ensure only one processor is doing this at the same time so that we don't
    // choose more dumps than necessary to purge. This can happen if the directory
    // size check and the selection of a purgeable dump are interleaved between
    // multiple processors.
    return withLock(connection, 'retention', purge)
}

/**
 * Remove db files that are not reachable from a pending or completed upload record.
 *
 * @param dumpManager The dumps manager instance.
 * @param storageRoot The path where SQLite databases are stored.
 * @param ctx The tracing context.
 */
async function removeDeadDumps(
    dumpManager: DumpManager,
    storageRoot: string,
    { logger = createSilentLogger() }: TracingContext = {}
): Promise<void> {
    let count = 0
    for (const basenames of chunk(
        await fs.readdir(path.join(storageRoot, constants.DBS_DIR)),
        settings.DEAD_DUMP_CHUNK_SIZE
    )) {
        const pathsById = new Map<number, string>()
        for (const basename of basenames) {
            const id = idFromFilename(basename)
            if (!id) {
                continue
            }

            pathsById.set(id, path.join(storageRoot, constants.DBS_DIR, basename))
        }

        const states = await dumpManager.getUploadStates(Array.from(pathsById.keys()))
        for (const [id, dbPath] of pathsById.entries()) {
            if (!states.has(id) || states.get(id) === 'errored') {
                count++
                await fs.unlink(dbPath)
            }
        }
    }

    if (count > 0) {
        logger.debug('Removed dead dumps', { count })
    }
}

/**
 * Remove upload and temp files that are older than `FAILED_UPLOAD_MAX_AGE`. This assumes
 * that an upload conversion's total duration (from enqueue to completion) is less than this
 * interval during healthy operation.
 *
 * @param ctx The tracing context.
 */
async function cleanFailedUploads({ logger = createSilentLogger() }: TracingContext): Promise<void> {
    let count = 0
    for await (const filename of candidateFiles()) {
        if (await purgeFile(filename)) {
            count++
        }
    }

    if (count > 0) {
        logger.debug('Removed old files', { count })
    }
}

/** Return an async iterable that yields the path of all files in the temp and uploads dir. */
async function* candidateFiles(): AsyncIterable<string> {
    for (const directory of [constants.TEMP_DIR, constants.UPLOADS_DIR]) {
        for (const basename of await fs.readdir(path.join(settings.STORAGE_ROOT, directory))) {
            yield path.join(settings.STORAGE_ROOT, directory, basename)
        }
    }
}

/**
 * Remove the given file if it was last modified longer than `FAILED_UPLOAD_MAX_AGE` seconds
 * ago. Returns true if the file was removed and false otherwise.
 *
 * @param filename The file to remove.
 */
async function purgeFile(filename: string): Promise<boolean> {
    if (Date.now() - (await fs.stat(filename)).mtimeMs < settings.FAILED_UPLOAD_MAX_AGE * 1000) {
        return false
    }

    await fs.unlink(filename)
    return true
}

/**
 * Calculate the size of a directory.
 *
 * @param directory The directory path.
 */
async function dirsize(directory: string): Promise<number> {
    return (
        await Promise.all((await fs.readdir(directory)).map(filename => filesize(path.join(directory, filename))))
    ).reduce((a, b) => a + b, 0)
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
        if (!(error && error.code === 'ENOENT')) {
            throw error
        }

        return 0
    }
}
