import * as settings from '../settings'
import express from 'express'
import { addTags, TracingContext } from '../../shared/tracing'
import { Logger } from 'winston'
import { pipeline as _pipeline } from 'stream'
import { Span } from 'opentracing'
import { wrap } from 'async-middleware'
import { Database } from '../backend/database'
import * as sqliteModels from '../../shared/models/sqlite'
import { InternalLocation } from '../backend/location'
import { dbFilename } from '../../shared/paths'
import * as lsp from 'vscode-languageserver-protocol'
import * as validation from '../../shared/api/middleware/validation'

/**
 * Create a router containing the SQLite query endpoints.
 *
 * @param logger The logger instance.
 */
export function createDatabaseRouter(logger: Logger): express.Router {
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

    const onDatabase = async <T>(
        req: express.Request,
        res: express.Response<T>,
        handler: (database: Database, ctx?: TracingContext) => Promise<T>
    ): Promise<void> => {
        const id = parseInt(req.params.id, 10)
        const ctx = createTracingContext(req, { id })
        const database = new Database(id, dbFilename(settings.STORAGE_ROOT, id))

        const payload = await handler(database, ctx)
        res.json(payload)
    }

    interface ExistsQueryArgs {
        path: string
    }

    type ExistsResponse = boolean

    router.get(
        '/:id([0-9]+)/exists',
        validation.validationMiddleware([validation.validateNonEmptyString('path')]),
        wrap(
            async (req: express.Request, res: express.Response<ExistsResponse>): Promise<void> => {
                const { path }: ExistsQueryArgs = req.query
                await onDatabase(req, res, (database, ctx) => database.exists(path, ctx))
            }
        )
    )

    interface DefinitionsQueryArgs {
        path: string
        line: number
        character: number
    }

    type DefinitionsResponse = InternalLocation[]

    router.get(
        '/:id([0-9]+)/definitions',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<DefinitionsResponse>): Promise<void> => {
                const { path, line, character }: DefinitionsQueryArgs = req.query
                await onDatabase(req, res, (database, ctx) => database.definitions(path, { line, character }, ctx))
            }
        )
    )

    interface ReferencesQueryArgs {
        path: string
        line: number
        character: number
    }

    type ReferencesResponse = InternalLocation[]

    router.get(
        '/:id([0-9]+)/references',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<ReferencesResponse>): Promise<void> => {
                const { path, line, character }: ReferencesQueryArgs = req.query
                await onDatabase(
                    req,
                    res,
                    async (database, ctx) => (await database.references(path, { line, character }, ctx)).values
                )
            }
        )
    )

    interface HoverQueryArgs {
        path: string
        line: number
        character: number
    }

    type HoverResponse = { text: string; range: lsp.Range } | null

    router.get(
        '/:id([0-9]+)/hover',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<HoverResponse>): Promise<void> => {
                const { path, line, character }: HoverQueryArgs = req.query
                await onDatabase(req, res, (database, ctx) => database.hover(path, { line, character }, ctx))
            }
        )
    )

    interface GetRangeByPositionQueryArgs {
        path: string
        line: number
        character: number
    }

    interface GetRangeByPositionResponse {
        document: sqliteModels.DocumentData | undefined
        ranges: sqliteModels.RangeData[]
    }

    router.get(
        '/:id([0-9]+)/getRangeByPosition',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<GetRangeByPositionResponse>): Promise<void> => {
                const { path, line, character }: GetRangeByPositionQueryArgs = req.query
                await onDatabase(req, res, (database, ctx) =>
                    database.getRangeByPosition(path, { line, character }, ctx)
                )
            }
        )
    )

    interface MonikerResultsQueryArgs {
        model: string
        scheme: string
        identifier: string
        skip?: number
        take?: number
    }

    interface MonikerResultsResponse {
        locations: { path: string; range: lsp.Range }[]
        count: number
    }

    router.get(
        '/:id([0-9]+)/monikerResults',
        validation.validationMiddleware([
            validation.validateNonEmptyString('model'),
            validation.validateNonEmptyString('scheme'),
            validation.validateNonEmptyString('identifier'),
            validation.validateOptionalInt('skip'),
            validation.validateOptionalInt('take'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<MonikerResultsResponse>): Promise<void> => {
                const { model, scheme, identifier, skip, take }: MonikerResultsQueryArgs = req.query
                await onDatabase(req, res, (database, ctx) =>
                    database.monikerResults(
                        model === 'definition' ? sqliteModels.DefinitionModel : sqliteModels.ReferenceModel,
                        { scheme, identifier },
                        { skip, take },
                        ctx
                    )
                )
            }
        )
    )

    interface GetDocumentByPathQueryArgs {
        path: string
    }

    type GetDocumentByPathResponse = sqliteModels.DocumentData | undefined

    router.get(
        '/:id([0-9]+)/getDocumentByPath',
        validation.validationMiddleware([validation.validateNonEmptyString('path')]),
        wrap(
            async (req: express.Request, res: express.Response<GetDocumentByPathResponse>): Promise<void> => {
                const { path }: GetDocumentByPathQueryArgs = req.query
                await onDatabase(req, res, (database, ctx) => database.getDocumentByPath(path, ctx))
            }
        )
    )

    return router
}
