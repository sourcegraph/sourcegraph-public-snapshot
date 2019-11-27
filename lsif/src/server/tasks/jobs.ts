import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as metrics from '../metrics'
import * as path from 'path'
import * as settings from '../settings'
import { createSilentLogger } from '../../shared/logging'
import { TracingContext } from '../../shared/tracing'
import { JobStatusClean, Queue } from 'bull'

const cleanStatuses: JobStatusClean[] = ['completed', 'wait', 'active', 'delayed', 'failed']

/**
 * Update the value of the unconverted uploads gauge.
 *
 * @param queue The queue instance.
 */
export const updateQueueSizeGauge = (queue: Queue): Promise<void> =>
    queue
        .getJobCountByTypes('waiting')
        // The type of this method is wrong in the types package: it says that
        // it returns a counts object, but it really returns a scalar count.
        .then((count: unknown) => metrics.queueSizeGauge.set(count as number))
        .catch(() => {})

/**
 * Remove all job data older than `JOB_MAX_AGE`.
 *
 * @param queue The queue instance.
 * @param ctx The tracing context.
 */
export const cleanOldJobs = async (queue: Queue, { logger = createSilentLogger() }: TracingContext): Promise<void> => {
    const removedJobs = await Promise.all(cleanStatuses.map(status => queue.clean(settings.JOB_MAX_AGE * 1000, status)))

    for (const [status, count] of removedJobs.map((jobs, i) => [cleanStatuses[i], jobs.length])) {
        if (count > 0) {
            logger.debug('Cleaned old jobs', { status, count })
        }
    }
}

/**
 * Remove upload and temp files that are older than `FAILED_JOB_MAX_AGE`. This assumes
 * that a job's total duration (from enqueue to completion) is less than this interval
 * during healthy operation.
 *
 * @param ctx The tracing context.
 */
export const cleanFailedJobs = async ({ logger = createSilentLogger() }: TracingContext): Promise<void> => {
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
 * Remove the given file if it was last modified longer than `FAILED_JOB_MAX_AGE` seconds
 * ago. Returns true if the file was removed and false otherwise.
 *
 * @param filename The file to remove.
 */
async function purgeFile(filename: string): Promise<boolean> {
    if (Date.now() - (await fs.stat(filename)).mtimeMs < settings.FAILED_JOB_MAX_AGE) {
        return false
    }

    await fs.unlink(filename)
    return true
}
