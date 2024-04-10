const BYTES_IN_KB = 1000
const LOG_BYTES_IN_KB = Math.log(BYTES_IN_KB)
const UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']

/**
 * Formates a number of bytes into a human readable string.
 *
 * Example: `formatBytes(1000)` returns `'1.00 KB'`.
 *
 * @param bytes The number of bytes to format.
 * @returns The human readable string.
 */
export function formatBytes(bytes: number): string {
    if (bytes === 0) {
        return '0 B'
    }
    const i = Math.floor(Math.log(bytes) / LOG_BYTES_IN_KB)
    const value = i === 0 ? bytes : (bytes / Math.pow(BYTES_IN_KB, i)).toFixed(2)
    return `${value} ${UNITS[i]}`
}
