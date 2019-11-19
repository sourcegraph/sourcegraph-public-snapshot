import * as constants from './constants'
import * as fs from 'mz/fs'
import * as path from 'path'

/**
 * Construct the path of the SQLite database file for the given dump.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param id The ID of the dump.
 */
export function dbFilename(storageRoot: string, id: number, repository: string, commit: string): string {
    return path.join(storageRoot, constants.DBS_DIR, `${id}-${encodeURIComponent(repository)}@${commit}.lsif.db`)
}

/**
 * Construct the path of the SQLite database file for the given repository and commit.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param repository The repository name.
 * @param commit The repository commit.
 */
export function dbFilenameOld(storageRoot: string, repository: string, commit: string): string {
    return path.join(storageRoot, `${encodeURIComponent(repository)}@${commit}.lsif.db`)
}

/**
 * Ensure the directory exists.
 *
 * @param directoryPath The directory path.
 */
export async function ensureDirectory(directoryPath: string): Promise<void> {
    try {
        await fs.mkdir(directoryPath)
    } catch (error) {
        if (!(error && error.code === 'EEXIST')) {
            throw error
        }
    }
}

/**
 * Delete the file if it exists. Throws errors that are not ENOENT.
 *
 * @param filePath The path to delete.
 */
export async function tryDeleteFile(filePath: string): Promise<boolean> {
    try {
        await fs.unlink(filePath)
        return true
    } catch (error) {
        if (!(error && error.code === 'ENOENT')) {
            throw error
        }

        return false
    }
}
