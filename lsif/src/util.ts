import * as path from 'path'
import { STORAGE_ROOT } from './settings'

/**
 * Reads an integer from an environment variable or defaults to the given value.
 *
 * @param key The environment variable name.
 * @param defaultValue The default value.
 */
export function readEnvInt(key: string, defaultValue: number): number {
    return (process.env[key] && parseInt(process.env[key] || '', 10)) || defaultValue
}

/* eslint-disable @typescript-eslint/no-explicit-any */

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
 *.Computes the filename of the LSIF dump from the given repository and commit hash.
 */
export function makeFilename(repository: string, commit: string): string {
    return path.join(STORAGE_ROOT, `${encodeURIComponent(repository)}@${commit}.lsif.db`)
}
