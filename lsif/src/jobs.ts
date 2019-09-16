import * as fs from 'mz/fs'
import * as path from 'path'
import uuid from 'uuid'
import { convertLsif } from './importer'
import { createDatabaseFilename } from './util'
import { logger } from './logger'
import { XrepoDatabase } from './xrepo'

/**
 * The names of jobs performed by the LSIF worker.
 */
export type JobClasses = 'convert'

/**
 * Create a job that takes a repository, commit, and filename containing the gzipped
 * input of an LSIF dump and converts it to a SQLite database. This will also populate
 * the cross-repo database for this dump.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param xrepoDatabase The cross-repo database.
 */
export function createConvertJob(
    storageRoot: string,
    xrepoDatabase: XrepoDatabase
): (repository: string, commit: string, filename: string) => Promise<void> {
    return async (repository, commit, filename) => {
        const jobLogger = logger.child({ repository, commit })
        jobLogger.info('Converting')

        const input = fs.createReadStream(filename)
        const tempFile = path.join(storageRoot, 'tmp', uuid.v4())

        try {
            // Create database in a temp path
            const { packages, references } = await convertLsif(input, tempFile)

            // Move the temp file where it can be found by the server
            await fs.rename(tempFile, createDatabaseFilename(storageRoot, repository, commit))

            // Add the new database to the xrepo db
            await xrepoDatabase.addPackagesAndReferences(repository, commit, packages, references)
        } catch (e) {
            jobLogger.error('Failed conversion', { repository, commit, error: e && e.message })

            // Don't leave busted artifacts
            await fs.unlink(tempFile)
            throw e
        }

        jobLogger.info('Successfully converted')

        // Remove input
        await fs.unlink(filename)
    }
}
