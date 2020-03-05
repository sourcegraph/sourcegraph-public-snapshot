import * as uuid from 'uuid'
import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from '../settings'
import * as pgModels from '../../shared/models/pg'
import { TracingContext } from '../../shared/tracing'
import { EntityManager } from 'typeorm'
import { convertLsif } from './importer'
import { createSilentLogger } from '../../shared/logging'
import { dbFilename } from '../../shared/paths'
import { DumpManager } from '../../shared/store/dumps'
import { DependencyManager } from '../../shared/store/dependencies'
import { PathExistenceChecker } from './existence'

/**
 * Convert the LSIF dump input into a SQLite database and populate the dependency tables
 * with packages and reference data.
 *
 * @param entityManager The EntityManager to use as part of a transaction.
 * @param dumpManager The dumps manager instance.
 * @param dependencyManager The dependency manager instance.
 * @param frontendUrl The url of the frontend internal API.
 * @param upload The unprocessed upload record.
 * @param ctx The tracing context.
 */
export async function convertDatabase(
    entityManager: EntityManager,
    dumpManager: DumpManager,
    dependencyManager: DependencyManager,
    frontendUrl: string,
    upload: pgModels.LsifUpload,
    { logger = createSilentLogger(), span }: TracingContext
): Promise<void> {
    const ctx = { logger, span }
    const tempFile = path.join(settings.STORAGE_ROOT, constants.TEMP_DIR, uuid.v4())

    try {
        const pathExistenceChecker = new PathExistenceChecker({
            repositoryId: upload.repositoryId,
            commit: upload.commit,
            root: upload.root,
            frontendUrl,
            ctx,
        })

        // Create database in a temp path
        const { packages, references } = await convertLsif({
            path: upload.filename,
            root: upload.root,
            database: tempFile,
            pathExistenceChecker,
            ctx,
        })

        // Insert dump and add packages and references to Postgres
        await dependencyManager.addPackagesAndReferences(upload.id, packages, references, ctx, entityManager)

        // Move the temp file where it can be found by the server
        await fs.rename(tempFile, dbFilename(settings.STORAGE_ROOT, upload.id))

        logger.info('Converted upload', {
            repositoryId: upload.repositoryId,
            commit: upload.commit,
            root: upload.root,
        })
    } catch (error) {
        // Don't leave busted artifacts
        await fs.unlink(tempFile)
        throw error
    }

    // Remove source upload
    await fs.unlink(upload.filename)
}
