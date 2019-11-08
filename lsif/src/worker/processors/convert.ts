import * as constants from '../../constants'
import * as fs from 'mz/fs'
import * as path from 'path'
import { ConfigurationFetcher } from '../../config'
import { convertLsif } from '../../importer'
import { dbFilename } from '../../util'
import { DBS_DIR_MAXIMUM_SIZE_BYTES, STORAGE_ROOT } from '../settings'
import { logAndTraceCall, TracingContext } from '../../tracing'
import { purgeOldDumps } from '../../retention'
import { XrepoDatabase } from '../../xrepo'

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
    fetchConfiguration: ConfigurationFetcher
) => async (
    { repository, commit, root, filename }: { repository: string; commit: string; root: string; filename: string },
    ctx: TracingContext
): Promise<void> => {
    await logAndTraceCall(ctx, 'converting LSIF data', async (ctx: TracingContext) => {
        const input = fs.createReadStream(filename)
        const tempFile = path.join(STORAGE_ROOT, constants.TEMP_DIR, path.basename(filename))

        try {
            // Create database in a temp path
            const { packages, references } = await convertLsif(input, tempFile, ctx)

            // Add packages and references to the xrepo db
            const dump = await logAndTraceCall(ctx, 'populating cross-repo database', () =>
                xrepoDatabase.addPackagesAndReferences(repository, commit, root, packages, references)
            )

            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, dbFilename(STORAGE_ROOT, dump.id, repository, commit))
        } catch (e) {
            // Don't leave busted artifacts
            await fs.unlink(tempFile)
            throw e
        }
    })

    // Update commit parentage information for this commit
    await xrepoDatabase.discoverAndUpdateCommit({
        repository,
        commit,
        gitserverUrls: fetchConfiguration().gitServers,
        ctx,
    })

    // Remove input
    await fs.unlink(filename)

    // Clean up disk space if necessary
    await purgeOldDumps(STORAGE_ROOT, xrepoDatabase, DBS_DIR_MAXIMUM_SIZE_BYTES, ctx)
}
