import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from '../settings'
import { addTags, TracingContext } from '../../shared/tracing'
import { Connection } from 'typeorm'
import { convertLsif } from '../importer/importer'
import { createSilentLogger } from '../../shared/logging'
import { dbFilename } from '../../shared/paths'
import { Job } from 'bull'
import { withLock } from '../../shared/locker/locker'
import { XrepoDatabase } from '../../shared/xrepo/xrepo'

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the cross-repo database for this dump.
 *
 * @param connection The Postgres connection.
 * @param xrepoDatabase The cross-repo database.
 * @param fetchConfiguration A function that returns the current configuration.
 */
export const createConvertJobProcessor = (
    connection: Connection,
    xrepoDatabase: XrepoDatabase,
    fetchConfiguration: () => { gitServers: string[] }
) => async (
    job: Job,
    { repository, commit, root, filename }: { repository: string; commit: string; root: string; filename: string },
    ctx: TracingContext
): Promise<void> => {
    const { logger = createSilentLogger(), span } = addTags(ctx, { repository, commit, root })
    await convertDatabase(xrepoDatabase, repository, commit, root, filename, job.timestamp, ctx)
    await updateCommitsAndDumpsVisibleFromTip(xrepoDatabase, fetchConfiguration, repository, commit, { logger, span })
    await purgeOldDumps(connection, xrepoDatabase, settings.STORAGE_ROOT, settings.DBS_DIR_MAXIMUM_SIZE_BYTES, ctx)
}

/**
 * Convert the LSIF dump input into a SQLite database, create an LSIF dump record in the
 * cross-repo database, and populate the cross-repo database with packages and reference
 * data.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param repository The repository.
 * @param commit The commit.
 * @param root The root of the dump.
 * @param filename The path to gzipped LSIF dump.
 * @param timestamp The time the job was enqueued.
 * @param ctx The tracing context.
 */
async function convertDatabase(
    xrepoDatabase: XrepoDatabase,
    repository: string,
    commit: string,
    root: string,
    filename: string,
    timestamp: number,
    { logger = createSilentLogger(), span }: TracingContext
): Promise<void> {
    const tempFile = path.join(settings.STORAGE_ROOT, constants.TEMP_DIR, path.basename(filename))

    try {
        // Create database in a temp path
        const { packages, references } = await convertLsif(filename, tempFile, { logger, span })

        // Insert dump and add packages and references to the xrepo db
        const dump = await xrepoDatabase.addPackagesAndReferences(
            repository,
            commit,
            root,
            new Date(timestamp),
            packages,
            references,
            { logger, span }
        )

        // Move the temp file where it can be found by the server
        await fs.rename(tempFile, dbFilename(settings.STORAGE_ROOT, dump.id, repository, commit))

        logger.info('Created dump', {
            repository: dump.repository,
            commit: dump.commit,
            root: dump.root,
        })
    } catch (error) {
        // Don't leave busted artifacts
        await fs.unlink(tempFile)
        throw error
    }

    await fs.unlink(filename)
}

/**
 * Update the commits for this repo, and update the visible_at_tip flag on the dumps
 * of this repository. This will query for commits starting from both the current tip
 * of the repo and from the commit that was just processed.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param fetchConfiguration A function that returns the current configuration.
 * @param repository The repository.
 * @param commit The commit.
 * @param ctx The tracing context.
 */
async function updateCommitsAndDumpsVisibleFromTip(
    xrepoDatabase: XrepoDatabase,
    fetchConfiguration: () => { gitServers: string[] },
    repository: string,
    commit: string,
    ctx: TracingContext
): Promise<void> {
    const gitserverUrls = fetchConfiguration().gitServers

    const tipCommit = await xrepoDatabase.discoverTip({ repository, gitserverUrls, ctx })
    if (tipCommit === undefined) {
        throw new Error('No tip commit available for repository')
    }

    const commits = await xrepoDatabase.discoverCommits({
        repository,
        commit,
        gitserverUrls,
        ctx,
    })

    if (tipCommit !== commit) {
        // If the tip is ahead of this commit, we also want to discover all of
        // the commits between this commit and the tip so that we can accurately
        // determine what is visible from the tip. If we do not do this before the
        // updateDumpsVisibleFromTip call below, no dumps will be reachable from
        // the tip and all dumps will be invisible.

        const tipCommits = await xrepoDatabase.discoverCommits({
            repository,
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

    await xrepoDatabase.updateCommits(repository, commits, ctx)
    await xrepoDatabase.updateDumpsVisibleFromTip(repository, tipCommit, ctx)
}

/**
 * Remove dumps until the space occupied by the dbs directory is below
 * the given limit.
 *
 * @param connection The Postgres connection.
 * @param xrepoDatabase The cross-repo database.
 * @param storageRoot The path where SQLite databases are stored.
 * @param maximumSizeBytes The maximum number of bytes.
 * @param ctx The tracing context.
 */
function purgeOldDumps(
    connection: Connection,
    xrepoDatabase: XrepoDatabase,
    storageRoot: string,
    maximumSizeBytes: number,
    { logger = createSilentLogger() }: TracingContext = {}
): Promise<void> {
    if (maximumSizeBytes < 0) {
        return Promise.resolve()
    }

    const purge = async (): Promise<void> => {
        let currentSizeBytes = await dirsize(path.join(storageRoot, constants.DBS_DIR))

        while (currentSizeBytes > maximumSizeBytes) {
            // While our current data usage is too big, find candidate dumps to delete
            const dump = await xrepoDatabase.getOldestPrunableDump()
            if (!dump) {
                logger.warn(
                    'Unable to reduce disk usage of the DB directory because deleting any single dump would drop in-use code intel for a repository.',
                    { currentSizeBytes, softMaximumSizeBytes: maximumSizeBytes }
                )

                break
            }

            logger.info('Pruning dump', {
                repository: dump.repository,
                commit: dump.commit,
                root: dump.root,
            })

            // Delete this dump and subtract its size from the current dir size
            const filename = dbFilename(storageRoot, dump.id, dump.repository, dump.commit)
            currentSizeBytes -= await filesize(filename)

            // This delete cascades to the packages and references tables as well
            await xrepoDatabase.deleteDump(dump)
        }
    }

    // Ensure only one worker is doing this at the same time so that we don't
    // choose more dumps than necessary to purge. This can happen if the directory
    // size check and the selection of a purgeable dump are interleaved between
    // multiple workers.
    return withLock(connection, 'retention', purge)
}

/**
 * Calculate the size of a directory.
 *
 * @param directory The directory path.
 */
async function dirsize(directory: string): Promise<number> {
    return (
        await Promise.all((await fs.readdir(directory)).map(filename => filesize(path.join(directory, filename))))
    ).reduce((a, b) => a + b, 0)
}

/**
 * Get the file size or zero if it doesn't exist.
 *
 * @param filename The filename.
 */
async function filesize(filename: string): Promise<number> {
    try {
        return (await fs.stat(filename)).size
    } catch (error) {
        if (!(error && error.code === 'ENOENT')) {
            throw error
        }

        return 0
    }
}
