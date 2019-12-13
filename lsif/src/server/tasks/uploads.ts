import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as metrics from '../metrics'
import * as path from 'path'
import * as settings from '../settings'
import { createSilentLogger } from '../../shared/logging'
import { TracingContext } from '../../shared/tracing'
import { UploadsManager } from '../../shared/uploads/uploads'

/**
 * Update the value of the unconverted uploads gauge.
 *
 * @param uploadsManager The uploads manager instance.
 */
export const updateQueueSizeGauge = async (uploadsManager: UploadsManager): Promise<void> =>
    metrics.unconvertedUploadSizeGauge.set(await uploadsManager.getCount('queued'))

/**
 * Move all unlocked uploads that have been in `processing` state for longer than
 * `STALLED_UPLOAD_MAX_AGE` back to the `queued` state.
 *
 * @param uploadsManager The uploads manager instance.
 * @param ctx The tracing context.
 */
export const resetStalledUploads = async (
    uploadsManager: UploadsManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> => {
    for (const id of await uploadsManager.resetStalled(settings.STALLED_UPLOAD_MAX_AGE)) {
        logger.debug('Reset stalled upload conversion', { id })
    }
}

/**
 * Remove all upload data older than `UPLOAD_MAX_AGE`.
 *
 * @param uploadsManager The uploads manager instance.
 * @param ctx The tracing context.
 */
export const cleanOldUploads = async (
    uploadsManager: UploadsManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> => {
    const count = await uploadsManager.clean(settings.UPLOAD_MAX_AGE)
    if (count > 0) {
        logger.debug('Cleaned old uploads', { count })
    }
}

/**
 * Remove upload and temp files that are older than `FAILED_UPLOAD_MAX_AGE`. This assumes
 * that an upload conversion's total duration (from enqueue to completion) is less than this
 * interval during healthy operation.
 *
 * @param ctx The tracing context.
 */
export const cleanFailedUploads = async ({ logger = createSilentLogger() }: TracingContext): Promise<void> => {
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

/**
 * Return an async iterable that yields the path of all files in the temp and uploads dir.
 */
export async function* candidateFiles(): AsyncIterable<string> {
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
    if (Date.now() - (await fs.stat(filename)).mtimeMs < settings.FAILED_UPLOAD_MAX_AGE) {
        return false
    }

    await fs.unlink(filename)
    return true
}
