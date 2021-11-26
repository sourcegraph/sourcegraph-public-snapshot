/**
 * Returns all but the last element of path, or "." if that would be the empty path.
 */
export function dirname(path: string): string {
    return path.split('/').slice(0, -1).join('/') || '.'
}

/**
 * Returns the last element of path, or "." if path is empty.
 */
export function basename(path: string): string {
    return path.split('/').slice(-1)[0] || '.'
}

/**
 * Reports whether path has the given prefix.
 */
export function pathHasPrefix(path: string, prefix: string): boolean {
    if (prefix === '' || prefix === '.') {
        return true
    }
    return path === prefix || path.startsWith(prefix + '/')
}

/**
 * Returns the path to `to` relative to `from`.
 */
export function pathRelative(from: string, to: string): string {
    if (from === '' || from === '.') {
        return to
    }
    return to.slice(from.length + 1)
}
