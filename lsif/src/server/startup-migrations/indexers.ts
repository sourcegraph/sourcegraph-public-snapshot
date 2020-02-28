import { Connection } from 'typeorm'
import * as pgModels from '../../shared/models/pg'
import { logAndTraceCall, TracingContext } from '../../shared/tracing'
import { ConnectionCache, DocumentCache, ResultChunkCache } from '../backend/cache'
import * as settings from '../settings'
import { Database } from '../backend/database'
import { extname } from 'path'
import { dbFilename } from '../../shared/paths'
import { createSilentLogger } from '../../shared/logging'
import { chunk } from 'lodash'

/**
 * How many databases to search in parallel. This value limits the number of open
 * SQLite handles, which could probably be much higher. There are only a few hundred
 * (at most) active dumps per instance at this point, so we don't need to go crazy
 * here either.
 */
const CONCURRENCY_LEVEL = 20

/**
 * Assign indexers to uploads that do not have it set.
 *
 * @param connection The Postgres connection.
 * @param ctx The tracing context.
 */
export function assignIndexer(connection: Connection, ctx: TracingContext): Promise<void> {
    return logAndTraceCall(ctx, 'Assigning indexers to dumps', async ({ logger = createSilentLogger() }) => {
        const entityManager = connection.createEntityManager()
        const connectionCache = new ConnectionCache(settings.CONNECTION_CACHE_CAPACITY)
        const documentCache = new DocumentCache(settings.DOCUMENT_CACHE_CAPACITY)
        const resultChunkCache = new ResultChunkCache(settings.RESULT_CHUNK_CACHE_CAPACITY)

        const createDatabase = (dump: pgModels.LsifDump): Database =>
            new Database(
                connectionCache,
                documentCache,
                resultChunkCache,
                dump,
                dbFilename(settings.STORAGE_ROOT, dump.id)
            )

        const updateDump = async (dump: pgModels.LsifDump): Promise<void> => {
            const indexer = determineIndexer(await createDatabase(dump).documentPaths())
            if (!indexer) {
                logger.warn(`Unable to determine indexer used to create dump ${dump.id}`)
                return
            }

            await entityManager.query('UPDATE lsif_uploads SET indexer = $1 WHERE id= $2', [indexer, dump.id])
        }

        const dumps = await entityManager
            .getRepository(pgModels.LsifDump)
            .createQueryBuilder()
            .select()
            .where({ indexer: '' })
            .getMany()

        for (const batch of chunk(dumps, CONCURRENCY_LEVEL)) {
            await Promise.all(batch.map(updateDump))
        }

        await entityManager
            .getRepository(pgModels.LsifDump)
            .query("UPDATE lsif_uploads SET indexer='lsif-tsc' WHERE indexer='lsif-node'")

        await connectionCache.flush()
        await documentCache.flush()
        await resultChunkCache.flush()
    })
}

const extensionsToIndexer = new Map([
    ['.c', 'lsif-cpp'],
    ['.cpp', 'lsif-cpp'],
    ['.dart', 'lsif-dart'],
    ['.go', 'lsif-go'],
    ['.h', 'lsif-cpp'],
    ['.h', 'lsif-cpp'],
    ['.java', 'lsif-java'],
    ['.scala', 'lsif-semanticdb'],
    ['.ts', 'lsif-tsc'],
])

function determineIndexer(paths: string[]): string | undefined {
    for (const ext of new Set(paths.map(extname))) {
        if (extensionsToIndexer.has(ext)) {
            return extensionsToIndexer.get(ext)
        }
    }

    return undefined
}
