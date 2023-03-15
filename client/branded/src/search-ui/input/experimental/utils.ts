/**
 * Takes in a file path and a desired length. Will shorten the path so that it's
 * as close to the length as possible (but doesn't gurantee it). The first two
 * and last two path segements are always visible.
 */
export function shortenPath(path: string, desiredChars: number): string {
    if (path.length <= desiredChars) {
        return path
    }

    const parts = path.split('/')
    if (parts.length < 5) {
        // We need at least 5 segements to shorten the path
        return path
    }
    const shortendPath = parts
        .slice(0, 2)
        .concat('...', ...parts.slice(-2))
        .join('/')
    return shortendPath.length < path.length ? shortendPath : path
}
