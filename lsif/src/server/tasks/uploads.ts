import * as metrics from '../metrics'
import * as settings from '../settings'
import { createSilentLogger } from '../../shared/logging'
import { TracingContext } from '../../shared/tracing'
import { UploadManager } from '../../shared/store/uploads'

/**
 * Update the value of the unconverted uploads gauge.
 *
 * @param uploadManager The uploads manager instance.
 */
export const updateQueueSizeGauge = async (uploadManager: UploadManager): Promise<void> =>
    metrics.unconvertedUploadSizeGauge.set(await uploadManager.getCount('queued'))

/**
 * Move all unlocked uploads that have been in `processing` state for longer than
 * `STALLED_UPLOAD_MAX_AGE` back to the `queued` state.
 *
 * @param uploadManager The uploads manager instance.
 * @param ctx The tracing context.
 */
export const resetStalledUploads = async (
    uploadManager: UploadManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> => {
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
export const cleanOldUploads = async (
    uploadManager: UploadManager,
    { logger = createSilentLogger() }: TracingContext
): Promise<void> => {
    // TODO - necessary to do anymore?
    const count = await uploadManager.clean(settings.UPLOAD_MAX_AGE)
    if (count > 0) {
        logger.debug('Cleaned old uploads', { count })
    }
}
