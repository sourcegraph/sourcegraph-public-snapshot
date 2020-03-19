import * as cache from './cache'
import * as sqliteModels from '../../shared/models/sqlite'
import * as lsp from 'vscode-languageserver-protocol'
import * as metrics from '../metrics'
import * as pgModels from '../../shared/models/pg'
import { Connection } from 'typeorm'
import { DefaultMap } from '../../shared/datastructures/default-map'
import { gunzipJSON } from '../../shared/encoding/json'
import { hashKey } from '../../shared/models/hash'
import { instrument } from '../../shared/metrics'
import { logSpan, TracingContext, logAndTraceCall, addTags } from '../../shared/tracing'
import { mustGet } from '../../shared/maps'
import { Logger } from 'winston'
import { createSilentLogger } from '../../shared/logging'
import { InternalLocation, OrderedLocationSet } from './location'
import * as settings from '../settings'

/** The maximum number of results in a logSpan value. */
const MAX_SPAN_ARRAY_LENGTH = 20

/** A wrapper around operations related to a single SQLite dump. */
export class Database {
    /**
     * A static map of database paths to the `numResultChunks` value of their
     * metadata row. This map is populated lazily as the values are needed.
     */
    private static numResultChunks = new Map<string, number>()
    private static connectionCache = new cache.ConnectionCache(settings.CONNECTION_CACHE_CAPACITY)
    private static documentCache = new cache.DocumentCache(settings.DOCUMENT_CACHE_CAPACITY)
    private static resultChunkCache = new cache.ResultChunkCache(settings.RESULT_CHUNK_CACHE_CAPACITY)

    /**
     * Create a new `Database` with the given dump record, and the SQLite file
     * on disk that contains data for a particular repository and commit.
     *
     * @param dumpId The identifier of the dump for which this database answers queries.
     * @param databasePath The path to the database file.
     */
    constructor(private dumpId: pgModels.DumpId, private databasePath: string) {}

    /**
     * Retrieve all document paths from the database.
     *
     * @param ctx The tracing context.
     */
    public documentPaths(ctx: TracingContext = {}): Promise<string[]> {
        return this.logAndTraceCall(ctx, 'Fetching document paths', async ctx => {
            const paths: { path: string }[] = await this.withConnection(
                connection =>
                    connection
                        .getRepository(sqliteModels.DocumentModel)
                        .createQueryBuilder()
                        .select('path')
                        .getRawMany(),
                ctx.logger
            )

            return paths.map(({ path }) => path)
        })
    }

    /**
     * Determine if data exists for a particular document in this database.
     *
     * @param path The path of the document.
     * @param ctx The tracing context.
     */
    public exists(path: string, ctx: TracingContext = {}): Promise<boolean> {
        return this.logAndTraceCall(
            ctx,
            'Checking if path exists',
            async () => (await this.getDocumentByPath(path)) !== undefined
        )
    }

