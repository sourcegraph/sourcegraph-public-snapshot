import * as path from 'path'

export function stripPathPrefix(prefix: string, filePath: string): string | null {
    if (prefix === filePath) {
        return path.sep
    }
    const prefixWithSep = prefix.endsWith(path.sep) ? prefix : prefix + path.sep
    if (filePath.indexOf(prefixWithSep) === 0) {
        return filePath.substring(prefixWithSep.length)
    }
    return null
}

// function that that takes in an ancestor directory path and a descendant directory path and returns every directory path between them
export function getDirectoryPathsBetween(ancestor: string, descendant: string): string[] {
    const ancestorWithSep = ancestor.endsWith(path.sep) ? ancestor : ancestor + path.sep
    if (ancestor !== descendant && descendant.indexOf(ancestorWithSep) !== 0) {
        return []
    }
    var paths = []
    let current = descendant
    while (current !== ancestor && current !== '/' && current.length > 0) {
        paths.push(current)
        current = path.dirname(current)
    }
    paths.push(current)
    return paths
}
