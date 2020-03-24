import express from 'express'
import { wrap } from 'async-middleware'
import { UploadManager } from '../../shared/store/uploads'
import { DumpManager } from '../../shared/store/dumps'
import { EntityManager } from 'typeorm'
import { SRC_FRONTEND_INTERNAL } from '../../shared/config/settings'
import { TracingContext, addTags } from '../../shared/tracing'
import { Span } from 'opentracing'
import { Logger } from 'winston'
import { updateCommitsAndDumpsVisibleFromTip } from '../../shared/visibility'
import { json } from 'body-parser'

/**
 * Create a router containing the endpoints used by the dump manager.
 *
 * @param dumpManager The dumps manager instance.
 * @param uploadManager The uploads manager instance.
 * @param logger The logger instance.
 */
export function createInternalRouter(
    dumpManager: DumpManager,
    uploadManager: UploadManager,
    logger: Logger
): express.Router {
    const router = express.Router()

    /**
     * Create a tracing context from the request logger and tracing span
     * tagged with the given values.
     *
     * @param req The express request.
     * @param tags The tags to apply to the logger and span.
     */
    const createTracingContext = (
        req: express.Request & { span?: Span },
        tags: { [K: string]: unknown }
    ): TracingContext => addTags({ logger, span: req.span }, tags)

    interface StatesBody {
        ids: number[]
    }

    type StatesResponse = Map<number, string>

    router.post(
        '/uploads',
        json(),
        wrap(
            async (req: express.Request, res: express.Response<StatesResponse>): Promise<void> => {
                const { ids }: StatesBody = req.body
                res.json(await dumpManager.getUploadStates(ids))
            }
        )
    )

    type PruneResponse = { id: number } | null

    router.post(
        '/prune',
        wrap(
            async (req: express.Request, res: express.Response<PruneResponse>): Promise<void> => {
                const ctx = createTracingContext(req, {})

                const dump = await dumpManager.getOldestPrunableDump()
                if (!dump) {
                    res.json(null)
                    return
                }

                logger.info('Pruning dump', {
                    repository: dump.repositoryId,
                    commit: dump.commit,
                    root: dump.root,
                })

                // This delete cascades to the packages and references tables as well
                await uploadManager.deleteUpload(
                    dump.id,
                    (entityManager: EntityManager, repositoryId: number): Promise<void> =>
                        updateCommitsAndDumpsVisibleFromTip({
                            entityManager,
                            dumpManager,
                            frontendUrl: SRC_FRONTEND_INTERNAL,
                            repositoryId,
                            ctx,
                        })
                )

                res.json({ id: dump.id })
            }
        )
    )

    return router
}
