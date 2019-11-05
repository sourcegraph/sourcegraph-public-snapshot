import * as fs from 'mz/fs'
import * as path from 'path'
import { Id } from 'lsif-protocol'
import * as constants from './constants'
import { DefinitionReferenceResultId } from './database.types'

/**
 * Reads an integer from an environment variable or defaults to the given value.
 *
 * @param key The environment variable name.
 * @param defaultValue The default value.
 */
export function readEnvInt(key: string, defaultValue: number): number {
    return (process.env[key] && parseInt(process.env[key] || '', 10)) || defaultValue
}

/**
 * Determine if an exception value has the given error code.
 *
 * @param e The exception value.
 * @param expectedCode The expected error code.
 */
export function hasErrorCode(e: any, expectedCode: string): boolean {
    return e && e.code === expectedCode
}

/**
 * Return the value of the given key from the given map. If the key does not
 * exist in the map, an exception is thrown with the given error text.
 *
 * @param map The map to query.
 * @param key The key to search for.
 * @param elementType The type of element (used for exception message).
 */
export function mustGet<K, V>(map: Map<K, V>, key: K, elementType: string): V {
    const value = map.get(key)
    if (value !== undefined) {
        return value
    }

    throw new Error(`Unknown ${elementType} '${key}'.`)
}

/**
 * Return the value of the given key from one of the given maps. The first
 * non-undefined value to be found is returned. If the key does not exist in
 * either map, an exception is thrown with the given error text.
 *
 * @param map1 The first map to query.
 * @param map2 The second map to query.
 * @param key The key to search for.
 * @param elementType The type of element (used for exception message).
 */
export function mustGetFromEither<K1, V1, K2, V2>(
    map1: Map<K1, V1>,
    map2: Map<K2, V2>,
    key: K1 & K2,
    elementType: string
): V1 | V2 {
    for (const map of [map1, map2]) {
        const value = map.get(key)
        if (value !== undefined) {
            return value
        }
    }

    throw new Error(`Unknown ${elementType} '${key}'.`)
}

/**
 * Return the value of `id`, or throw an exception if it is undefined.
 *
 * @param id The identifier.
 */
export function assertId<T extends Id>(id: T | undefined): T {
    if (id !== undefined) {
        return id
    }

    throw new Error('id is undefined')
}

/**
 * Hash a string or numeric identifier into the range `[0, maxIndex)`. The
 * hash algorithm here is similar to the one used in Java's String.hashCode.
 *
 * @param id The identifier to hash.
 * @param maxIndex The maximum of the range.
 */
export function hashKey(id: DefinitionReferenceResultId, maxIndex: number): number {
    const s = `${id}`

    let hash = 0
    for (let i = 0; i < s.length; i++) {
        const chr = s.charCodeAt(i)
        hash = (hash << 5) - hash + chr
        hash |= 0
    }

    // Hash value may be negative - must unset sign bit before modulus
    return Math.abs(hash) % maxIndex
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
 * Construct the path of the SQLite database file for the given dump.
 *
 * @param storageRoot The path where SQLite databases are stored.
 * @param id The ID of the dump.
 */
export function dbFilename(storageRoot: string, id: number, repository: string, commit: string): string {
    return path.join(storageRoot, constants.DBS_DIR, `${id}-${encodeURIComponent(repository)}@${commit}.lsif.db`)
}

/**
 * Ensure the directory exists.
 *
 * @param directoryPath The directory path.
 */
export async function ensureDirectory(directoryPath: string): Promise<void> {
    try {
        await fs.mkdir(directoryPath)
    } catch (e) {
        if (!hasErrorCode(e, 'EEXIST')) {
            throw e
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
        if (!hasErrorCode(error, 'ENOENT')) {
            throw error
        }

        return false
    }
}

/**
 * Determine the table inserter batch size for an entity given the number of
 * fields inserted for that entity. We cannot perform an insert operation with
 * more than 999 placeholder variables, so we need to flush our batch before
 * we reach that amount. If fields are added to the models, the argument to
 * this function also needs to change.
 *
 * @param numFields The number of fields for an entity.
 */
export function getBatchSize(numFields: number): number {
    return Math.floor(999 / numFields)
}

/**
 * No matching LSIF dump was found. This could be because:
 *
 * - You're currently browsing while on a commit that is too far away from the
 *   last uploaded LSIF dump
 * - You're currently viewing a file that is under a different root from what
 *   the LSIF dump is associated with (i.e. the current file is not contained in
 *   the dump)
 * - You're currently viewing a file that is not part of the LSIF dump (e.g. due
 *   to tsconfig.json exclude rules)
 */
export class NoLSIFDumpError extends Error {}
