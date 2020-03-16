import * as uuid from 'uuid'
import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from '../settings'
import * as pgModels from '../../shared/models/pg'
import { TracingContext, logAndTraceCall } from '../../shared/tracing'
import { EntityManager } from 'typeorm'
import { convertLsif } from './importer'
import { createSilentLogger } from '../../shared/logging'
import { DependencyManager } from '../../shared/store/dependencies'
import { PathExistenceChecker } from './existence'
import got from 'got'
import { pipeline as _pipeline } from 'stream'
import { promisify } from 'util'

const pipeline = promisify(_pipeline)

/**
 * Convert the LSIF dump input into a SQLite database and populate the dependency tables
 * with packages and reference data.
 *
 * @param entityManager The EntityManager to use as part of a transaction.
 * @param dependencyManager The dependency manager instance.
 * @param frontendUrl The url of the frontend internal API.
 * @param upload The unprocessed upload record.
 * @param ctx The tracing context.
 */
export async function convertDatabase(
    entityManager: EntityManager,
    dependencyManager: DependencyManager,
    frontendUrl: string,
    upload: pgModels.LsifUpload,
    { logger = createSilentLogger(), span }: TracingContext
): Promise<void> {
    const ctx = { logger, span }
    const uploadFile = path.join(settings.STORAGE_ROOT, constants.TEMP_DIR, uuid.v4())
    const databaseFile = path.join(settings.STORAGE_ROOT, constants.TEMP_DIR, uuid.v4())
    const uploadUrl = new URL(`http://localhost:3188/${upload.payloadId}/raw`).href // TODO
    const databaseUrl = new URL(`http://localhost:3188/${upload.id}`).href // TODO

    try {
        const pathExistenceChecker = new PathExistenceChecker({
            repositoryId: upload.repositoryId,
            commit: upload.commit,
            root: upload.root,
            frontendUrl,
            ctx,
        })

        // Retrieve dump to process locally
        await logAndTraceCall(ctx, 'Retrieving upload file', () =>
            pipeline(got.stream(uploadUrl), fs.createWriteStream(uploadFile))
        )

        // Create database in a temp path
        const { packages, references } = await convertLsif({
            path: uploadFile,
            root: upload.root,
            database: databaseFile,
            pathExistenceChecker,
            ctx,
        })

        // Insert dump and add packages and references to Postgres
        await dependencyManager.addPackagesAndReferences(upload.id, packages, references, ctx, entityManager)

        // Move the temp file where it can be found by the server
        await logAndTraceCall(ctx, 'Uploading database to storage server', () =>
            pipeline(fs.createReadStream(databaseFile), got.stream.post(databaseUrl))
        )

        logger.info('Converted upload', {
            repositoryId: upload.repositoryId,
            commit: upload.commit,
            root: upload.root,
        })
    } catch (error) {
        // Don't leave busted artifacts
        await fs.unlink(uploadFile)
        await fs.unlink(databaseFile)
        throw error
    }

    // Cleanup
    await fs.unlink(uploadFile)
    await fs.unlink(databaseFile)
    await got.delete(uploadUrl)
}
