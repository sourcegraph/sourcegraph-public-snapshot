import * as settings from '../settings'
import * as validation from '../middleware/validation'
import express from 'express'
import { Backend } from '../backend/backend'
import { nextLink } from '../pagination/link'
import { wrap } from 'async-middleware'
import { extractLimitOffset } from '../pagination/limit-offset'

/**
 * Create a router containing the LSIF dump endpoints.
 *
 * @param backend The backend instance.
 */
export function createDumpRouter(backend: Backend): express.Router {
    const router = express.Router()

    interface DumpsQueryArgs {
        query: string
        visibleAtTip: boolean
    }

    router.get(
        '/dumps/:repository',
        validation.validationMiddleware([
            validation.validateQuery,
            validation.validateOptionalBoolean('visibleAtTip'),
            validation.validateLimit,
            validation.validateOffset,
        ]),
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const { query, visibleAtTip }: DumpsQueryArgs = req.query
                const { limit, offset } = extractLimitOffset(req.query, settings.DEFAULT_DUMP_PAGE_SIZE)

                const { dumps, totalCount } = await backend.dumps(
                    decodeURIComponent(req.params.repository),
                    query,
                    visibleAtTip,
                    limit,
                    offset
                )

                if (offset + dumps.length < totalCount) {
                    res.set('Link', nextLink(req, { limit, offset: offset + dumps.length }))
                }

                res.json({ dumps, totalCount })
            }
        )
    )

    router.get(
        '/dumps/:repository/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const dump = await backend.dump(parseInt(req.params.id, 10))
                if (dump?.repository === decodeURIComponent(req.params.repository)) {
                    res.json(dump)
                    return
                }

                throw Object.assign(new Error('LSIF dump not found'), { status: 404 })
            }
        )
    )

    router.delete(
        '/dumps/:repository/:id([0-9]+)',
        wrap(
            async (req: express.Request, res: express.Response): Promise<void> => {
                const dump = await backend.dump(parseInt(req.params.id, 10))
                if (dump?.repository === decodeURIComponent(req.params.repository)) {
                    await backend.deleteDump(dump)
                    res.status(204).send()
                    return
                }

                throw Object.assign(new Error('LSIF dump not found'), { status: 404 })
            }
        )
    )

    return router
}
