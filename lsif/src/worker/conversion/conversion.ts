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

/**
 * Convert the LSIF dump input into a SQLite database and populate the dependency tables
 * with packages and reference data.
 *
 * @param entityManager The EntityManager to use as part of a transaction.
 * @param dumpManager The dumps manager instance.
 * @param dependencyManager The dependency manager instance.
 * @param upload THe unprocessed upload record.
 * @param ctx The tracing context.
 */
export async function convertDatabase(
    entityManager: EntityManager,
    dumpManager: DumpManager,
    dependencyManager: DependencyManager,
    upload: pgModels.LsifUpload,
    { logger = createSilentLogger(), span }: TracingContext
): Promise<void> {
    const tempFile = path.join(settings.STORAGE_ROOT, constants.TEMP_DIR, path.basename(upload.filename))

    try {
        // Create database in a temp path
        const { packages, references } = await convertLsif(upload.filename, tempFile, { logger, span })

        // Insert dump and add packages and references to Postgres
        await dependencyManager.addPackagesAndReferences(
            upload.id,
            packages,
            references,
            { logger, span },
            entityManager
        )

        // Move the temp file where it can be found by the server
        await fs.rename(tempFile, dbFilename(settings.STORAGE_ROOT, upload.id, upload.repository, upload.commit))

        logger.info('Converted upload', {
            repository: upload.repository,
            commit: upload.commit,
            root: upload.root,
        })
    } catch (error) {
        // Don't leave busted artifacts
        await fs.unlink(tempFile)
        throw error
    }

    await fs.unlink(upload.filename)
}

/**
 * Update the commits for this repo, and update the visible_at_tip flag on the dumps
 * of this repository. This will query for commits starting from both the current tip
 * of the repo and from the commit that was just processed.
 *
 * @param entityManager The EntityManager to use as part of a transaction.
 * @param dumpManager The dumps manager instance.
 * @param fetchConfiguration A function that returns the current configuration.
 * @param upload The processed upload record.
 * @param ctx The tracing context.
 */
export async function updateCommitsAndDumpsVisibleFromTip(
    entityManager: EntityManager,
    dumpManager: DumpManager,
    fetchConfiguration: () => { gitServers: string[] },
    upload: pgModels.LsifUpload,
    ctx: TracingContext
): Promise<void> {
    const gitserverUrls = fetchConfiguration().gitServers

    const tipCommit = await dumpManager.discoverTip({ repository: upload.repository, gitserverUrls, ctx })
    if (tipCommit === undefined) {
        throw new Error('No tip commit available for repository')
    }

    const commits = await dumpManager.discoverCommits({
        repository: upload.repository,
        commit: upload.commit,
        gitserverUrls,
        ctx,
    })

    if (tipCommit !== upload.commit) {
        // If the tip is ahead of this commit, we also want to discover all of
        // the commits between this commit and the tip so that we can accurately
        // determine what is visible from the tip. If we do not do this before the
        // updateDumpsVisibleFromTip call below, no dumps will be reachable from
        // the tip and all dumps will be invisible.

        const tipCommits = await dumpManager.discoverCommits({
            repository: upload.repository,
            commit: tipCommit,
            gitserverUrls,
            ctx,
        })

        for (const [k, v] of tipCommits.entries()) {
            commits.set(
                k,
                new Set<string>([...(commits.get(k) || []), ...v])
            )
        }
    }

    await dumpManager.updateCommits(upload.repository, commits, ctx, entityManager)
    await dumpManager.updateDumpsVisibleFromTip(upload.repository, tipCommit, ctx, entityManager)
}
