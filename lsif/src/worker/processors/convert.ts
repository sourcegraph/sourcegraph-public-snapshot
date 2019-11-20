import * as constants from '../../shared/constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import * as settings from '../settings'
import { convertLsif } from '../importer/importer'
import { createSilentLogger } from '../../shared/logging'
import { dbFilename } from '../../shared/paths'
import { TracingContext, addTags } from '../../shared/tracing'
import { XrepoDatabase } from '../../shared/xrepo/xrepo'
import { Job } from 'bull'

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the cross-repo database for this dump.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param fetchConfiguration A function that returns the current configuration.
 */
export const createConvertJobProcessor = (
    xrepoDatabase: XrepoDatabase,
    fetchConfiguration: () => { gitServers: string[] }
) => async (
    job: Job,
    { repository, commit, root, filename }: { repository: string; commit: string; root: string; filename: string },
    ctx: TracingContext
): Promise<void> => {
    const { logger = createSilentLogger(), span } = addTags(ctx, { repository, commit, root })
    const tempFile = path.join(settings.STORAGE_ROOT, constants.TEMP_DIR, path.basename(filename))

    try {
        // Create database in a temp path
        const { packages, references } = await convertLsif(filename, tempFile, { logger, span })

        // Insert dump and add packages and references to the xrepo db
        const dump = await xrepoDatabase.addPackagesAndReferences(
            repository,
            commit,
            root,
            new Date(job.timestamp),
            packages,
            references,
            ctx
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

    // Remove input
    await fs.unlink(filename)

    try {
        // Update commit parentage information for this commit
        await xrepoDatabase.discoverAndUpdateCommit({
            repository,
            commit,
            gitserverUrls: fetchConfiguration().gitServers,
            ctx: { logger, span },
        })
    } catch (error) {
        // At this point, the job has already completed successfully. Catch
        // any error that happens from `discoverAndUpdateCommit` and swallow
        // it. There is no need to log here as any error that occurs within
        // the call will already be logged by `instrument` blocks.
    }

    try {
        // Clean up disk space if necessary - use original tracing context so the labels
        // repository, commit, and root do not get ambiguous between the job arguments
        // and the properties of the dump being purged.
        await purgeOldDumps(settings.STORAGE_ROOT, xrepoDatabase, settings.DBS_DIR_MAXIMUM_SIZE_BYTES, ctx)
    } catch (error) {
        // At this point, the job has already completed successfully. Catch
        // any error that happens from `purgeOldDumps` and swallow it. There
        // is no need to log here as any error that occurs within the call
        // will already be logged by `instrument` blocks.
    }
}

/**
 * Remove dumps until the space occupied by the dbs directory is below
 * the given limit.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param xrepoDatabase The cross-repo database.
 * @param maximumSizeBytes The maximum number of bytes.
 * @param ctx The tracing context.
 */
function purgeOldDumps(
    storageRoot: string,
    xrepoDatabase: XrepoDatabase,
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
    return withLock(xrepoDatabase, 'retention', purge)
}

/**
 * Hold a Postgres advisory lock while executing the given function.
 *
 * @param xrepoDatabase The cross-repo database.
 * @param name The name of the lock.
 * @param f The function to execute while holding the lock.
 */
async function withLock<T>(xrepoDatabase: XrepoDatabase, name: string, f: () => Promise<T>): Promise<T> {
    await xrepoDatabase.lock(name)
    try {
        return await f()
    } finally {
        await xrepoDatabase.unlock(name)
    }
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
