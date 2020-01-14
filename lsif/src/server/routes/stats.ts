import express from 'express'
import { UploadManager } from '../../shared/store/uploads'

/**
 * Create a router containing the stats endpoint.
 *
 * @param uploadManager The uploads manager instance.
 */
export function createStatsRouter(uploadManager: UploadManager): express.Router {
    const router = express.Router()
    router.get('/stats', async (_, res) =>
        res.send({
            completedUploads: await uploadManager.getCount('completed'),
            uniqueRepositories: await uploadManager.countRepositories(),
        })
    )

    return router
}