    /**
     * Return a list of locations that define the symbol at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async definitions(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<InternalLocation[]> {
        return this.logAndTraceCall(ctx, 'Fetching definitions', async ctx => {
            const { document, ranges } = await this.getRangeByPosition(path, position, ctx)
            if (!document || ranges.length === 0) {
                return []
            }

            for (const range of ranges) {
                if (!range.definitionResultId) {
                    continue
                }

                const definitionResults = await this.getResultById(range.definitionResultId)
                this.logSpan(ctx, 'definition_results', {
                    definitionResultId: range.definitionResultId,
                    definitionResults: definitionResults.slice(0, MAX_SPAN_ARRAY_LENGTH),
                    numDefinitionResults: definitionResults.length,
                })

                // TODO - due to some bugs in tsc... this fixes the tests and some typescript examples
                // Not sure of a better way to do this right now until we work through how to patch
                // lsif-tsc to handle node_modules inclusion (or somehow blacklist it on import).

                if (!definitionResults.some(v => v.documentPath.includes('node_modules'))) {
                    return this.convertRangesToInternalLocations(path, document, definitionResults)
                }
            }

            return []
        })
    }

    /**
     * Return a list of unique locations that reference the symbol at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async references(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<OrderedLocationSet> {
        return this.logAndTraceCall(ctx, 'Fetching references', async ctx => {
            const { document, ranges } = await this.getRangeByPosition(path, position, ctx)
            if (!document || ranges.length === 0) {
                return new OrderedLocationSet()
            }

            const locationSet = new OrderedLocationSet()
            for (const range of ranges) {
                if (range.referenceResultId) {
                    const referenceResults = await this.getResultById(range.referenceResultId)
                    this.logSpan(ctx, 'reference_results', {
                        referenceResultId: range.referenceResultId,
                        referenceResults: referenceResults.slice(0, MAX_SPAN_ARRAY_LENGTH),
                        numReferenceResults: referenceResults.length,
                    })

                    if (referenceResults.length > 0) {
                        for (const location of await this.convertRangesToInternalLocations(
                            path,
                            document,
                            referenceResults
                        )) {
                            locationSet.push(location)
                        }
                    }
                }
            }

            return locationSet
        })
    }

    /**
     * Return the hover content for the symbol at the given position.
     *
     * @param path The path of the document to which the position belongs.
     * @param position The current hover position.
     * @param ctx The tracing context.
     */
    public async hover(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<{ text: string; range: lsp.Range } | null> {
        return this.logAndTraceCall(ctx, 'Fetching hover', async ctx => {
            const { document, ranges } = await this.getRangeByPosition(path, position, ctx)
            if (!document || ranges.length === 0) {
                return null
            }

            for (const range of ranges) {
                if (!range.hoverResultId) {
                    continue
                }

                this.logSpan(ctx, 'hover_result', { hoverResultId: range.hoverResultId })

                // Extract text
                const text = mustGet(document.hoverResults, range.hoverResultId, 'hoverResult')

                // Return first defined hover result for the inner-most range. This response
                // includes the entire range so that the highlighted portion in the UI can be
                // accurate (rather than approximated by the tokenizer).
                return { text, range: createRange(range) }
            }

            return null
        })
    }

    /**
     * Return a parsed document that describes the given path as well as the ranges
     * from that document that contains the given position. If multiple ranges are
     * returned, then the inner-most ranges will occur before the outer-most ranges.
     *
     * @param path The path of the document.
     * @param position The user's hover position.
     * @param ctx The tracing context.
     */
    public getRangeByPosition(
        path: string,
        position: lsp.Position,
        ctx: TracingContext = {}
    ): Promise<{ document: sqliteModels.DocumentData | undefined; ranges: sqliteModels.RangeData[] }> {
        return this.logAndTraceCall(ctx, 'Fetching range by position', async ctx => {
            const document = await this.getDocumentByPath(path)
            if (!document) {
                return { document: undefined, ranges: [] }
            }

            const ranges = findRanges(document.ranges.values(), position)
            this.logSpan(ctx, 'matching_ranges', { ranges: cleanRanges(ranges) })
            return { document, ranges }
        })
    }

    /**
     * Query the definitions or references table of `db` for items that match the given moniker.
     * Convert each result into an `InternalLocation`. The `pathTransformer` function is invoked
     * on each result item to modify the resulting locations.
     *
     * @param model The constructor for the model type.
     * @param moniker The target moniker.
     * @param pagination A limit and offset to use for the query.
     * @param ctx The tracing context.
     */
    public monikerResults(
        model: typeof sqliteModels.DefinitionModel | typeof sqliteModels.ReferenceModel,
        moniker: Pick<sqliteModels.MonikerData, 'scheme' | 'identifier'>,
        pagination: { skip?: number; take?: number },
        ctx: TracingContext = {}
    ): Promise<{ locations: InternalLocation[]; count: number }> {
        return this.logAndTraceCall(ctx, 'Fetching moniker results', async ctx => {
            const [results, count] = await this.withConnection(
                connection =>
                    connection
                        .getRepository<sqliteModels.DefinitionModel | sqliteModels.ReferenceModel>(model)
                        .findAndCount({
                            where: {
                                scheme: moniker.scheme,
                                identifier: moniker.identifier,
                            },
                            ...pagination,
                        }),
                ctx.logger
            )

            this.logSpan(ctx, 'symbol_results', {
                moniker,
                results: results.slice(0, MAX_SPAN_ARRAY_LENGTH),
                numResults: results.length,
            })

            const locations = results.map(result => ({
                dumpId: this.dumpId,
                path: result.documentPath,
                range: createRange(result),
            }))

            return { locations, count }
        })
    }

    /**
     * Return a parsed document that describes the given path. The result of this
     * method is cached across all database instances. If the document is not found
     * it returns undefined; other errors will throw.
     *
     * @param path The path of the document.
     * @param ctx The tracing context.
     */
    public async getDocumentByPath(
        path: string,
        ctx: TracingContext = {}
    ): Promise<sqliteModels.DocumentData | undefined> {
        const factory = async (): Promise<cache.EncodedJsonCacheValue<sqliteModels.DocumentData>> => {
            const document = await this.withConnection(
                connection => connection.getRepository(sqliteModels.DocumentModel).findOneOrFail(path),
                ctx.logger
            )

            return {
                size: document.data.length,
                data: await gunzipJSON<sqliteModels.DocumentData>(document.data),
            }
        }

        try {
            return await Database.documentCache.withValue(`${this.databasePath}::${path}`, factory, document =>
                Promise.resolve(document.data)
            )
        } catch (error) {
            if (error.name === 'EntityNotFound') {
                return undefined
            }

            throw error
        }
    }

    //
    // Helper Functions

    /**
     * Convert a set of range-document pairs (from a definition or reference query) into
     * a set of `InternalLocation` object. Each pair holds the range identifier as well as
     * the document path. For document paths matching the loaded document, find the range
     * data locally. For all other paths, find the document in this database and find the
     * range in that document.
     *
     * @param path The path of the document for this query.
     * @param document The document object for this query.
     * @param resultData A list of range ids and the document they belong to.
     */
    private async convertRangesToInternalLocations(
        path: string,
        document: sqliteModels.DocumentData,
        resultData: sqliteModels.DocumentPathRangeId[]
    ): Promise<InternalLocation[]> {
        // Group by document path so we only have to load each document once
        const groupedResults = new DefaultMap<string, Set<sqliteModels.RangeId>>(() => new Set())

        for (const { documentPath, rangeId } of resultData) {
            groupedResults.getOrDefault(documentPath).add(rangeId)
        }

        let results: InternalLocation[] = []
        for (const [documentPath, rangeIds] of groupedResults) {
            if (documentPath === path) {
                // If the document path is this document, convert the locations directly
                results = results.concat(mapRangesToInternalLocations(this.dumpId, document.ranges, path, rangeIds))
                continue
            }

            // Otherwise, we need to get the correct document
            const sibling = await this.getDocumentByPath(documentPath)
            if (!sibling) {
                continue
            }

            // Then finally convert the locations in the sibling document
            results = results.concat(mapRangesToInternalLocations(this.dumpId, sibling.ranges, documentPath, rangeIds))
        }

        return results
    }

    /**
     * Convert a list of ranges with document ids into a list of ranges with
     * document paths by looking into the result chunks table and parsing the
     * data associated with the given identifier.
     *
     * @param id The identifier of the definition or reference result.
     */
    private async getResultById(
        id: sqliteModels.DefinitionReferenceResultId
    ): Promise<sqliteModels.DocumentPathRangeId[]> {
        const { documentPaths, documentIdRangeIds } = await this.getResultChunkByResultId(id)
        const ranges = mustGet(documentIdRangeIds, id, 'documentIdRangeId')

        return ranges.map(range => ({
            documentPath: mustGet(documentPaths, range.documentId, 'documentPath'),
            rangeId: range.rangeId,
        }))
    }

    /**
     * Return a parsed result chunk that contains the given identifier.
     *
     * @param id An identifier contained in the result chunk.
     * @param ctx The tracing context.
     */
    private async getResultChunkByResultId(
        id: sqliteModels.DefinitionReferenceResultId,
        ctx: TracingContext = {}
    ): Promise<sqliteModels.ResultChunkData> {
        // Find the result chunk index this id belongs to
        const index = hashKey(id, await this.getNumResultChunks())

        const factory = async (): Promise<cache.EncodedJsonCacheValue<sqliteModels.ResultChunkData>> => {
            const resultChunk = await this.withConnection(
                connection => connection.getRepository(sqliteModels.ResultChunkModel).findOneOrFail(index),
                ctx.logger
            )

            return {
                size: resultChunk.data.length,
                data: await gunzipJSON<sqliteModels.ResultChunkData>(resultChunk.data),
            }
        }

        return Database.resultChunkCache.withValue(`${this.databasePath}::${index}`, factory, resultChunk =>
            Promise.resolve(resultChunk.data)
        )
    }

    /**
     * Get the `numResultChunks` value from this database's metadata row.
     *
     * @param ctx The tracing context.
     */
    private async getNumResultChunks(ctx: TracingContext = {}): Promise<number> {
        const numResultChunks = Database.numResultChunks.get(this.databasePath)
        if (numResultChunks !== undefined) {
            return numResultChunks
        }

        // Not in the shared map, need to query it
        const meta = await this.withConnection(
            connection => connection.getRepository(sqliteModels.MetaModel).findOneOrFail(1),
            ctx.logger
        )
        Database.numResultChunks.set(this.databasePath, meta.numResultChunks)
        return meta.numResultChunks
    }

    /**
     * Invoke `callback` with a SQLite connection object obtained from the
     * cache or created on cache miss.
     *
     * @param callback The function invoke with the SQLite connection.
     * @param logger The logger instance.
     */
    private withConnection<T>(
        callback: (connection: Connection) => Promise<T>,
        logger: Logger = createSilentLogger()
    ): Promise<T> {
        return Database.connectionCache.withConnection(this.databasePath, sqliteModels.entities, logger, connection =>
            instrument(metrics.databaseQueryDurationHistogram, metrics.databaseQueryErrorsCounter, () =>
                callback(connection)
            )
        )
    }

    /**
     * Log and trace the execution of a function.
     *
     * @param ctx The tracing context.
     * @param name The name of the span and text of the log message.
     * @param f  The function to invoke.
     */
    private logAndTraceCall<T>(ctx: TracingContext, name: string, f: (ctx: TracingContext) => Promise<T>): Promise<T> {
        return logAndTraceCall(addTags(ctx, { dbID: this.dumpId }), name, f)
    }

    /**
     * Logs an event to the span of the tracing context, if its defined.
     *
     * @param ctx The tracing context.
     * @param event The name of the event.
     * @param pairs The values to log.
     */
    private logSpan(ctx: TracingContext, event: string, pairs: { [name: string]: unknown }): void {
        logSpan(ctx, event, { ...pairs, dbID: this.dumpId })
    }
}

/**
 * Return the set of ranges that contain the given position. If multiple ranges
 * are returned, then the inner-most ranges will occur before the outer-most
 * ranges.
 *
 * @param ranges The set of possible ranges.
 * @param position The user's hover position.
 */
export function findRanges(ranges: Iterable<sqliteModels.RangeData>, position: lsp.Position): sqliteModels.RangeData[] {
    const filtered = []
    for (const range of ranges) {
        if (comparePosition(range, position) === 0) {
            filtered.push(range)
        }
    }

    return filtered.sort((a, b) => {
        if (comparePosition(a, { line: b.startLine, character: b.startCharacter }) === 0) {
            return +1
        }

        return -1
    })
}

/**
 * Compare a position against a range. Returns 0 if the position occurs
 * within the range (inclusive bounds), -1 if the position occurs after
 * it, and +1 if the position occurs before it.
 *
 * @param range The range.
 * @param position The position.
 */
export function comparePosition(range: sqliteModels.RangeData, position: lsp.Position): number {
    if (position.line < range.startLine) {
        return +1
    }

    if (position.line > range.endLine) {
        return -1
    }

    if (position.line === range.startLine && position.character < range.startCharacter) {
        return +1
    }

    if (position.line === range.endLine && position.character > range.endCharacter) {
        return -1
    }

    return 0
}

/**
 * Construct an LSP range from a flat range.
 *
 * @param result The start/end line/character of the range.
 */
function createRange(result: {
    startLine: number
    startCharacter: number
    endLine: number
    endCharacter: number
}): lsp.Range {
    return lsp.Range.create(result.startLine, result.startCharacter, result.endLine, result.endCharacter)
}

/**
 * Convert the given range identifiers into an `InternalLocation` objects.
 *
 * @param dumpId The identifier of the dump to which the ranges belong.
 * @param ranges The map of ranges of the document.
 * @param uri The location URI.
 * @param ids The set of range identifiers for each resulting location.
 */
export function mapRangesToInternalLocations(
    dumpId: pgModels.DumpId,
    ranges: Map<sqliteModels.RangeId, sqliteModels.RangeData>,
    uri: string,
    ids: Set<sqliteModels.RangeId>
): InternalLocation[] {
    const locations = []
    for (const id of ids) {
        locations.push({
            dumpId,
            path: uri,
            range: createRange(mustGet(ranges, id, 'range')),
        })
    }

    return locations
}

/**
 * Format ranges to be serialized in opentracing logs.
 *
 * @param ranges The ranges to clean.
 */
function cleanRanges(
    ranges: sqliteModels.RangeData[]
): (Omit<sqliteModels.RangeData, 'monikerIds'> & { monikerIds: sqliteModels.MonikerId[] })[] {
    // We need to array-ize sets otherwise we get a "0 key" object
    return ranges.map(r => ({ ...r, monikerIds: Array.from(r.monikerIds) }))
}
