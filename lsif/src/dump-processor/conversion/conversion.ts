import * as pgModels from '../../shared/models/pg'
import { TracingContext } from '../../shared/tracing'
import { EntityManager } from 'typeorm'
import { convertLsif } from './importer'
import { createSilentLogger } from '../../shared/logging'
import { DependencyManager } from '../../shared/store/dependencies'
import { PathExistenceChecker } from './existence'

/**
 * Convert the LSIF dump input into a SQLite database and populate the dependency tables
 * with packages and reference data.
 *
 * @param entityManager The EntityManager to use as part of a transaction.
 * @param dependencyManager The dependency manager instance.
 * @param frontendUrl The url of the frontend internal API.
 * @param upload The unprocessed upload record.
 * @param uploadPath The path of the upload file to read.
 * @param databasePath The path of the database file to populate.
 * @param ctx The tracing context.
 */
export async function convertDatabase(
    entityManager: EntityManager,
    dependencyManager: DependencyManager,
    frontendUrl: string,
    upload: pgModels.LsifUpload,
    uploadPath: string,
    databasePath: string,
    { logger = createSilentLogger(), span }: TracingContext
): Promise<void> {
    const ctx = { logger, span }

    const pathExistenceChecker = new PathExistenceChecker({
        repositoryId: upload.repositoryId,
        commit: upload.commit,
        root: upload.root,
        frontendUrl,
        ctx,
    })

    // Create database in a temp path
    const { packages, references } = await convertLsif({
        path: uploadPath,
        root: upload.root,
        database: databasePath,
        pathExistenceChecker,
        ctx,
    })

    // Insert dump and add packages and references to Postgres
    await dependencyManager.addPackagesAndReferences(upload.id, packages, references, ctx, entityManager)

    logger.info('Converted upload', {
        repositoryId: upload.repositoryId,
        commit: upload.commit,
        root: upload.root,
    })
}
