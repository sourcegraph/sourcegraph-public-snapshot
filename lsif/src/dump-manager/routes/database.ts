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
 * For now, each public method of Database (see sif/src/dump-manager/backend/database.ts) is
 * exposed at `/<database-id>/<method>`. This interface is likely to change soon.
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

    const withDatabase = async <T>(
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
        '/dbs/:id([0-9]+)/exists',
        validation.validationMiddleware([validation.validateNonEmptyString('path')]),
        wrap(
            async (req: express.Request, res: express.Response<ExistsResponse>): Promise<void> => {
                const { path }: ExistsQueryArgs = req.query
                await withDatabase(req, res, (database, ctx) => database.exists(path, ctx))
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
        '/dbs/:id([0-9]+)/definitions',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<DefinitionsResponse>): Promise<void> => {
                const { path, line, character }: DefinitionsQueryArgs = req.query
                await withDatabase(req, res, (database, ctx) => database.definitions(path, { line, character }, ctx))
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
        '/dbs/:id([0-9]+)/references',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<ReferencesResponse>): Promise<void> => {
                const { path, line, character }: ReferencesQueryArgs = req.query
                await withDatabase(
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
        '/dbs/:id([0-9]+)/hover',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<HoverResponse>): Promise<void> => {
                const { path, line, character }: HoverQueryArgs = req.query
                await withDatabase(req, res, (database, ctx) => database.hover(path, { line, character }, ctx))
            }
        )
    )

    interface MonikersByPositionQueryArgs {
        path: string
        line: number
        character: number
    }

    type MonikersByPositionResponse = sqliteModels.MonikerData[][]

    router.get(
        '/dbs/:id([0-9]+)/monikersByPosition',
        validation.validationMiddleware([
            validation.validateNonEmptyString('path'),
            validation.validateInt('line'),
            validation.validateInt('character'),
        ]),
        wrap(
            async (req: express.Request, res: express.Response<MonikersByPositionResponse>): Promise<void> => {
                const { path, line, character }: MonikersByPositionQueryArgs = req.query
                await withDatabase(req, res, (database, ctx) =>
                    database.monikersByPosition(path, { line, character }, ctx)
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
        '/dbs/:id([0-9]+)/monikerResults',
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
                await withDatabase(req, res, (database, ctx) =>
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

    interface PackageInformationQueryArgs {
        path: string
        packageInformationId: number
    }

    type PackageInformationResponse = sqliteModels.PackageInformationData | undefined

    router.get(
        '/:id([0-9]+)/packageInformation',
        validation.validationMiddleware([validation.validateNonEmptyString('path')]),
        wrap(
            async (req: express.Request, res: express.Response<PackageInformationResponse>): Promise<void> => {
                const { path, packageInformationId }: PackageInformationQueryArgs = req.query
                await withDatabase(req, res, (database, ctx) =>
                    database.packageInformation(path, packageInformationId, ctx)
                )
            }
        )
    )

    return router
}
